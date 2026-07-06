package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

var (
	ErrInvalidPredicate = errors.New("invalid segment predicate")
	keyPattern          = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

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
		case "eq", "neq", "contains", "gt", "lt":
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
			return "(" + column + ")::numeric > " + placeholder, []any{rule.Value}
		case "lt":
			return "(" + column + ")::numeric < " + placeholder, []any{rule.Value}
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
