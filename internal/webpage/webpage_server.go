// Package webpage is the front desk of goholic.in.
//
// Readers ask this desk for the landing page and an article page.
// The writer asks this desk for the dashboard.
// The desk never opens the sqlite file itself. It asks the librarian.
package webpage

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/anti-gravity-bit/goholic/internal/article"
	"github.com/anti-gravity-bit/goholic/internal/articlelibrarian"
	"github.com/anti-gravity-bit/goholic/internal/websiteidentity"
	"github.com/anti-gravity-bit/goholic/internal/writerpassport"
)

// WebpageServer is the front desk that answers HTTP questions.
type WebpageServer struct {
	websiteIdentity  websiteidentity.WebsiteIdentity
	articleLibrarian *articlelibrarian.ArticleLibrarian
	writerPassport   *writerpassport.WriterPassport
	parsedTemplates  *template.Template
}

type htmlPageView struct {
	PageTitle                string
	WebsiteIdentity          websiteidentity.WebsiteIdentity
	SearchWords              string
	PublishedArticles        []article.Article
	EveryArticleForTheWriter []article.Article
	OpenedArticle            article.Article
	ArticleBodyHtml          template.HTML
	FlashMessage             string
}

// NewWebpageServer builds the front desk and parses HTML once.
func NewWebpageServer(
	websiteIdentity websiteidentity.WebsiteIdentity,
	articleLibrarian *articlelibrarian.ArticleLibrarian,
	writerPassport *writerpassport.WriterPassport,
) (*WebpageServer, error) {
	parsedTemplates := template.New("goholic")
	_, parseError := parsedTemplates.New("landing").Parse(landingPageTemplate)
	if parseError != nil {
		return nil, parseError
	}

	_, parseError = parsedTemplates.New("article").Parse(articlePageTemplate)
	if parseError != nil {
		return nil, parseError
	}

	_, parseError = parsedTemplates.New("login").Parse(writerLoginPageTemplate)
	if parseError != nil {
		return nil, parseError
	}

	_, parseError = parsedTemplates.New("dashboard").Parse(writerDashboardTemplate)
	if parseError != nil {
		return nil, parseError
	}

	return &WebpageServer{
		websiteIdentity:  websiteIdentity,
		articleLibrarian: articleLibrarian,
		writerPassport:   writerPassport,
		parsedTemplates:  parsedTemplates,
	}, nil
}

// BindRoutesToDefaultServeMux attaches every public and writer URL.
func (webpageServer *WebpageServer) BindRoutesToDefaultServeMux() {
	http.HandleFunc("/", webpageServer.ShowLandingPageOrNotFound)
	http.HandleFunc("/article/", webpageServer.ShowOnePublishedArticlePage)
	http.HandleFunc("/writer/login", webpageServer.ShowOrSubmitWriterLogin)
	http.HandleFunc("/writer/logout", webpageServer.LockTheWriterDesk)
	http.HandleFunc("/writer/dashboard", webpageServer.ShowWriterDashboard)
	http.HandleFunc("/writer/article/", webpageServer.HandleWriterArticleAction)
	http.HandleFunc("/writer/article", webpageServer.SaveBrandNewDraftArticle)
}

// ShowLandingPageOrNotFound shows the home page for "/" and 404 for unknown paths.
func (webpageServer *WebpageServer) ShowLandingPageOrNotFound(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	if incomingRequest.URL.Path != "/" {
		http.NotFound(responseWriter, incomingRequest)
		return
	}

	searchWords := strings.TrimSpace(incomingRequest.URL.Query().Get("q"))
	publishedArticles, listError := webpageServer.articleLibrarian.SearchPublishedArticlesForReaders(searchWords)
	if listError != nil {
		http.Error(responseWriter, "the shelf could not be read", http.StatusInternalServerError)
		return
	}

	webpageServer.writeHtmlTemplate(responseWriter, "landing", htmlPageView{
		PageTitle:         "Stories",
		WebsiteIdentity:   webpageServer.websiteIdentity,
		SearchWords:       searchWords,
		PublishedArticles: publishedArticles,
	})
}

