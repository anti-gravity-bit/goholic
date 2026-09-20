// Command goholic-static turns content/articles markdown into public/ HTML.
// Vercel serves public/. The writer desk is not exported.
package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/anti-gravity-bit/goholic/internal/webpage"
	"github.com/anti-gravity-bit/goholic/internal/websiteidentity"
)

type articleFile struct {
	Title   string
	Slug    string
	Summary string
	Date    time.Time
	Project string
	Kind    string
	BodyMD  string
	BodyHTML template.HTML
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	articles, err := loadArticles(filepath.Join(root, "content", "articles"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "goholic-static: %v\n", err)
		os.Exit(1)
	}
	publicDir := filepath.Join(root, "public")
	if err := os.RemoveAll(publicDir); err != nil {
		fmt.Fprintf(os.Stderr, "goholic-static: %v\n", err)
		os.Exit(1)
	}
	identity := websiteidentity.NewGoholicWebsiteIdentity()
	if err := writeLanding(publicDir, identity, articles); err != nil {
		fmt.Fprintf(os.Stderr, "goholic-static: %v\n", err)
		os.Exit(1)
	}
	for _, article := range articles {
		if err := writeArticle(publicDir, identity, article); err != nil {
			fmt.Fprintf(os.Stderr, "goholic-static: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("goholic-static: wrote %d articles to %s\n", len(articles), publicDir)
}

func loadArticles(dir string) ([]articleFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var articles []articleFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		article, parseErr := parseFrontMatter(string(raw))
		if parseErr != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), parseErr)
		}
		article.BodyHTML = template.HTML(webpage.TurnMarkdownIntoSafeHtml(article.BodyMD))
		articles = append(articles, article)
	}
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})
	return articles, nil
}

func parseFrontMatter(raw string) (articleFile, error) {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(raw, "---\n") {
		return articleFile{}, fmt.Errorf("missing front matter")
	}
	rest := raw[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return articleFile{}, fmt.Errorf("unterminated front matter")
	}
	meta := rest[:end]
	body := strings.TrimSpace(rest[end+5:])
	article := articleFile{BodyMD: body}
	for _, line := range strings.Split(meta, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "title":
			article.Title = value
		case "slug":
			article.Slug = value
		case "summary":
			article.Summary = value
		case "project":
			article.Project = value
		case "kind":
			article.Kind = value
		case "date":
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				return articleFile{}, err
			}
			article.Date = parsed
		}
	}
	if article.Title == "" || article.Slug == "" {
		return articleFile{}, fmt.Errorf("title and slug are required")
	}
	return article, nil
}

const stylesheet = `
:root {
  --ink: #102017;
  --paper: #f4efe4;
  --card: #fffaf1;
  --moss: #1f6b4a;
  --moss-dark: #154a33;
  --line: #d7ccb8;
  --mute: #5c6b62;
}
* { box-sizing: border-box; }
html, body { margin: 0; padding: 0; }
body {
  font-family: "Iowan Old Style", "Palatino Linotype", Palatino, Georgia, serif;
  background: radial-gradient(1200px 500px at 10% -10%, #e7f3ea 0%, transparent 55%), var(--paper);
  color: var(--ink);
  line-height: 1.55;
}
a { color: var(--moss-dark); }
.site-shell { max-width: 920px; margin: 0 auto; padding: 28px 20px 80px; }
.site-header, .site-footer {
  display: flex; justify-content: space-between; align-items: baseline; gap: 16px;
}
.site-wordmark {
  font-size: 28px; letter-spacing: 0.08em; text-transform: uppercase; text-decoration: none; color: var(--ink);
}
.site-nav { display: flex; gap: 16px; font-size: 15px; }
.eyebrow { color: var(--moss); letter-spacing: 0.12em; text-transform: uppercase; font-size: 12px; }
.hero { margin: 36px 0 28px; }
.hero h1 { font-size: 42px; line-height: 1.15; margin: 8px 0; }
.hero p { font-size: 20px; color: var(--mute); max-width: 640px; }
.article-card, .panel {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: 18px;
  padding: 20px 22px;
  margin-bottom: 16px;
}
.article-card h2 { margin: 6px 0 8px; }
.meta { color: var(--mute); font-size: 14px; }
.article-body { font-size: 19px; }
.article-body code { background: #efe6d4; padding: 0 5px; border-radius: 6px; }
.article-body a { overflow-wrap: anywhere; }
@media (max-width: 700px) {
  .hero h1 { font-size: 32px; }
  .site-header { flex-direction: column; }
}
`

func writeLanding(publicDir string, identity websiteidentity.WebsiteIdentity, articles []articleFile) error {
	var cards bytes.Buffer
	for _, article := range articles {
		fmt.Fprintf(&cards, `<article class="article-card">
      <div class="eyebrow">%s · %s</div>
      <h2><a href="/article/%s/">%s</a></h2>
      <p>%s</p>
    </article>
`, html.EscapeString(article.Date.Format("2 January 2006")), html.EscapeString(article.Project),
			html.EscapeString(article.Slug), html.EscapeString(article.Title), html.EscapeString(article.Summary))
	}
	page := wrap(identity, identity.AuthorFullName+" · stories", fmt.Sprintf(`
    <section class="hero">
      <div class="eyebrow">%s</div>
      <h1>%s</h1>
      <p>%s</p>
    </section>
    <div id="article-card-list">
    %s
    </div>
`, html.EscapeString(identity.WebsiteHomeAddress), html.EscapeString(identity.AuthorFullName),
		html.EscapeString(identity.ShortWelcomeLine), cards.String()))
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(publicDir, "index.html"), []byte(page), 0o644)
}

func writeArticle(publicDir string, identity websiteidentity.WebsiteIdentity, article articleFile) error {
	dir := filepath.Join(publicDir, "article", article.Slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	page := wrap(identity, article.Title, fmt.Sprintf(`
    <article class="hero">
      <div class="eyebrow">%s · %s</div>
      <h1>%s</h1>
      <p>%s</p>
    </article>
    <div class="article-body">
      %s
    </div>
`, html.EscapeString(article.Date.Format("2 January 2006")), html.EscapeString(article.Project),
		html.EscapeString(article.Title), html.EscapeString(article.Summary), string(article.BodyHTML)))
	return os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644)
}

func wrap(identity websiteidentity.WebsiteIdentity, title, inner string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s · %s</title>
  <style>%s</style>
</head>
<body>
  <div class="site-shell">
    <header class="site-header">
      <a class="site-wordmark" href="/">%s</a>
      <nav class="site-nav">
        <a href="/">Stories</a>
        <a href="https://github.com/anti-gravity-bit">GitHub</a>
        <a href="%s">LinkedIn</a>
      </nav>
    </header>
    %s
    <footer class="site-footer" style="margin-top:48px;color:var(--mute);font-size:14px;">
      <div>%s · %s</div>
      <div>%s</div>
    </footer>
  </div>
</body>
</html>
`, html.EscapeString(title), html.EscapeString(identity.WebsiteDisplayName), stylesheet,
		html.EscapeString(identity.WebsiteDisplayName), html.EscapeString(identity.AuthorLinkedinPage),
		inner,
		html.EscapeString(identity.AuthorFullName), html.EscapeString(identity.AuthorJobTitle),
		html.EscapeString(identity.AuthorHomeCity))
}
