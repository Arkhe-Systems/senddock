package richtext

import (
	"strings"
	"testing"
)

func TestSanitizeKeepsFormatting(t *testing.T) {
	in := `<p>Hola <strong>mundo</strong> y <em>saltos</em></p><ul><li>uno</li><li>dos</li></ul>`
	got := Sanitize(in)
	for _, want := range []string{"<strong>mundo</strong>", "<em>saltos</em>", "<li>uno</li>"} {
		if !contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestSanitizeStripsScripts(t *testing.T) {
	in := `<p>ok</p><script>alert(1)</script><img src=x onerror=alert(1)>`
	got := Sanitize(in)
	if contains(got, "script") || contains(got, "onerror") {
		t.Fatalf("dangerous markup survived: %q", got)
	}
	if !contains(got, "<p>ok</p>") {
		t.Fatalf("safe markup dropped: %q", got)
	}
}

func TestSanitizeAllowsSafeLinks(t *testing.T) {
	got := Sanitize(`<a href="https://example.com">x</a><a href="javascript:alert(1)">y</a>`)
	if !contains(got, `href="https://example.com"`) {
		t.Fatalf("safe link dropped: %q", got)
	}
	if contains(got, "javascript:") {
		t.Fatalf("javascript link survived: %q", got)
	}
}

func TestSanitizePage_PreservesStyleBlockInHead(t *testing.T) {
	in := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Unsubscribe</title><style>.btn{color:#00af7b}</style></head><body><h1>Unsubscribe</h1><script>alert(1)</script></body></html>`
	out := SanitizePage(in)

	if !strings.Contains(out, `.btn{color:#00af7b}`) {
		t.Fatalf("style block dropped: %q", out)
	}
	if strings.Contains(out, "<script") || strings.Contains(out, "alert") {
		t.Fatalf("script survived: %q", out)
	}
	if !strings.Contains(out, "<h1>Unsubscribe</h1>") {
		t.Fatalf("body content dropped: %q", out)
	}
}

func TestSanitizePage_NoStyleBlockUnchanged(t *testing.T) {
	in := `<div>no style</div>`
	if out := SanitizePage(in); !strings.Contains(out, "no style") {
		t.Fatalf("content dropped without style block: %q", out)
	}
}

func TestPlainTextStripsEverything(t *testing.T) {
	got := PlainText(`<b>Hola</b> <a href="#">x</a>`)
	if contains(got, "<") {
		t.Fatalf("tags survived in plain text: %q", got)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestSanitizePage(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		mustKeep []string
		mustDrop []string
	}{
		{"strips script", `<div>hi</div><script>alert(1)</script>`, []string{"<div>hi</div>"}, []string{"script", "alert"}},
		{"strips iframe", `<p>x</p><iframe src="https://evil.example"></iframe>`, []string{"<p>x</p>"}, []string{"iframe"}},
		{"strips event handlers", `<img src="https://x.example/a.png" onerror="alert(1)">`, []string{"<img"}, []string{"onerror"}},
		{"strips forms and inputs", `<form action="/steal"><input name="pw"><button>go</button></form>`, nil, []string{"<form", "<input", "<button"}},
		{"strips javascript urls", `<a href="javascript:alert(1)">x</a>`, nil, []string{"javascript:"}},
		{"keeps tables", `<table><tr><td colspan="2">x</td></tr></table>`, []string{"<table>", `colspan="2"`}, nil},
		{"keeps inline styles", `<div style="background:#000;color:#fff">x</div>`, []string{`style="background:#000;color:#fff"`}, nil},
		{"keeps class and id attributes", `<div class="card highlight" id="hero">x</div>`, []string{`class="card highlight"`, `id="hero"`}, nil},
		{"keeps style element content", `<style>.a{color:red}</style><p>x</p>`, []string{`<style>.a{color:red}</style>`, "<p>x</p>"}, nil},
		{"keeps deep headings and images", `<h5>t</h5><img src="https://x.example/logo.png" width="120">`, []string{"<h5>t</h5>", `width="120"`}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := SanitizePage(tc.in)
			for _, want := range tc.mustKeep {
				if !strings.Contains(out, want) {
					t.Errorf("expected output to keep %q, got %q", want, out)
				}
			}
			for _, bad := range tc.mustDrop {
				if strings.Contains(out, bad) {
					t.Errorf("expected output to drop %q, got %q", bad, out)
				}
			}
		})
	}
}
