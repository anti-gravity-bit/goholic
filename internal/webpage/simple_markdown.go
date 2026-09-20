package webpage

import (
	"html"
	"strings"
)

// TurnMarkdownIntoSafeHtml turns a small set of markdown marks into HTML.
//
// A five year old version of the rules:
// - a line that starts with "# " is a big heading
// - a line that starts with "## " is a smaller heading
// - a line that starts with "- " is a list item
// - text between **stars** is strong
// - text between `ticks` is code
// - [label](https://...) becomes a link; only http(s) URLs
// - blank lines split paragraphs
func TurnMarkdownIntoSafeHtml(articleBodyMarkdown string) string {
	trimmedMarkdown := strings.ReplaceAll(articleBodyMarkdown, "\r\n", "\n")
	markdownLines := strings.Split(trimmedMarkdown, "\n")

	var htmlBuilder strings.Builder
	insideAnUnorderedList := false

	flushListIfOpen := func() {
		if insideAnUnorderedList {
			htmlBuilder.WriteString("</ul>\n")
			insideAnUnorderedList = false
		}
	}

	for _, rawMarkdownLine := range markdownLines {
		markdownLine := strings.TrimRight(rawMarkdownLine, " ")

		if strings.TrimSpace(markdownLine) == "" {
			flushListIfOpen()
			continue
		}

		if strings.HasPrefix(markdownLine, "# ") {
			flushListIfOpen()
			htmlBuilder.WriteString("<h1>")
			htmlBuilder.WriteString(decorateInlineMarkdown(markdownLine[2:]))
			htmlBuilder.WriteString("</h1>\n")
			continue
		}

		if strings.HasPrefix(markdownLine, "## ") {
			flushListIfOpen()
			htmlBuilder.WriteString("<h2>")
			htmlBuilder.WriteString(decorateInlineMarkdown(markdownLine[3:]))
			htmlBuilder.WriteString("</h2>\n")
			continue
		}

		if strings.HasPrefix(markdownLine, "### ") {
			flushListIfOpen()
			htmlBuilder.WriteString("<h3>")
			htmlBuilder.WriteString(decorateInlineMarkdown(markdownLine[4:]))
			htmlBuilder.WriteString("</h3>\n")
			continue
		}

		if strings.HasPrefix(markdownLine, "- ") {
			if !insideAnUnorderedList {
				htmlBuilder.WriteString("<ul>\n")
				insideAnUnorderedList = true
			}

			htmlBuilder.WriteString("<li>")
			htmlBuilder.WriteString(decorateInlineMarkdown(markdownLine[2:]))
			htmlBuilder.WriteString("</li>\n")
			continue
		}

		flushListIfOpen()
		htmlBuilder.WriteString("<p>")
		htmlBuilder.WriteString(decorateInlineMarkdown(markdownLine))
		htmlBuilder.WriteString("</p>\n")
	}

	flushListIfOpen()
	return htmlBuilder.String()
}

func decorateInlineMarkdown(rawLine string) string {
	escapedLine := html.EscapeString(rawLine)
	escapedLine = replaceWrappedMarks(escapedLine, "**", "<strong>", "</strong>")
	escapedLine = replaceWrappedMarks(escapedLine, "`", "<code>", "</code>")
	escapedLine = replaceMarkdownLinks(escapedLine)
	return escapedLine
}

func replaceMarkdownLinks(escapedLine string) string {
	var rebuilt strings.Builder
	cursor := 0
	for cursor < len(escapedLine) {
		openBracket := strings.Index(escapedLine[cursor:], "[")
		if openBracket < 0 {
			rebuilt.WriteString(escapedLine[cursor:])
			break
		}
		openBracket += cursor
		closeLabel := strings.Index(escapedLine[openBracket:], "](")
		if closeLabel < 0 {
			rebuilt.WriteString(escapedLine[cursor:])
			break
		}
		closeLabel += openBracket
		closeURL := strings.Index(escapedLine[closeLabel+2:], ")")
		if closeURL < 0 {
			rebuilt.WriteString(escapedLine[cursor:])
			break
		}
		closeURL += closeLabel + 2
		label := escapedLine[openBracket+1 : closeLabel]
		url := escapedLine[closeLabel+2 : closeURL]
		if !httpURL(url) {
			rebuilt.WriteString(escapedLine[cursor : closeURL+1])
			cursor = closeURL + 1
			continue
		}
		rebuilt.WriteString(escapedLine[cursor:openBracket])
		rebuilt.WriteString(`<a href="`)
		rebuilt.WriteString(url)
		rebuilt.WriteString(`">`)
		rebuilt.WriteString(label)
		rebuilt.WriteString(`</a>`)
		cursor = closeURL + 1
	}
	return rebuilt.String()
}

func httpURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")
}

func replaceWrappedMarks(
	sourceText string,
	markText string,
	openingHtmlTag string,
	closingHtmlTag string,
) string {
	pieces := strings.Split(sourceText, markText)
	if len(pieces) < 3 {
		return sourceText
	}

	var rebuilt strings.Builder
	for pieceIndex, piece := range pieces {
		if pieceIndex%2 == 1 && pieceIndex < len(pieces)-1 {
			rebuilt.WriteString(openingHtmlTag)
			rebuilt.WriteString(piece)
			rebuilt.WriteString(closingHtmlTag)
			continue
		}

		rebuilt.WriteString(piece)
	}

	return rebuilt.String()
}
