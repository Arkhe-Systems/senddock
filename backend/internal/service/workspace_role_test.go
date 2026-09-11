package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func TestNormalizeRoleRejectsEmptyAndUnknown(t *testing.T) {
	cases := map[string]string{
		"owner":     WorkspaceRoleOwner,
		"admin":     WorkspaceRoleAdmin,
		"developer": WorkspaceRoleDeveloper,
		"viewer":    WorkspaceRoleViewer,
		" Admin ":   WorkspaceRoleAdmin,
		// Empty and unknown roles must normalize to "" so callers reject them with
		// ErrInvalidRole instead of quietly picking a role for the caller.
		"":        "",
		"   ":     "",
		"bogus":   "",
		"root":    "",
		"manager": "",
		// "member" is not a role the database accepts, so treating it as valid could
		// only ever end in a constraint violation.
		"member": "",
		"MEMBER": "",
	}
	for input, want := range cases {
		if got := normalizeRole(input); got != want {
			t.Errorf("normalizeRole(%q) = %q, want %q", input, got, want)
		}
	}
}

// Every role except owner can strip ownership from a workspace, so each one has to run the
// last-owner guard. Tying the guard to a single role is what left demoted workspaces
// without an owner.
func TestDemotionNeedsOwnerGuard(t *testing.T) {
	cases := map[string]bool{
		WorkspaceRoleOwner:     false,
		WorkspaceRoleAdmin:     true,
		WorkspaceRoleDeveloper: true,
		WorkspaceRoleViewer:    true,
	}
	for role, want := range cases {
		if got := demotionNeedsOwnerGuard(role); got != want {
			t.Errorf("demotionNeedsOwnerGuard(%q) = %v, want %v", role, got, want)
		}
	}
}

// A role the database refuses is a bad request, not a server failure.
func TestIsCheckConstraintViolation(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"check violation", &pq.Error{Code: "23514"}, true},
		{"wrapped check violation", fmt.Errorf("update member role: %w", &pq.Error{Code: "23514"}), true},
		{"unique violation", &pq.Error{Code: "23505"}, false},
		{"plain error", errors.New("boom"), false},
	}
	for _, tc := range cases {
		if got := isCheckConstraintViolation(tc.err); got != tc.want {
			t.Errorf("%s: isCheckConstraintViolation(%v) = %v, want %v", tc.name, tc.err, got, tc.want)
		}
	}
}
