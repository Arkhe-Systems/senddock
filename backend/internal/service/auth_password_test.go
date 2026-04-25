package service

import "testing"

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name      string
		password  string
		wantError bool
	}{
		{"too short", "Aa1!", true},
		{"missing uppercase", "lowercase1!", true},
		{"missing number", "Password!", true},
		{"missing special", "Password1", true},
		{"all numbers (no letter, no special)", "12345678", true},
		{"common weak password", "password", true},
		{"qwerty", "qwertyuiop", true},
		{"all same letter", "aaaaaaaa", true},
		{"valid simple", "Password1!", false},
		{"valid mixed", "MyApp2026!", false},
		{"valid with symbols", "Test@1234", false},
		{"valid long", "Correct-Horse-Battery-Staple-1", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if tc.wantError && err == nil {
				t.Errorf("expected error for %q, got nil", tc.password)
			}
			if !tc.wantError && err != nil {
				t.Errorf("unexpected error for %q: %v", tc.password, err)
			}
		})
	}
}
