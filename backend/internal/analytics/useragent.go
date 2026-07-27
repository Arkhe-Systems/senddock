package analytics

import "strings"

// deviceClass buckets a user-agent into desktop / mobile / tablet / unknown.
// This is a coarse heuristic, not a full UA parser — enough to answer "are my
// readers on phones or desktops" without pulling in a dependency.
func deviceClass(ua string) string {
	s := strings.ToLower(ua)
	if s == "" {
		return "unknown"
	}
	switch {
	case strings.Contains(s, "ipad"), strings.Contains(s, "tablet"):
		return "tablet"
	case strings.Contains(s, "mobi"), strings.Contains(s, "iphone"), strings.Contains(s, "android"):
		return "mobile"
	default:
		return "desktop"
	}
}

// mailClient identifies the reading environment from a user-agent. Open/click
// user-agents are usually the proxy or webmail that fetched the pixel, so the
// useful signal is Apple Mail Privacy, Gmail, Outlook, Yahoo, or a plain browser.
func mailClient(ua string) string {
	s := strings.ToLower(ua)
	switch {
	case s == "":
		return "Unknown"
	case strings.Contains(s, "googleimageproxy"):
		return "Gmail"
	case strings.Contains(s, "applemail"), strings.Contains(s, "apple mail"):
		return "Apple Mail"
	case strings.Contains(s, "outlook"), strings.Contains(s, "microsoft"):
		return "Outlook"
	case strings.Contains(s, "yahoo"):
		return "Yahoo Mail"
	case strings.Contains(s, "thunderbird"):
		return "Thunderbird"
	case strings.Contains(s, "iphone"), strings.Contains(s, "ipad"), strings.Contains(s, "macintosh") && strings.Contains(s, "mail"):
		return "Apple Mail"
	case strings.Contains(s, "chrome"):
		return "Chrome"
	case strings.Contains(s, "firefox"):
		return "Firefox"
	case strings.Contains(s, "safari"):
		return "Safari"
	default:
		return "Other"
	}
}
