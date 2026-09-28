package service

import (
	"testing"
	"time"
)

// Startup recovery reclaims broadcast jobs whose lease has expired. If the lease were
// shorter than a send attempt, an instance starting up could reclaim a job another
// instance is still delivering and send the same email twice — the exact failure the
// lease exists to prevent.
func TestBroadcastJobLeaseOutlivesSendAttempt(t *testing.T) {
	attempt := smtpConnectTimeout + smtpSessionTimeout
	if broadcastJobLease <= attempt {
		t.Fatalf("broadcastJobLease (%s) must exceed a full send attempt (%s)", broadcastJobLease, attempt)
	}
}

// A claim that is immediately past its deadline would be reclaimed by the next startup,
// so the window has to run forward from the moment of the claim.
func TestLeaseDeadlineRunsForward(t *testing.T) {
	now := time.Now()
	if !leaseDeadline(now).After(now) {
		t.Fatal("leaseDeadline must return a time after the claim")
	}
}
