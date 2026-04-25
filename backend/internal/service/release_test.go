package service

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		remote string
		local  string
		want   bool
	}{
		{"0.4.0", "0.4.0", false},
		{"0.4.1", "0.4.0", true},
		{"0.4.0", "0.4.1", false},
		{"1.0.0", "0.99.0", true},
		{"0.5.0", "0.4.99", true},
		{"0.4.0", "1.0.0", false},
		{"0.4.0", "0.4", false},
		{"0.4.0", "0.4.0-beta", false},
		{"v0.4.1", "0.4.0", true},
	}

	for _, tc := range cases {
		t.Run(tc.remote+"_vs_"+tc.local, func(t *testing.T) {
			got := isNewer(tc.remote, tc.local)
			if got != tc.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tc.remote, tc.local, got, tc.want)
			}
		})
	}
}

func TestSplitVersion(t *testing.T) {
	cases := []struct {
		input string
		want  [3]int
	}{
		{"0.4.0", [3]int{0, 4, 0}},
		{"1.2.3", [3]int{1, 2, 3}},
		{"0.4", [3]int{0, 4, 0}},
		{"v1.2.3", [3]int{0, 2, 3}},
		{"0.4.0-beta", [3]int{0, 4, 0}},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := splitVersion(tc.input)
			if got != tc.want {
				t.Errorf("splitVersion(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
