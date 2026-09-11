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
		{"custom gt with a date", `{"rules":[{"field":"custom.signup","op":"gt","value":"2026-01-01"}]}`, false},
		{"custom gt with a number", `{"rules":[{"field":"custom.seats","op":"gt","value":42}]}`, false},
		{"custom gt with a numeric string", `{"rules":[{"field":"custom.seats","op":"gt","value":"42"}]}`, false},
		{"custom lt with a date", `{"rules":[{"field":"custom.signup","op":"lt","value":"2026-06-30"}]}`, false},
		// A value the database cannot order would only fail when the segment is evaluated,
		// so it has to be refused at save time.
		{"custom gt with text", `{"rules":[{"field":"custom.plan","op":"gt","value":"pro"}]}`, true},
		{"custom lt with an impossible date", `{"rules":[{"field":"custom.signup","op":"lt","value":"2026-13-45"}]}`, true},
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

// A date field stored as YYYY-MM-DD has to be compared as a date. Casting it to numeric is
// what made a saved segment fail at query time, taking the preview and the broadcast with it.
func TestBuildWhereCastsComparisonsByValue(t *testing.T) {
	cases := []struct {
		name     string
		value    any
		wantCast string
	}{
		{"date", "2026-01-01", "::date"},
		{"padded date", " 2026-01-01 ", "::date"},
		{"number", float64(42), "::numeric"},
		{"numeric string", "42", "::numeric"},
		{"decimal string", "4.5", "::numeric"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			where, args := BuildWhere(Predicate{
				Match: "all",
				Rules: []Rule{{Field: "custom.f", Op: "gt", Value: tc.value}},
			}, 1)

			if !strings.Contains(where, tc.wantCast) {
				t.Fatalf("comparison %v compiled without %s: %q", tc.value, tc.wantCast, where)
			}
			other := "::numeric"
			if tc.wantCast == "::numeric" {
				other = "::date"
			}
			if strings.Contains(where, other) {
				t.Fatalf("comparison %v compiled with the wrong cast %s: %q", tc.value, other, where)
			}
			if len(args) != 1 {
				t.Fatalf("expected one argument, got %d", len(args))
			}
		})
	}
}
