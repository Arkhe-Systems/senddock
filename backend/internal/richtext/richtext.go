package richtext

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlPolicy  *bluemonday.Policy
	plainPolicy *bluemonday.Policy
	once        sync.Once
)

func policies() (*bluemonday.Policy, *bluemonday.Policy) {
	once.Do(func() {
		p := bluemonday.NewPolicy()
		p.AllowElements(
			"p", "br", "span", "div",
			"strong", "b", "em", "i", "u", "s",
			"ul", "ol", "li",
			"h1", "h2", "h3",
			"blockquote", "pre", "code",
			"hr",
		)
		p.AllowAttrs("href").OnElements("a")
		p.AllowStandardURLs()
		p.RequireNoFollowOnLinks(true)
		p.AddTargetBlankToFullyQualifiedLinks(true)
		p.RequireNoReferrerOnLinks(true)
		htmlPolicy = p
		plainPolicy = bluemonday.StrictPolicy()
	})
	return htmlPolicy, plainPolicy
}

// Sanitize returns email-safe HTML: formatting and links survive, scripts,
// event handlers, styles and unknown tags are stripped.
func Sanitize(s string) string {
	html, _ := policies()
	return html.Sanitize(s)
}

// PlainText strips every tag, leaving only the text. Used where HTML makes no
// sense, such as the subject line.
func PlainText(s string) string {
	_, plain := policies()
	return plain.Sanitize(s)
}
