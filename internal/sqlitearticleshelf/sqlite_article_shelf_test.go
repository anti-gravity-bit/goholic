package sqlitearticleshelf

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/anti-gravity-bit/goholic/internal/article"
)

func TestSqliteArticleShelfRemembersAnArticleAfterItIsSaved(t *testing.T) {
	sqliteFilePath := filepath.Join(t.TempDir(), "goholic-test.sqlite")
	sqliteArticleShelf, openError := OpenSqliteArticleShelfFile(sqliteFilePath)
	if openError != nil {
		t.Fatalf("could not open sqlite: %v", openError)
	}
	defer sqliteArticleShelf.CloseTheCabinet()

	writtenAt := time.Date(2026, time.March, 1, 8, 0, 0, 0, time.UTC)
	draftArticle, _ := article.NewDraftArticle(
		"sqlite-keeps-stories",
		"Sqlite keeps stories",
		"A file can remember a blog post.",
		"The body lives on disk.",
		writtenAt,
	)

	savedArticle, saveError := sqliteArticleShelf.PutArticleOnTheShelf(draftArticle)
	if saveError != nil {
		t.Fatalf("save failed: %v", saveError)
	}

	if savedArticle.ArticleNumber == 0 {
		t.Fatal("sqlite should give the article a number")
	}

	foundArticle, findError := sqliteArticleShelf.FindArticleByPublicWebAddressSlug("sqlite-keeps-stories")
	if findError != nil {
		t.Fatalf("find failed: %v", findError)
	}

	if foundArticle.ArticleTitle != "Sqlite keeps stories" {
		t.Fatalf("title = %q", foundArticle.ArticleTitle)
	}
}

func TestSqliteArticleShelfRefusesASecondArticleWithTheSameSlug(t *testing.T) {
	sqliteFilePath := filepath.Join(t.TempDir(), "goholic-test.sqlite")
	sqliteArticleShelf, _ := OpenSqliteArticleShelfFile(sqliteFilePath)
	defer sqliteArticleShelf.CloseTheCabinet()

	firstArticle, _ := article.NewDraftArticle("same-slug", "One", "Summary one", "Body one", time.Now())
	secondArticle, _ := article.NewDraftArticle("same-slug", "Two", "Summary two", "Body two", time.Now())

	_, firstSaveError := sqliteArticleShelf.PutArticleOnTheShelf(firstArticle)
	if firstSaveError != nil {
		t.Fatalf("first save failed: %v", firstSaveError)
	}

	_, secondSaveError := sqliteArticleShelf.PutArticleOnTheShelf(secondArticle)
	if secondSaveError == nil {
		t.Fatal("the second article with the same slug should be refused")
	}
}
