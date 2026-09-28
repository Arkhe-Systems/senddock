package service

import (
	"strings"
	"testing"
)

var mimeHeaderPrefixes = []string{
	"From:",
	"To:",
	"Subject:",
	"MIME-Version:",
	"Content-Type:",
}

// A subject is built from subscriber-supplied text, so a value carrying a line break
// must not be able to close the Subject line and start a header of its own.
func TestBuildMIMERejectsHeaderInjection(t *testing.T) {
	msg := string(buildMIME(
		"Sender <from@example.test>",
		"recipient@example.test",
		"Hello\r\nBcc: victim@example.test\r\nX-Injected: 1",
		"https://example.test/unsubscribe?t=abc\r\nBcc: victim2@example.test",
		"<p>body</p>",
	))

	headerBlock, body, found := strings.Cut(msg, "\r\n\r\n")
	if !found {
		t.Fatalf("message has no header/body separator:\n%q", msg)
	}

	want := append(append([]string{}, mimeHeaderPrefixes...), "List-Unsubscribe:", "List-Unsubscribe-Post:")
	lines := strings.Split(headerBlock, "\r\n")
	if len(lines) != len(want) {
		t.Fatalf("message carries %d header lines, want %d — an injected header got through:\n%s", len(lines), len(want), headerBlock)
	}
	for i, prefix := range want {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("header %d is %q, want a line starting with %q", i, lines[i], prefix)
		}
	}
	if body != "<p>body</p>" {
		t.Errorf("body = %q, want it untouched", body)
	}
}

// The fix strips line breaks rather than truncating, so the rest of the value survives
// as part of the same header.
func TestBuildMIMEKeepsInjectedTextInsideTheHeaderValue(t *testing.T) {
	msg := string(buildMIME("from@example.test", "to@example.test",
		"Hi\r\nBcc: victim@example.test", "", "<p>body</p>"))

	headerBlock, _, _ := strings.Cut(msg, "\r\n\r\n")
	if !strings.Contains(headerBlock, "Subject: HiBcc: victim@example.test") {
		t.Fatalf("subject value was not preserved on one line:\n%s", headerBlock)
	}
}

// The ordinary shape must not change: headers, blank line, body.
func TestBuildMIMEKeepsTheOrdinaryShape(t *testing.T) {
	bare := string(buildMIME("from@example.test", "to@example.test", "Subject", "", "<p>hi</p>"))
	headerBlock, body, found := strings.Cut(bare, "\r\n\r\n")
	if !found {
		t.Fatalf("no separator in:\n%q", bare)
	}
	if lines := strings.Split(headerBlock, "\r\n"); len(lines) != len(mimeHeaderPrefixes) {
		t.Fatalf("message without an unsubscribe URL carries %d headers, want %d:\n%s", len(lines), len(mimeHeaderPrefixes), headerBlock)
	}
	if body != "<p>hi</p>" {
		t.Errorf("body = %q", body)
	}
}
