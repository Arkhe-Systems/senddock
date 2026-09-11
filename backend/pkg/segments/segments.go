package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

var (
	ErrInvalidPredicate = errors.New("invalid segment predicate")
	keyPattern          = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	datePattern         = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// comparisonCast picks the SQL cast for a greater/less-than rule. Custom field values live
// in the metadata JSON as text, so the cast has to match the field: a date stored as
// YYYY-MM-DD cast to numeric fails when the query runs, which turns a saved segment into a
// preview error and a broadcast that never sends. The rule value carries the format the
// picker produced, and that decides which cast applies.
func comparisonCast(value any) string {
	if isDateValue(value) {
		return "::date"
	}
	return "::numeric"
}

func isDateValue(value any) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	if !datePattern.MatchString(s) {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// comparableValue reports whether a greater/less-than rule carries something the database
// can order, so a rule that could only fail later is rejected while the segment is saved.
func comparableValue(value any) bool {
	switch v := value.(type) {
	case float64, int, int64:
		return true
	case string:
		if isDateValue(v) {
			return true
		}
		_, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return err == nil
	}
	return false
}

type Predicate struct {
	Match string `json:"match"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

func ParsePredicate(raw json.RawMessage) (Predicate, error) {
	pred := Predicate{Match: "all"}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &pred); err != nil {
			return pred, fmt.Errorf("%w: not valid json", ErrInvalidPredicate)
		}
	}
	if pred.Match == "" {
		pred.Match = "all"
	}
	if pred.Match != "all" && pred.Match != "any" {
		return pred, fmt.Errorf("%w: match must be 'all' or 'any'", ErrInvalidPredicate)
	}
	for _, rule := range pred.Rules {
		if err := validateRule(rule); err != nil {
			return pred, err
		}
	}
	return pred, nil
}

func validateRule(rule Rule) error {
	switch {
	case rule.Field == "status":
		if rule.Op != "eq" && rule.Op != "neq" {
			return fmt.Errorf("%w: status supports eq/neq", ErrInvalidPredicate)
		}
	case rule.Field == "tags":
		if rule.Op != "includes_any" && rule.Op != "includes_all" && rule.Op != "excludes" {
			return fmt.Errorf("%w: tags supports includes_any/includes_all/excludes", ErrInvalidPredicate)
		}
	case strings.HasPrefix(rule.Field, "custom."):
		key := strings.TrimPrefix(rule.Field, "custom.")
		if !keyPattern.MatchString(key) {
			return fmt.Errorf("%w: invalid custom field key", ErrInvalidPredicate)
		}
		switch rule.Op {
		case "eq", "neq", "contains":
		case "gt", "lt":
			if !comparableValue(rule.Value) {
				return fmt.Errorf("%w: greater/less than needs a number or a YYYY-MM-DD date", ErrInvalidPredicate)
			}
		default:
			return fmt.Errorf("%w: custom fields support eq/neq/contains/gt/lt", ErrInvalidPredicate)
		}
	default:
		return fmt.Errorf("%w: unknown field %s", ErrInvalidPredicate, rule.Field)
	}
	return nil
}

func toStringSlice(v any) []string {
	switch vv := v.(type) {
	case []string:
		return vv
	case []any:
		out := make([]string, 0, len(vv))
		for _, e := range vv {
			out = append(out, fmt.Sprintf("%v", e))
		}
		return out
	case string:
		return []string{vv}
	}
	return nil
}

func buildRuleSQL(rule Rule, argIdx int) (string, []any) {
	placeholder := fmt.Sprintf("$%d", argIdx)
	switch {
	case rule.Field == "status":
		if rule.Op == "neq" {
			return "status <> " + placeholder, []any{fmt.Sprintf("%v", rule.Value)}
		}
		return "status = " + placeholder, []any{fmt.Sprintf("%v", rule.Value)}
	case rule.Field == "tags":
		tags := toStringSlice(rule.Value)
		switch rule.Op {
		case "includes_all":
			return "tags @> " + placeholder + "::text[]", []any{pq.Array(tags)}
		case "excludes":
			return "NOT (tags && " + placeholder + "::text[])", []any{pq.Array(tags)}
		default:
			return "tags && " + placeholder + "::text[]", []any{pq.Array(tags)}
		}
	default:
		key := strings.TrimPrefix(rule.Field, "custom.")
		column := fmt.Sprintf("metadata->>'%s'", key)
		switch rule.Op {
		case "neq":
			return column + " IS DISTINCT FROM " + placeholder, []any{fmt.Sprintf("%v", rule.Value)}
		case "contains":
			return column + " ILIKE '%' || " + placeholder + " || '%'", []any{fmt.Sprintf("%v", rule.Value)}
		case "gt":
			cast := comparisonCast(rule.Value)
			return "(" + column + ")" + cast + " > " + placeholder + cast, []any{fmt.Sprintf("%v", rule.Value)}
		case "lt":
			cast := comparisonCast(rule.Value)
			return "(" + column + ")" + cast + " < " + placeholder + cast, []any{fmt.Sprintf("%v", rule.Value)}
		default:
			return column + " = " + placeholder, []any{fmt.Sprintf("%v", rule.Value)}
		}
	}
}

func BuildWhere(pred Predicate, startArg int) (string, []any) {
	if len(pred.Rules) == 0 {
		return "", nil
	}
	joiner := " AND "
	if pred.Match == "any" {
		joiner = " OR "
	}
	fragments := make([]string, 0, len(pred.Rules))
	args := make([]any, 0, len(pred.Rules))
	argIdx := startArg
	for _, rule := range pred.Rules {
		frag, ruleArgs := buildRuleSQL(rule, argIdx)
		fragments = append(fragments, frag)
		args = append(args, ruleArgs...)
		argIdx += len(ruleArgs)
	}
	return strings.Join(fragments, joiner), args
}
