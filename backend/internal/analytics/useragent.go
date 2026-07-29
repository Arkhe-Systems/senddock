package analytics

import "strings"

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
