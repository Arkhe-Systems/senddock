package handler

import (
	"strings"
	"testing"

	"github.com/arkhe-systems/senddock/internal/service"
)

func TestRenderBrandedPageSanitizesBeforeSubstituting(t *testing.T) {
	tpl := `<div style="background:#111"><h1>Leave {{newsletter_name}}?</h1><script>alert(1)</script>{{confirm_button}}</div>`
	ctx := service.UnsubscribeContext{ProjectName: "Acme", Email: "a@b.co", NewsletterID: "nl-1", NewsletterName: "Dev Tips", AllToken: "all-token"}

	out := renderBrandedUnsubscribePage(tpl, ctx, unsubPageConfirm, "p-1", "s-1", "tok")

	if strings.Contains(out, "<script") || strings.Contains(out, "alert(1)") {
		t.Fatalf("script survived sanitization: %q", out)
	}
	if !strings.Contains(out, `<form method="POST" action="/unsubscribe/p-1/s-1?n=nl-1&t=tok"`) {
		t.Fatalf("trusted confirm form missing: %q", out)
	}
	if !strings.Contains(out, "Leave Dev Tips?") {
		t.Fatalf("newsletter name not substituted: %q", out)
	}
	if !strings.Contains(out, "Unsubscribe from Dev Tips") {
		t.Fatalf("button label missing: %q", out)
	}
}

func TestRenderBrandedPageEscapesValues(t *testing.T) {
	tpl := `<p>{{project_name}} — {{email}}</p>{{confirm_button}}`
	ctx := service.UnsubscribeContext{ProjectName: `<img src=x onerror=alert(1)>`, Email: `"><script>x</script>`}

	out := renderBrandedUnsubscribePage(tpl, ctx, unsubPageConfirm, "p-1", "s-1", "tok")

	if strings.Contains(out, "<script>") || strings.Contains(out, "<img src=x") {
		t.Fatalf("unescaped value injected live markup: %q", out)
	}
	if !strings.Contains(out, "&lt;img") || !strings.Contains(out, "&lt;script&gt;") {
		t.Fatalf("values were not entity-escaped: %q", out)
	}
}

func TestRenderBrandedPageForcesConfirmButton(t *testing.T) {
	tpl := `<html><body><p>Bye</p></body></html>`
	ctx := service.UnsubscribeContext{ProjectName: "Acme", Email: "a@b.co"}

	out := renderBrandedUnsubscribePage(tpl, ctx, unsubPageConfirm, "p-1", "s-1", "tok")

	if !strings.Contains(out, `action="/unsubscribe/p-1/s-1?t=tok"`) {
		t.Fatalf("confirm form was not force-injected: %q", out)
	}
	if strings.Index(out, "action=") > strings.Index(strings.ToLower(out), "</body>") {
		t.Fatalf("form injected after body close: %q", out)
	}
}

func TestRenderBrandedPageDoneMode(t *testing.T) {
	tpl := `<p>{{project_name}}</p>{{confirm_button}}`
	ctx := service.UnsubscribeContext{ProjectName: "Acme", Email: "a@b.co"}

	out := renderBrandedUnsubscribePage(tpl, ctx, unsubPageDone, "p-1", "s-1", "tok")

	if strings.Contains(out, "<form") {
		t.Fatalf("done page must not carry a form: %q", out)
	}
	if !strings.Contains(out, "You have been unsubscribed") {
		t.Fatalf("done confirmation missing: %q", out)
	}
}

func TestRenderBrandedPageFallbacksAndWrapping(t *testing.T) {
	tpl := `<p>{{newsletter_name}}</p>{{manage_preferences}}{{confirm_button}}`
	ctx := service.UnsubscribeContext{ProjectName: "Acme", Email: "a@b.co"}

	out := renderBrandedUnsubscribePage(tpl, ctx, unsubPageConfirm, "p-1", "s-1", "tok")

	if !strings.Contains(out, "<p>Acme</p>") {
		t.Fatalf("newsletter name must fall back to project name: %q", out)
	}
	if strings.Contains(out, "{{manage_preferences}}") {
		t.Fatalf("reserved placeholder leaked: %q", out)
	}
	if !strings.HasPrefix(out, "<!DOCTYPE html>") {
		t.Fatalf("fragment was not wrapped in a document: %q", out[:60])
	}
}
