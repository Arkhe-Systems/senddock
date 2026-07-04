package service

import (
	"encoding/json"
	"testing"

	"github.com/arkhe-systems/senddock/internal/db"
)

func subWithMeta(meta string) db.Subscriber {
	return db.Subscriber{Metadata: json.RawMessage(meta)}
}

func TestReplaceCustomVariables(t *testing.T) {
	sub := subWithMeta(`{"plan_tier":"pro","country":"CO","age":30}`)

	cases := []struct {
		name   string
		text   string
		escape bool
		want   string
	}{
		{"single string field", "Hi {{custom.plan_tier}}!", false, "Hi pro!"},
		{"multiple fields", "{{custom.plan_tier}} / {{custom.country}}", false, "pro / CO"},
		{"number rendered", "Age: {{custom.age}}", false, "Age: 30"},
		{"unknown key left intact", "{{custom.missing}}", false, "{{custom.missing}}"},
		{"no custom token untouched", "Hi {{name}}", false, "Hi {{name}}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := replaceCustomVariables(tc.text, sub, tc.escape); got != tc.want {
				t.Fatalf("replaceCustomVariables(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}

func TestReplaceCustomVariablesEscaping(t *testing.T) {
	sub := subWithMeta(`{"bio":"<b>hi</b>"}`)
	if got := replaceCustomVariables("{{custom.bio}}", sub, true); got != "&lt;b&gt;hi&lt;/b&gt;" {
		t.Fatalf("escaped body render = %q", got)
	}
	if got := replaceCustomVariables("{{custom.bio}}", sub, false); got != "<b>hi</b>" {
		t.Fatalf("unescaped (subject) render = %q", got)
	}
}

func TestReplaceCustomVariablesNoMetadata(t *testing.T) {
	if got := replaceCustomVariables("{{custom.x}}", db.Subscriber{}, true); got != "{{custom.x}}" {
		t.Fatalf("empty metadata should leave text intact, got %q", got)
	}
	if got := replaceCustomVariables("{{custom.x}}", subWithMeta(`not json`), true); got != "{{custom.x}}" {
		t.Fatalf("invalid metadata should leave text intact, got %q", got)
	}
}
