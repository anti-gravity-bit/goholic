package webpage

const sharedStylesheet = `
:root {
  --ink: #102017;
  --paper: #f4efe4;
  --card: #fffaf1;
  --moss: #1f6b4a;
  --moss-dark: #154a33;
  --clay: #c45c26;
  --line: #d7ccb8;
  --mute: #5c6b62;
  --danger: #9b2c2c;
  --ok: #1b5e3b;
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
.search-row { display: flex; gap: 8px; margin: 18px 0 28px; }
input, textarea, button {
  font: inherit;
}
input, textarea {
  width: 100%;
  border: 1px solid var(--line);
  background: #fff;
  padding: 10px 12px;
  border-radius: 10px;
}
textarea { min-height: 280px; }
button, .button-link {
  background: var(--moss);
  color: white;
  border: 0;
  border-radius: 999px;
  padding: 10px 16px;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
}
.button-quiet { background: transparent; color: var(--ink); border: 1px solid var(--line); }
.button-danger { background: var(--danger); }
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
.article-body h1, .article-body h2, .article-body h3 { line-height: 1.2; }
.article-body code {
  background: #efe6d4;
  padding: 0 5px;
  border-radius: 6px;
}
.flash-ok { color: var(--ok); }
.flash-bad { color: var(--danger); }
.writer-table { width: 100%; border-collapse: collapse; }
.writer-table th, .writer-table td {
  text-align: left; padding: 10px 8px; border-bottom: 1px solid var(--line); vertical-align: top;
}
.form-grid { display: grid; gap: 12px; }
.row-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.status-pill {
  display: inline-block;
  border-radius: 999px;
  padding: 2px 10px;
  font-size: 12px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.status-live { background: #d9f0e2; color: var(--ok); }
.status-draft { background: #f3e4d4; color: var(--clay); }
@media (max-width: 700px) {
  .hero h1 { font-size: 32px; }
  .site-header { flex-direction: column; }
}
`

const htmlPageStart = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.PageTitle}} · {{.WebsiteIdentity.WebsiteDisplayName}}</title>
  <style>` + sharedStylesheet + `</style>
  <script src="https://unpkg.com/htmx.org@2.0.4" defer></script>
</head>
<body>
  <div class="site-shell">
    <header class="site-header">
      <a class="site-wordmark" href="/">{{.WebsiteIdentity.WebsiteDisplayName}}</a>
      <nav class="site-nav">
        <a href="/">Stories</a>
        <a href="/writer/dashboard">Writer desk</a>
      </nav>
    </header>
`

const htmlPageEnd = `
    <footer class="site-footer" style="margin-top:48px;color:var(--mute);font-size:14px;">
      <div>{{.WebsiteIdentity.AuthorFullName}} · {{.WebsiteIdentity.AuthorJobTitle}}</div>
      <div>{{.WebsiteIdentity.AuthorHomeCity}}</div>
    </footer>
  </div>
</body>
</html>
`

const landingPageTemplate = htmlPageStart + `
    <section class="hero">
      <div class="eyebrow">{{.WebsiteIdentity.WebsiteHomeAddress}}</div>
      <h1>{{.WebsiteIdentity.AuthorFullName}}</h1>
      <p>{{.WebsiteIdentity.ShortWelcomeLine}}</p>
    </section>
    <form class="search-row"
          hx-get="/"
          hx-target="#article-card-list"
          hx-select="#article-card-list"
          hx-push-url="true">
      <input type="search" name="q" value="{{.SearchWords}}" placeholder="Search published stories">
      <button type="submit">Look</button>
    </form>
    <div id="article-card-list">
      {{if .PublishedArticles}}
        {{range .PublishedArticles}}
          <article class="article-card">
            <div class="eyebrow">{{.MomentTheArticleWasWritten.Format "2 January 2006"}}</div>
            <h2><a href="/article/{{.PublicWebAddressSlug}}">{{.ArticleTitle}}</a></h2>
            <p>{{.ArticleSummary}}</p>
          </article>
        {{end}}
      {{else}}
        <div class="panel">No published stories match those words yet.</div>
      {{end}}
    </div>
` + htmlPageEnd

const articlePageTemplate = htmlPageStart + `
    <article class="hero">
      <div class="eyebrow">{{.OpenedArticle.MomentTheArticleWasWritten.Format "2 January 2006"}}</div>
      <h1>{{.OpenedArticle.ArticleTitle}}</h1>
      <p>{{.OpenedArticle.ArticleSummary}}</p>
    </article>
    <div class="article-body">
      {{.ArticleBodyHtml}}
    </div>
