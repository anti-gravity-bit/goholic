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
	return escapedLine
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
