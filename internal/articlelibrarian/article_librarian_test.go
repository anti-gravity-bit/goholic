package articlelibrarian

import (
	"testing"
	"time"

	"github.com/anti-gravity-bit/goholic/internal/memoryarticleshelf"
	"github.com/anti-gravity-bit/goholic/internal/wallclock"
)

func TestArticleLibrarianHidesANewDraftFromReadersUntilTheWriterPublishesIt(t *testing.T) {
	frozenClock := wallclock.FrozenWallClock{
		FrozenMoment: time.Date(2026, time.September, 17, 12, 0, 0, 0, time.UTC),
	}
	articleLibrarian := NewArticleLibrarian(memoryarticleshelf.NewMemoryArticleShelf(), frozenClock)

	savedDraft, writeError := articleLibrarian.WriteNewDraftArticle(
		"clean-go-packages",
		"Clean Go packages",
		"How a small blog keeps its rooms tidy.",
		"Packages should tell a story.",
	)
	if writeError != nil {
		t.Fatalf("could not write draft: %v", writeError)
	}

	if savedDraft.IsVisibleToThePublic {
		t.Fatal("a new draft must stay hidden")
	}

	_, readerError := articleLibrarian.FindPublishedArticleForAReader("clean-go-packages")
	if readerError == nil {
		t.Fatal("a reader must not see a hidden draft")
	}

	publishedArticle, publishError := articleLibrarian.ShowArticleToThePublic("clean-go-packages")
	if publishError != nil {
		t.Fatalf("could not publish: %v", publishError)
	}

	if !publishedArticle.IsVisibleToThePublic {
		t.Fatal("after publishing, readers should see the article")
	}

	foundArticle, findError := articleLibrarian.FindPublishedArticleForAReader("clean-go-packages")
	if findError != nil {
		t.Fatalf("reader should now find the article: %v", findError)
	}

	if foundArticle.ArticleTitle != "Clean Go packages" {
		t.Fatalf("title = %q", foundArticle.ArticleTitle)
	}
}

func TestArticleLibrarianSearchOnlyLooksInsidePublishedStories(t *testing.T) {
	articleLibrarian := NewArticleLibrarian(
		memoryarticleshelf.NewMemoryArticleShelf(),
		wallclock.RealWallClock{},
	)

	_, _ = articleLibrarian.WriteNewDraftArticle(
		"hidden-draft",
		"Secret Kubernetes notes",
		"This draft talks about Kubernetes.",
		"Draft body.",
	)
	_, _ = articleLibrarian.WriteNewDraftArticle(
		"public-go",
		"Go services at work",
		"Latency, tests, and small interfaces.",
		"Published body.",
	)
	_, _ = articleLibrarian.ShowArticleToThePublic("public-go")

	matchingArticles, searchError := articleLibrarian.SearchPublishedArticlesForReaders("kubernetes")
	if searchError != nil {
		t.Fatalf("search failed: %v", searchError)
	}

	if len(matchingArticles) != 0 {
		t.Fatalf("drafts must not appear in public search, got %d", len(matchingArticles))
	}

	matchingGoArticles, goSearchError := articleLibrarian.SearchPublishedArticlesForReaders("latency")
	if goSearchError != nil {
		t.Fatalf("search failed: %v", goSearchError)
	}

	if len(matchingGoArticles) != 1 {
		t.Fatalf("expected one published match, got %d", len(matchingGoArticles))
	}
}

func TestArticleLibrarianCanRewriteAndThenThrowAnArticleAway(t *testing.T) {
	articleLibrarian := NewArticleLibrarian(
		memoryarticleshelf.NewMemoryArticleShelf(),
		wallclock.RealWallClock{},
	)

	_, writeError := articleLibrarian.WriteNewDraftArticle(
		"rewrite-me",
		"Old title",
		"Old summary",
		"Old body",
	)
	if writeError != nil {
		t.Fatalf("write failed: %v", writeError)
	}

	rewrittenArticle, rewriteError := articleLibrarian.RewriteExistingArticle(
		"rewrite-me",
		"New title",
		"New summary",
		"New body",
	)
	if rewriteError != nil {
		t.Fatalf("rewrite failed: %v", rewriteError)
	}

	if rewrittenArticle.ArticleTitle != "New title" {
		t.Fatalf("title = %q", rewrittenArticle.ArticleTitle)
	}

	throwAwayError := articleLibrarian.ThrowArticleAway("rewrite-me")
	if throwAwayError != nil {
		t.Fatalf("throw away failed: %v", throwAwayError)
	}

	_, findError := articleLibrarian.FindAnyArticleForTheWriter("rewrite-me")
	if findError == nil {
		t.Fatal("the thrown away article should be gone")
	}
}
