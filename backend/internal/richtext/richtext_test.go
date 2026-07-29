package richtext

import "testing"

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
