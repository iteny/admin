package service

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var htmlPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "blockquote", "pre", "code", "strong", "b", "em", "i",
		"br", "hr", "span", "div", "table", "thead", "tbody", "tr", "td", "th",
		"figure", "figcaption", "img", "a")
	p.AllowAttrs("href", "title", "target", "rel").OnElements("a")
	p.AllowAttrs("src", "alt", "title", "width", "height", "loading").OnElements("img")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)).Globally()
	p.AllowAttrs("style").Matching(regexp.MustCompile(`^[a-zA-Z0-9#%.,;:\s\-()]+$`)).Globally()
	p.AllowAttrs("colspan", "rowspan").OnElements("td", "th")
	p.AllowStandardURLs()
	p.RequireParseableURLs(true)
	p.AllowRelativeURLs(true)
	p.RequireNoFollowOnLinks(false)
	return p
}()

var unsafeStyle = regexp.MustCompile(`(?i)\sstyle="[^"]*(?:url\(|expression|javascript:|@import|behavior|-moz-binding)[^"]*"`)

// SanitizeHTML 清洗富文本，只保留受信任标签和属性。
func SanitizeHTML(input string) string {
	clean := htmlPolicy.Sanitize(input)
	clean = unsafeStyle.ReplaceAllString(clean, "")
	return strings.TrimSpace(clean)
}

// SanitizeText 清理纯文本输入中的控制字符。
func SanitizeText(input string) string {
	return strings.TrimSpace(strings.ReplaceAll(input, "\x00", ""))
}
