package article

import (
	"testing"
	"time"
)

func TestNewDraftArticleKeepsCleanWordsAndStartsHiddenFromThePublic(t *testing.T) {
	writtenAt := time.Date(2026, time.September, 17, 10, 0, 0, 0, time.UTC)

	builtArticle, buildError := NewDraftArticle(
		" Hello Goholic World ",
		"  Shipping a tiny Go blog  ",
		"A short note about clean packages.",
		"This is the body of the article.",
		writtenAt,
	)
	if buildError != nil {
		t.Fatalf("expected a valid draft, got error: %v", buildError)
	}

	if builtArticle.PublicWebAddressSlug != "hello-goholic-world" {
		t.Fatalf("slug = %q, want hello-goholic-world", builtArticle.PublicWebAddressSlug)
	}

	if builtArticle.ArticleTitle != "Shipping a tiny Go blog" {
		t.Fatalf("title = %q", builtArticle.ArticleTitle)
	}

	if builtArticle.IsVisibleToThePublic {
		t.Fatal("a brand new draft must stay hidden from the public")
	}
}

func TestNewDraftArticleRejectsAnEmptyTitle(t *testing.T) {
	_, buildError := NewDraftArticle(
		"missing-title",
		"   ",
		"summary",
		"body",
		time.Now(),
	)
	if buildError == nil {
		t.Fatal("an empty title should be rejected")
	}
}

func TestCleanPublicWebAddressSlugThrowsAwayStrangeMarks(t *testing.T) {
	cleanedSlug, cleanError := CleanPublicWebAddressSlug("Go & Clean-Architecture!!!")
	if cleanError != nil {
		t.Fatalf("unexpected error: %v", cleanError)
	}

	if cleanedSlug != "go-clean-architecture" {
		t.Fatalf("slug = %q, want go-clean-architecture", cleanedSlug)
	}
}

func TestShownToThePublicFlipsTheVisibilityFlag(t *testing.T) {
	writtenAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	draftArticle, _ := NewDraftArticle("slug", "Title", "Summary", "Body", writtenAt)

	publishedArticle := draftArticle.ShownToThePublic(writtenAt.Add(time.Hour))
	if !publishedArticle.IsVisibleToThePublic {
		t.Fatal("published article should be visible to the public")
	}
}
