package analytics

import "testing"

func TestDeviceClass(t *testing.T) {
	cases := map[string]string{
		"": "unknown",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Mobile/15E148": "mobile",
		"Mozilla/5.0 (Linux; Android 14) Mobile Safari/537.36":                 "mobile",
		"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X)":                        "tablet",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/131.0":         "desktop",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)":                            "desktop",
	}
	for ua, want := range cases {
		if got := deviceClass(ua); got != want {
			t.Errorf("deviceClass(%q) = %q, want %q", ua, got, want)
		}
	}
}

func TestMailClient(t *testing.T) {
	cases := map[string]string{
		"":                            "Unknown",
		"Mozilla/5.0 GoogleImageProxy": "Gmail",
		"Outlook-iOS/2.0":              "Outlook",
		"Mozilla/5.0 (Windows) Firefox/120": "Firefox",
	}
	for ua, want := range cases {
		if got := mailClient(ua); got != want {
			t.Errorf("mailClient(%q) = %q, want %q", ua, got, want)
		}
	}
}
