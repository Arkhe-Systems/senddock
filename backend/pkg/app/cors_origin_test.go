package app

import "testing"

func TestResolveCORSOrigin(t *testing.T) {
	cases := []struct {
		name        string
		frontendURL string
		publicURL   string
		want        string
	}{
		{
			name:        "dashboard origin wins over the public url",
			frontendURL: "http://localhost:5173",
			publicURL:   "http://localhost:8080",
			want:        "http://localhost:5173",
		},
		{
			name:        "public url is used when no dashboard origin is configured",
			frontendURL: "",
			publicURL:   "https://mail.example.com",
			want:        "https://mail.example.com",
		},
		{
			name:        "falls back to the dev origin when nothing is configured",
			frontendURL: "",
			publicURL:   "",
			want:        devFrontendOrigin,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCORSOrigin(tc.frontendURL, tc.publicURL); got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
