package service

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

type BounceIMAPPoller struct {
	queries      *db.Queries
	suppressions *SuppressionService
	encSecret    string
	interval     time.Duration
}

func NewBounceIMAPPoller(queries *db.Queries, suppressions *SuppressionService, encSecret string) *BounceIMAPPoller {
	return &BounceIMAPPoller{
		queries:      queries,
		suppressions: suppressions,
		encSecret:    encSecret,
		interval:     5 * time.Minute,
	}
}

func (p *BounceIMAPPoller) Start(ctx context.Context) {
	go p.run(ctx)
	log.Println("bounce imap poller: started")
}

func (p *BounceIMAPPoller) run(ctx context.Context) {
	t := time.NewTicker(p.interval)
	defer t.Stop()
	p.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.tick(ctx)
		}
	}
}

func (p *BounceIMAPPoller) tick(ctx context.Context) {
	projects, err := p.queries.ListProjectsWithBounceIMAP(ctx)
	if err != nil {
		log.Printf("bounce imap poller: list projects failed: %v", err)
		return
	}
	for _, project := range projects {
		if err := p.pollProject(ctx, project); err != nil {
			log.Printf("bounce imap poller: project %s failed: %v", project.ID, err)
		}
	}
}

func (p *BounceIMAPPoller) pollProject(ctx context.Context, project db.Project) error {
	if !project.BounceImapHost.Valid || !project.BounceImapUser.Valid || !project.BounceImapPasswordEncrypted.Valid {
		return nil
	}
	host := project.BounceImapHost.String
	port := int32(993)
	if project.BounceImapPort.Valid {
		port = project.BounceImapPort.Int32
	}
	user := project.BounceImapUser.String
	pass := project.BounceImapPasswordEncrypted.String
	if dec, err := Decrypt(pass, p.encSecret); err == nil {
		pass = dec
	}
	folder := project.BounceImapFolder
	if folder == "" {
		folder = "INBOX"
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := imapclient.DialTLS(addr, &imapclient.Options{TLSConfig: &tls.Config{ServerName: host}})
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	if err := client.Login(user, pass).Wait(); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	defer client.Logout()

	if _, err := client.Select(folder, nil).Wait(); err != nil {
		return fmt.Errorf("select %q: %w", folder, err)
	}

	criteria := &imap.SearchCriteria{NotFlag: []imap.Flag{imap.FlagSeen}}
	searchData, err := client.Search(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	uids := searchData.AllSeqNums()
	if len(uids) == 0 {
		return nil
	}

	seqSet := imap.SeqSet{}
	for _, n := range uids {
		seqSet.AddNum(n)
	}
	fetchOpts := &imap.FetchOptions{BodySection: []*imap.FetchItemBodySection{{}}}
	msgs, err := client.Fetch(seqSet, fetchOpts).Collect()
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	processed := 0
	for _, msg := range msgs {
		body := msg.FindBodySection(&imap.FetchItemBodySection{})
		if len(body) == 0 {
			continue
		}
		emails := extractBounceRecipients(string(body))
		for _, email := range emails {
			if p.suppressions != nil {
				_, _ = p.suppressions.Add(ctx, project.ID, email, SuppressionReasonBounce, "imap dsn poll")
			}
			_ = p.queries.MarkLatestLogBouncedByEmail(ctx, db.MarkLatestLogBouncedByEmailParams{
				ProjectID: project.ID,
				ToEmail:   email,
				Error:     sql.NullString{String: "bounce reported via IMAP DSN", Valid: true},
			})
		}
		if len(emails) > 0 {
			processed++
		}
	}

	if processed > 0 {
		log.Printf("bounce imap poller: project %s · %d/%d messages produced suppressions", project.ID, processed, len(uids))
	}

	storeFlags := imap.StoreFlags{Op: imap.StoreFlagsAdd, Silent: true, Flags: []imap.Flag{imap.FlagSeen}}
	if err := client.Store(seqSet, &storeFlags, nil).Close(); err != nil {
		return fmt.Errorf("store seen: %w", err)
	}
	return nil
}

var (
	dsnRecipientRE = regexp.MustCompile(`(?im)^Final-Recipient:\s*[^;]+;\s*([^\s<>]+@[^\s<>]+)`)
	angleEmailRE   = regexp.MustCompile(`<([^@<>\s]+@[^@<>\s]+)>`)
	bareEmailRE    = regexp.MustCompile(`[\w._%+-]+@[\w.-]+\.[A-Za-z]{2,}`)
	hardBounceLine = regexp.MustCompile(`(?i)\b(550|551|553|554|521)\b`)
)

func extractBounceRecipients(raw string) []string {
	seen := map[string]struct{}{}
	out := []string{}

	for _, m := range dsnRecipientRE.FindAllStringSubmatch(raw, -1) {
		add := strings.ToLower(strings.TrimSpace(m[1]))
		if _, ok := seen[add]; !ok && validBounceCandidate(add) {
			seen[add] = struct{}{}
			out = append(out, add)
		}
	}
	if len(out) > 0 {
		return out
	}

	for _, line := range strings.Split(raw, "\n") {
		if !hardBounceLine.MatchString(line) {
			continue
		}
		for _, m := range angleEmailRE.FindAllStringSubmatch(line, -1) {
			add := strings.ToLower(strings.TrimSpace(m[1]))
			if _, ok := seen[add]; !ok && validBounceCandidate(add) {
				seen[add] = struct{}{}
				out = append(out, add)
			}
		}
		if len(out) > 0 {
			continue
		}
		for _, m := range bareEmailRE.FindAllString(line, -1) {
			add := strings.ToLower(strings.TrimSpace(m))
			if _, ok := seen[add]; !ok && validBounceCandidate(add) {
				seen[add] = struct{}{}
				out = append(out, add)
			}
		}
	}
	return out
}

func validBounceCandidate(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	lower := strings.ToLower(email)
	bad := []string{"mailer-daemon@", "postmaster@", "noreply@", "no-reply@"}
	for _, b := range bad {
		if strings.HasPrefix(lower, b) {
			return false
		}
	}
	return true
}
