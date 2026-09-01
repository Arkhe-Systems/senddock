package handler

import (
	"html"
	"strings"

	"github.com/arkhe-systems/senddock/internal/richtext"
	"github.com/arkhe-systems/senddock/internal/service"
)

type unsubPageMode string

const (
	unsubPageConfirm unsubPageMode = "confirm"
	unsubPageDone    unsubPageMode = "done"
)

func brandedConfirmButton(action, label string) string {
	return `<form method="POST" action="` + action + `" style="margin:16px 0 0 0"><button type="submit" style="display:inline-block;padding:10px 20px;border:0;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;background:#111111;color:#ffffff">` + label + `</button></form>`
}

func brandedDoneBlock(email string) string {
	return `<div style="margin:16px 0 0 0;padding:12px 16px;border-radius:8px;background:#f0fdf4;color:#166534;font-size:14px">You have been unsubscribed. ` + html.EscapeString(email) + ` will no longer receive these emails.</div>`
}

func renderBrandedUnsubscribePage(rawTemplateHTML string, ctx service.UnsubscribeContext, mode unsubPageMode, projectID, subscriberID, token string) string {
	out := richtext.SanitizePage(rawTemplateHTML)

	newsletterName := ctx.NewsletterName
	if newsletterName == "" {
		newsletterName = ctx.ProjectName
	}

	action := "/unsubscribe/" + projectID + "/" + subscriberID + "?t=" + token
	if ctx.NewsletterID != "" {
		action = "/unsubscribe/" + projectID + "/" + subscriberID + "?n=" + ctx.NewsletterID + "&t=" + token
	}

	allLink := ""
	if ctx.NewsletterID != "" && ctx.AllToken != "" {
		allLink = `<a href="/unsubscribe/` + projectID + `/` + subscriberID + `?t=` + ctx.AllToken + `" style="font-size:13px;color:#6b7280;text-decoration:underline">Unsubscribe from all emails</a>`
	}

	replacer := strings.NewReplacer(
		"{{project_name}}", html.EscapeString(ctx.ProjectName),
		"{{email}}", html.EscapeString(ctx.Email),
		"{{newsletter_name}}", html.EscapeString(newsletterName),
		"{{unsubscribe_all_link}}", allLink,
		"{{manage_preferences}}", "",
	)
	out = replacer.Replace(out)

	var actionBlock string
	if mode == unsubPageDone {
		actionBlock = brandedDoneBlock(ctx.Email)
	} else {
		label := "Confirm unsubscribe"
		if ctx.NewsletterID != "" && ctx.NewsletterName != "" {
			label = "Unsubscribe from " + html.EscapeString(ctx.NewsletterName)
		}
		actionBlock = brandedConfirmButton(action, label)
	}

	if strings.Contains(out, "{{confirm_button}}") {
		out = strings.ReplaceAll(out, "{{confirm_button}}", actionBlock)
	} else {
		if idx := strings.LastIndex(strings.ToLower(out), "</body>"); idx != -1 {
			out = out[:idx] + actionBlock + out[idx:]
		} else {
			out = out + actionBlock
		}
	}

	trimmed := strings.TrimSpace(strings.ToLower(out))
	if !strings.HasPrefix(trimmed, "<!doctype") && !strings.HasPrefix(trimmed, "<html") {
		out = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Unsubscribe</title></head><body>` + out + `</body></html>`
	} else if !strings.HasPrefix(trimmed, "<!doctype") {
		out = "<!DOCTYPE html>" + out
	}

	return out
}
