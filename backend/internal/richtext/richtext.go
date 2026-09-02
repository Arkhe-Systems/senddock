package richtext

import (
	"regexp"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlPolicy  *bluemonday.Policy
	plainPolicy *bluemonday.Policy
	pagePolicy  *bluemonday.Policy
	once        sync.Once
)

func policies() (*bluemonday.Policy, *bluemonday.Policy, *bluemonday.Policy) {
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

		pg := bluemonday.NewPolicy()
		pg.AllowElements(
			"html", "head", "body", "title",
			"p", "br", "span", "div", "center", "section", "header", "footer", "main",
			"strong", "b", "em", "i", "u", "s", "small", "sub", "sup",
			"ul", "ol", "li",
			"h1", "h2", "h3", "h4", "h5", "h6",
			"blockquote", "pre", "code",
			"hr", "img",
			"table", "thead", "tbody", "tfoot", "tr", "td", "th", "caption",
			"font",
		)
		pg.AllowAttrs("href").OnElements("a")
		pg.AllowAttrs("src", "alt", "width", "height").OnElements("img")
		pg.AllowAttrs("colspan", "rowspan", "align", "valign", "width", "height", "bgcolor", "cellpadding", "cellspacing", "border").OnElements("table", "tr", "td", "th")
		pg.AllowAttrs("face", "size", "color").OnElements("font")
		pg.AllowAttrs("style").Globally()
		pg.AllowAttrs("class", "id").Globally()
		pg.AllowAttrs("charset", "name", "content").OnElements("meta")
		pg.AllowElements("meta")
		pg.AllowStandardURLs()
		pg.AllowDataURIImages()
		pagePolicy = pg
	})
	return htmlPolicy, plainPolicy, pagePolicy
}

func Sanitize(s string) string {
	html, _, _ := policies()
	return html.Sanitize(s)
}

func PlainText(s string) string {
	_, plain, _ := policies()
	return plain.Sanitize(s)
}

var styleBlockRe = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)

func SanitizePage(s string) string {
	_, _, page := policies()

	// bluemonday always drops <style> content (it treats it as an unsafe
	// raw-text element), but a branded page template may carry a stylesheet.
	// Extract the blocks, sanitize the rest, then re-insert them untouched.
	styles := styleBlockRe.FindAllString(s, -1)
	stripped := styleBlockRe.ReplaceAllString(s, "")
	out := page.Sanitize(stripped)

	if len(styles) == 0 {
		return out
	}

	head := strings.Join(styles, "\n")
	if idx := strings.Index(strings.ToLower(out), "</head>"); idx != -1 {
		return out[:idx] + head + out[idx:]
	}
	return head + out
}