// ShowOnePublishedArticlePage shows one live story.
func (webpageServer *WebpageServer) ShowOnePublishedArticlePage(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	publicWebAddressSlug := strings.TrimPrefix(incomingRequest.URL.Path, "/article/")
	openedArticle, findError := webpageServer.articleLibrarian.FindPublishedArticleForAReader(publicWebAddressSlug)
	if findError != nil {
		http.NotFound(responseWriter, incomingRequest)
		return
	}

	webpageServer.writeHtmlTemplate(responseWriter, "article", htmlPageView{
		PageTitle:       openedArticle.ArticleTitle,
		WebsiteIdentity: webpageServer.websiteIdentity,
		OpenedArticle:   openedArticle,
		ArticleBodyHtml: template.HTML(TurnMarkdownIntoSafeHtml(openedArticle.ArticleBodyMarkdown)),
	})
}

// ShowOrSubmitWriterLogin shows the keyhole or checks the key.
func (webpageServer *WebpageServer) ShowOrSubmitWriterLogin(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	if incomingRequest.Method == http.MethodGet {
		webpageServer.writeHtmlTemplate(responseWriter, "login", htmlPageView{
			PageTitle:       "Writer login",
			WebsiteIdentity: webpageServer.websiteIdentity,
		})
		return
	}

	if incomingRequest.Method != http.MethodPost {
		http.Error(responseWriter, "that method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	parseError := incomingRequest.ParseForm()
	if parseError != nil {
		http.Error(responseWriter, "the form could not be read", http.StatusBadRequest)
		return
	}

	signedSessionToken, openError := webpageServer.writerPassport.TryToOpenTheWriterDoor(
		incomingRequest.FormValue("writer_username"),
		incomingRequest.FormValue("writer_password"),
	)
	if openError != nil {
		responseWriter.WriteHeader(http.StatusUnauthorized)
		webpageServer.writeHtmlTemplate(responseWriter, "login", htmlPageView{
			PageTitle:       "Writer login",
			WebsiteIdentity: webpageServer.websiteIdentity,
			FlashMessage:    "That username or password is wrong.",
		})
		return
	}

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     writerpassport.WriterSessionCookieName(),
		Value:    signedSessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   incomingRequest.TLS != nil || incomingRequest.Header.Get("X-Forwarded-Proto") == "https",
	})
	http.Redirect(responseWriter, incomingRequest, "/writer/dashboard", http.StatusSeeOther)
}

// LockTheWriterDesk throws the cookie away.
func (webpageServer *WebpageServer) LockTheWriterDesk(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     writerpassport.WriterSessionCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(responseWriter, incomingRequest, "/", http.StatusSeeOther)
}

// ShowWriterDashboard shows every story and the writing form.
func (webpageServer *WebpageServer) ShowWriterDashboard(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	if !webpageServer.theWriterIsSignedIn(incomingRequest) {
		http.Redirect(responseWriter, incomingRequest, "/writer/login", http.StatusSeeOther)
		return
	}

	everyArticleForTheWriter, listError := webpageServer.articleLibrarian.ListEveryArticleForTheWriter()
	if listError != nil {
		http.Error(responseWriter, "the shelf could not be read", http.StatusInternalServerError)
		return
	}

	openedArticle := article.Article{}
	slugToEdit := incomingRequest.URL.Query().Get("edit")
	if slugToEdit != "" {
		foundArticle, findError := webpageServer.articleLibrarian.FindAnyArticleForTheWriter(slugToEdit)
		if findError == nil {
			openedArticle = foundArticle
		}
	}

	webpageServer.writeHtmlTemplate(responseWriter, "dashboard", htmlPageView{
		PageTitle:                "Writer desk",
		WebsiteIdentity:          webpageServer.websiteIdentity,
		EveryArticleForTheWriter: everyArticleForTheWriter,
		OpenedArticle:            openedArticle,
		FlashMessage:             incomingRequest.URL.Query().Get("flash"),
	})
}

