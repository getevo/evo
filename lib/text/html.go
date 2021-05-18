package text

import "regexp"

var htmlLineSeparator = regexp.MustCompile(`(?m)(<\/p>|<br.*\/>|<hr.*\/>)`)
var htmlTag = regexp.MustCompile(`(?m)<\/{0,1}[a-zA-Z0-9]+.*?\/{0,1}>`)

func FromHTML(html string) string {
	html = htmlLineSeparator.ReplaceAllString(html, "\n")
	html = htmlTag.ReplaceAllString(html, "")
	return html
}
