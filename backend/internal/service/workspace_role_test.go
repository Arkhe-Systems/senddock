package service

import "testing"

func TestNormalizeRoleRejectsEmptyAndUnknown(t *testing.T) {
	cases := map[string]string{
		"owner":     WorkspaceRoleOwner,
		"admin":     WorkspaceRoleAdmin,
		"developer": WorkspaceRoleDeveloper,
		"viewer":    WorkspaceRoleViewer,
		"member":    WorkspaceRoleMember,
		"MEMBER":    WorkspaceRoleMember,
		" Admin ":   WorkspaceRoleAdmin,
		// empty and unknown roles must normalize to "" so callers reject them
		// instead of silently defaulting to the near-admin "member" role.
		"":        "",
		"   ":     "",
		"bogus":   "",
		"root":    "",
		"manager": "",
	}
	for input, want := range cases {
		if got := normalizeRole(input); got != want {
			t.Errorf("normalizeRole(%q) = %q, want %q", input, got, want)
		}
	}
}
