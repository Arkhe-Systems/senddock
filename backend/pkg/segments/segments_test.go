package segments

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParsePredicateValidation(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"empty defaults to all", ``, false},
		{"valid all", `{"match":"all","rules":[{"field":"status","op":"eq","value":"active"}]}`, false},
		{"valid tags", `{"match":"any","rules":[{"field":"tags","op":"includes_any","value":["vip"]}]}`, false},
		{"valid custom", `{"rules":[{"field":"custom.plan","op":"eq","value":"pro"}]}`, false},
		{"bad match", `{"match":"none","rules":[]}`, true},
		{"bad status op", `{"rules":[{"field":"status","op":"contains","value":"x"}]}`, true},
		{"bad tags op", `{"rules":[{"field":"tags","op":"eq","value":["x"]}]}`, true},
		{"unknown field", `{"rules":[{"field":"foo","op":"eq","value":"x"}]}`, true},
		{"bad custom key", `{"rules":[{"field":"custom.a-b","op":"eq","value":"x"}]}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePredicate(json.RawMessage(tc.raw))
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParsePredicate(%s) err = %v, wantErr = %v", tc.raw, err, tc.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidPredicate) {
				t.Fatalf("expected ErrInvalidPredicate, got %v", err)
			}
		})
	}
}

func TestBuildWherePlaceholders(t *testing.T) {
	pred := Predicate{
		Match: "all",
		Rules: []Rule{
			{Field: "status", Op: "eq", Value: "active"},
			{Field: "tags", Op: "includes_any", Value: []any{"vip", "customer"}},
			{Field: "custom.plan", Op: "eq", Value: "pro"},
		},
	}
	where, args := BuildWhere(pred, 2)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	for _, ph := range []string{"$2", "$3", "$4"} {
		if !strings.Contains(where, ph) {
			t.Fatalf("expected placeholder %s in %q", ph, where)
		}
	}
	if !strings.Contains(where, " AND ") {
		t.Fatalf("expected AND joiner for match=all, got %q", where)
	}
	if !strings.Contains(where, "metadata->>'plan'") {
		t.Fatalf("expected custom field expression, got %q", where)
	}
}

func TestBuildWhereAnyJoiner(t *testing.T) {
	pred := Predicate{
		Match: "any",
		Rules: []Rule{
			{Field: "status", Op: "eq", Value: "active"},
			{Field: "status", Op: "neq", Value: "pending"},
		},
	}
	where, _ := BuildWhere(pred, 2)
	if !strings.Contains(where, " OR ") {
		t.Fatalf("expected OR joiner for match=any, got %q", where)
	}
}

func TestBuildWhereEmpty(t *testing.T) {
	where, args := BuildWhere(Predicate{Match: "all"}, 2)
	if where != "" || args != nil {
		t.Fatalf("expected empty where for no rules, got %q / %v", where, args)
	}
}