// SaveBrandNewDraftArticle stores a new hidden story.
func (webpageServer *WebpageServer) SaveBrandNewDraftArticle(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	if !webpageServer.theWriterIsSignedIn(incomingRequest) {
		http.Redirect(responseWriter, incomingRequest, "/writer/login", http.StatusSeeOther)
		return
	}

	if incomingRequest.Method != http.MethodPost {
		http.Error(responseWriter, "that method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	parseError := incomingRequest.ParseForm()
	if parseError != nil {
		http.Error(responseWriter, "the form could not be read", http.StatusBadRequest)
		return
	}

	_, saveError := webpageServer.articleLibrarian.WriteNewDraftArticle(
		incomingRequest.FormValue("public_web_address_slug"),
		incomingRequest.FormValue("article_title"),
		incomingRequest.FormValue("article_summary"),
		incomingRequest.FormValue("article_body_markdown"),
	)
	if saveError != nil {
		http.Error(responseWriter, saveError.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(responseWriter, incomingRequest, "/writer/dashboard?flash=Draft+saved.", http.StatusSeeOther)
}

// HandleWriterArticleAction edits, publishes, hides, or deletes one story.
func (webpageServer *WebpageServer) HandleWriterArticleAction(
	responseWriter http.ResponseWriter,
	incomingRequest *http.Request,
) {
	if !webpageServer.theWriterIsSignedIn(incomingRequest) {
		http.Redirect(responseWriter, incomingRequest, "/writer/login", http.StatusSeeOther)
		return
	}

	pathAfterPrefix := strings.TrimPrefix(incomingRequest.URL.Path, "/writer/article/")
	pathPieces := strings.Split(pathAfterPrefix, "/")
	publicWebAddressSlug := pathPieces[0]
	actionName := ""
	if len(pathPieces) > 1 {
		actionName = pathPieces[1]
	}

	var actionError error
	switch {
	case incomingRequest.Method == http.MethodPost && actionName == "publish":
		_, actionError = webpageServer.articleLibrarian.ShowArticleToThePublic(publicWebAddressSlug)
	case incomingRequest.Method == http.MethodPost && actionName == "hide":
		_, actionError = webpageServer.articleLibrarian.HideArticleFromThePublic(publicWebAddressSlug)
	case incomingRequest.Method == http.MethodDelete && actionName == "":
		actionError = webpageServer.articleLibrarian.ThrowArticleAway(publicWebAddressSlug)
	case incomingRequest.Method == http.MethodPost && actionName == "":
		parseError := incomingRequest.ParseForm()
		if parseError != nil {
			http.Error(responseWriter, "the form could not be read", http.StatusBadRequest)
			return
		}
		_, actionError = webpageServer.articleLibrarian.RewriteExistingArticle(
			publicWebAddressSlug,
			incomingRequest.FormValue("article_title"),
			incomingRequest.FormValue("article_summary"),
			incomingRequest.FormValue("article_body_markdown"),
		)
	default:
		http.Error(responseWriter, "that writer action is unknown", http.StatusBadRequest)
		return
	}

	if actionError != nil {
		http.Error(responseWriter, actionError.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(responseWriter, incomingRequest, "/writer/dashboard?flash=Saved.", http.StatusSeeOther)
}

func (webpageServer *WebpageServer) theWriterIsSignedIn(incomingRequest *http.Request) bool {
	sessionCookie, cookieError := incomingRequest.Cookie(writerpassport.WriterSessionCookieName())
	if cookieError != nil {
		return false
	}

	return webpageServer.writerPassport.DoesThisSignedSessionStillCount(sessionCookie.Value)
}

func (webpageServer *WebpageServer) writeHtmlTemplate(
	responseWriter http.ResponseWriter,
	templateName string,
	htmlPageView htmlPageView,
) {
	responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	executeError := webpageServer.parsedTemplates.ExecuteTemplate(responseWriter, templateName, htmlPageView)
	if executeError != nil {
		http.Error(responseWriter, "the page could not be drawn", http.StatusInternalServerError)
	}
}