` + htmlPageEnd

const writerLoginPageTemplate = htmlPageStart + `
    <section class="hero">
      <div class="eyebrow">Private</div>
      <h1>Writer desk</h1>
      <p>Only the owner of goholic.in should open this door.</p>
    </section>
    <form class="panel form-grid" method="post" action="/writer/login">
      {{if .FlashMessage}}<div class="flash-bad">{{.FlashMessage}}</div>{{end}}
      <label>Username
        <input name="writer_username" autocomplete="username" required>
      </label>
      <label>Password
        <input name="writer_password" type="password" autocomplete="current-password" required>
      </label>
      <button type="submit">Open the desk</button>
    </form>
` + htmlPageEnd

const writerDashboardTemplate = htmlPageStart + `
    <section class="hero">
      <div class="eyebrow">Private</div>
      <h1>Writer desk</h1>
      <p>Write a draft. Publish it when it is ready. Hide it if you change your mind.</p>
    </section>
    {{if .FlashMessage}}<div class="panel flash-ok">{{.FlashMessage}}</div>{{end}}
    <div class="panel">
      <h2>{{if .OpenedArticle.PublicWebAddressSlug}}Rewrite this story{{else}}Write a new draft{{end}}</h2>
      <form class="form-grid"
            method="post"
            action="{{if .OpenedArticle.PublicWebAddressSlug}}/writer/article/{{.OpenedArticle.PublicWebAddressSlug}}{{else}}/writer/article{{end}}"
            hx-post="{{if .OpenedArticle.PublicWebAddressSlug}}/writer/article/{{.OpenedArticle.PublicWebAddressSlug}}{{else}}/writer/article{{end}}"
            hx-target="body">
        <label>Web address slug
          <input name="public_web_address_slug" value="{{.OpenedArticle.PublicWebAddressSlug}}" {{if .OpenedArticle.PublicWebAddressSlug}}readonly{{end}} required>
        </label>
        <label>Title
          <input name="article_title" value="{{.OpenedArticle.ArticleTitle}}" required>
        </label>
        <label>Summary
          <input name="article_summary" value="{{.OpenedArticle.ArticleSummary}}" required>
        </label>
        <label>Body in simple markdown
          <textarea name="article_body_markdown" required>{{.OpenedArticle.ArticleBodyMarkdown}}</textarea>
        </label>
        <div class="row-actions">
          <button type="submit">Save draft</button>
          <a class="button-link button-quiet" href="/writer/dashboard">Clear form</a>
        </div>
      </form>
    </div>
    <div class="panel" id="writer-article-table">
      <h2>Every story on the shelf</h2>
      <table class="writer-table">
        <tr>
          <th>Title</th>
          <th>State</th>
          <th>Actions</th>
        </tr>
        {{range .EveryArticleForTheWriter}}
          <tr>
            <td>
              <strong>{{.ArticleTitle}}</strong><br>
              <span class="meta">/article/{{.PublicWebAddressSlug}}</span>
            </td>
            <td>
              {{if .IsVisibleToThePublic}}
                <span class="status-pill status-live">Live</span>
              {{else}}
                <span class="status-pill status-draft">Draft</span>
              {{end}}
            </td>
            <td class="row-actions">
              <a class="button-link button-quiet" href="/writer/dashboard?edit={{.PublicWebAddressSlug}}">Edit</a>
              {{if .IsVisibleToThePublic}}
                <button class="button-quiet"
                        hx-post="/writer/article/{{.PublicWebAddressSlug}}/hide"
                        hx-target="body">Hide</button>
              {{else}}
                <button
                        hx-post="/writer/article/{{.PublicWebAddressSlug}}/publish"
                        hx-target="body">Publish</button>
              {{end}}
              <button class="button-danger"
                      hx-delete="/writer/article/{{.PublicWebAddressSlug}}"
                      hx-confirm="Throw this story away?"
                      hx-target="body">Delete</button>
            </td>
          </tr>
        {{end}}
      </table>
      <form method="post" action="/writer/logout" style="margin-top:18px;">
        <button class="button-quiet" type="submit">Lock the desk</button>
      </form>
    </div>
` + htmlPageEnd
