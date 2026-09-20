package webpage

import "testing"

func TestTurnMarkdownIntoSafeHtmlEscapesDangerousTagsAndKeepsHeadings(t *testing.T) {
	htmlResult := TurnMarkdownIntoSafeHtml("# Hello\n\nPlease do not run <script>alert(1)</script>\n\n- one\n- two")

	if !containsAll(htmlResult, "<h1>Hello</h1>", "<ul>", "<li>one</li>", "&lt;script&gt;") {
		t.Fatalf("unexpected html: %s", htmlResult)
	}

	if containsAll(htmlResult, "<script>alert(1)</script>") {
		t.Fatal("raw script tags must never survive")
	}
}

func containsAll(sourceText string, pieces ...string) bool {
	for _, piece := range pieces {
		if !containsString(sourceText, piece) {
			return false
		}
	}

	return true
}

func containsString(sourceText string, piece string) bool {
	return len(sourceText) >= len(piece) && (sourceText == piece || len(piece) == 0 ||
		(func() bool {
			for index := 0; index+len(piece) <= len(sourceText); index++ {
				if sourceText[index:index+len(piece)] == piece {
					return true
				}
			}
			return false
		})())
}
