package webpage

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/anti-gravity-bit/goholic/internal/articlelibrarian"
	"github.com/anti-gravity-bit/goholic/internal/memoryarticleshelf"
	"github.com/anti-gravity-bit/goholic/internal/wallclock"
	"github.com/anti-gravity-bit/goholic/internal/websiteidentity"
	"github.com/anti-gravity-bit/goholic/internal/writerpassport"
)

func TestLandingPageShowsOnlyPublishedArticlesAndHidesDrafts(t *testing.T) {
	testServer := startTestWebsite(t)
	defer testServer.Close()

	landingPageBody := readBodyFromGet(t, testServer.URL+"/")
	if !strings.Contains(landingPageBody, "Interfaces are enough") {
		t.Fatalf("published title missing: %s", landingPageBody)
	}

	if strings.Contains(landingPageBody, "Hidden kitchen notes") {
		t.Fatal("a draft must not appear on the landing page")
	}
}

func TestArticlePageRendersMarkdownAndRejectsADraftSlug(t *testing.T) {
	testServer := startTestWebsite(t)
	defer testServer.Close()

	articlePageBody := readBodyFromGet(t, testServer.URL+"/article/interfaces-are-enough")
	if !strings.Contains(articlePageBody, "<h2>A small rule</h2>") {
		t.Fatalf("markdown heading missing: %s", articlePageBody)
	}

	draftResponse, getError := http.Get(testServer.URL + "/article/hidden-kitchen-notes")
	if getError != nil {
		t.Fatalf("request failed: %v", getError)
	}
	defer draftResponse.Body.Close()

	if draftResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("draft slug status = %d, want 404", draftResponse.StatusCode)
	}
}

func TestWriterCanLogInPublishAndSeeTheStoryOnTheLandingPage(t *testing.T) {
	testServer := startTestWebsite(t)
	defer testServer.Close()

	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	loginResponse, loginError := httpClient.PostForm(testServer.URL+"/writer/login", url.Values{
		"writer_username": {"abir"},
		"writer_password": {"correct-horse-battery"},
	})
	if loginError != nil {
		t.Fatalf("login failed: %v", loginError)
	}
	defer loginResponse.Body.Close()

	publishResponse, publishError := httpClient.PostForm(testServer.URL+"/writer/article/hidden-kitchen-notes/publish", url.Values{})
	if publishError != nil {
		t.Fatalf("publish failed: %v", publishError)
	}
	defer publishResponse.Body.Close()

	landingPageBody := readBodyFromGet(t, testServer.URL+"/")
	if !strings.Contains(landingPageBody, "Hidden kitchen notes") {
		t.Fatalf("newly published title missing: %s", landingPageBody)
	}
}

func startTestWebsite(t *testing.T) *httptest.Server {
	t.Helper()

	articleLibrarian := articlelibrarian.NewArticleLibrarian(
		memoryarticleshelf.NewMemoryArticleShelf(),
		wallclock.FrozenWallClock{
			FrozenMoment: time.Date(2026, time.September, 17, 16, 0, 0, 0, time.UTC),
		},
	)
	_, _ = articleLibrarian.WriteNewDraftArticle(
		"interfaces-are-enough",
		"Interfaces are enough",
		"A cabinet promise is better than a brand name.",
		"## A small rule\n\nDepend on a promise, not on sqlite by name.",
	)
	_, _ = articleLibrarian.ShowArticleToThePublic("interfaces-are-enough")
	_, _ = articleLibrarian.WriteNewDraftArticle(
		"hidden-kitchen-notes",
		"Hidden kitchen notes",
		"This draft stays in the kitchen.",
		"Nobody outside should see this yet.",
	)

	writerPassport, passportError := writerpassport.NewWriterPassport(
		"abir",
		"correct-horse-battery",
		"sixteen-or-more-chars",
		wallclock.RealWallClock{},
	)
	if passportError != nil {
		t.Fatalf("passport failed: %v", passportError)
	}

	webpageServer, serverError := NewWebpageServer(
		websiteidentity.NewGoholicWebsiteIdentity(),
		articleLibrarian,
		writerPassport,
	)
	if serverError != nil {
		t.Fatalf("webpage server failed: %v", serverError)
	}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", webpageServer.ShowLandingPageOrNotFound)
	serveMux.HandleFunc("/article/", webpageServer.ShowOnePublishedArticlePage)
	serveMux.HandleFunc("/writer/login", webpageServer.ShowOrSubmitWriterLogin)
	serveMux.HandleFunc("/writer/dashboard", webpageServer.ShowWriterDashboard)
	serveMux.HandleFunc("/writer/article/", webpageServer.HandleWriterArticleAction)
	serveMux.HandleFunc("/writer/article", webpageServer.SaveBrandNewDraftArticle)

	return httptest.NewServer(serveMux)
}

func readBodyFromGet(t *testing.T, pageAddress string) string {
	t.Helper()

	response, getError := http.Get(pageAddress)
	if getError != nil {
		t.Fatalf("GET %s failed: %v", pageAddress, getError)
	}
	defer response.Body.Close()

	bodyBytes, readError := io.ReadAll(response.Body)
	if readError != nil {
		t.Fatalf("read failed: %v", readError)
	}

	return string(bodyBytes)
}
