// Package articlelibrarian is the librarian who follows the blog rules.
//
// The webpage talks to the librarian.
// The librarian talks to a cabinet (ArticleShelf).
// The librarian never talks to sqlite or memory by name.
// That is the Dependency Inversion idea: depend on a promise, not a brand of cabinet.
package articlelibrarian

import (
	"strings"

	"github.com/anti-gravity-bit/goholic/internal/article"
	"github.com/anti-gravity-bit/goholic/internal/articleshelf"
	"github.com/anti-gravity-bit/goholic/internal/wallclock"
)

// ArticleLibrarian knows how writers and readers may use articles.
type ArticleLibrarian struct {
	articleShelf articleshelf.ArticleShelf
	wallClock    wallclock.WallClock
}

// NewArticleLibrarian hires a librarian for one cabinet and one clock.
func NewArticleLibrarian(
	articleShelf articleshelf.ArticleShelf,
	wallClock wallclock.WallClock,
) *ArticleLibrarian {
	return &ArticleLibrarian{
		articleShelf: articleShelf,
		wallClock:    wallClock,
	}
}

// WriteNewDraftArticle puts a hidden story on the shelf.
func (articleLibrarian *ArticleLibrarian) WriteNewDraftArticle(
	publicWebAddressSlug string,
	articleTitle string,
	articleSummary string,
	articleBodyMarkdown string,
) (article.Article, error) {
	draftArticle, buildError := article.NewDraftArticle(
		publicWebAddressSlug,
		articleTitle,
		articleSummary,
		articleBodyMarkdown,
		articleLibrarian.wallClock.WhatTimeIsItRightNow(),
	)
	if buildError != nil {
		return article.Article{}, buildError
	}

	savedArticle, saveError := articleLibrarian.articleShelf.PutArticleOnTheShelf(draftArticle)
	if saveError != nil {
		return article.Article{}, saveError
	}

	return savedArticle, nil
}

// RewriteExistingArticle changes the words of a story that already exists.
func (articleLibrarian *ArticleLibrarian) RewriteExistingArticle(
	publicWebAddressSlug string,
	articleTitle string,
	articleSummary string,
	articleBodyMarkdown string,
) (article.Article, error) {
	existingArticle, findError := articleLibrarian.articleShelf.FindArticleByPublicWebAddressSlug(publicWebAddressSlug)
	if findError != nil {
		return article.Article{}, findError
	}

	rewrittenArticle, rewriteError := existingArticle.WithUpdatedWriting(
		articleTitle,
		articleSummary,
		articleBodyMarkdown,
		articleLibrarian.wallClock.WhatTimeIsItRightNow(),
	)
	if rewriteError != nil {
		return article.Article{}, rewriteError
	}

	savedArticle, saveError := articleLibrarian.articleShelf.ReplaceArticleOnTheShelf(rewrittenArticle)
	if saveError != nil {
		return article.Article{}, saveError
	}

	return savedArticle, nil
}

// ShowArticleToThePublic lets readers on the internet see the story.
func (articleLibrarian *ArticleLibrarian) ShowArticleToThePublic(
	publicWebAddressSlug string,
) (article.Article, error) {
	existingArticle, findError := articleLibrarian.articleShelf.FindArticleByPublicWebAddressSlug(publicWebAddressSlug)
	if findError != nil {
		return article.Article{}, findError
	}

	publishedArticle := existingArticle.ShownToThePublic(articleLibrarian.wallClock.WhatTimeIsItRightNow())
	savedArticle, saveError := articleLibrarian.articleShelf.ReplaceArticleOnTheShelf(publishedArticle)
	if saveError != nil {
		return article.Article{}, saveError
	}

	return savedArticle, nil
}

// HideArticleFromThePublic turns a published story back into a draft.
func (articleLibrarian *ArticleLibrarian) HideArticleFromThePublic(
	publicWebAddressSlug string,
) (article.Article, error) {
	existingArticle, findError := articleLibrarian.articleShelf.FindArticleByPublicWebAddressSlug(publicWebAddressSlug)
	if findError != nil {
		return article.Article{}, findError
	}

	hiddenArticle := existingArticle.HiddenFromThePublic(articleLibrarian.wallClock.WhatTimeIsItRightNow())
	savedArticle, saveError := articleLibrarian.articleShelf.ReplaceArticleOnTheShelf(hiddenArticle)
	if saveError != nil {
		return article.Article{}, saveError
	}

	return savedArticle, nil
}

// ThrowArticleAway removes a story from the cabinet.
func (articleLibrarian *ArticleLibrarian) ThrowArticleAway(
	publicWebAddressSlug string,
) error {
	return articleLibrarian.articleShelf.RemoveArticleByPublicWebAddressSlug(publicWebAddressSlug)
}

// FindPublishedArticleForAReader only returns stories the public may see.
func (articleLibrarian *ArticleLibrarian) FindPublishedArticleForAReader(
	publicWebAddressSlug string,
) (article.Article, error) {
	existingArticle, findError := articleLibrarian.articleShelf.FindArticleByPublicWebAddressSlug(publicWebAddressSlug)
	if findError != nil {
		return article.Article{}, findError
	}

	if !existingArticle.IsVisibleToThePublic {
		return article.Article{}, articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: publicWebAddressSlug,
		}
	}

	return existingArticle, nil
}

// FindAnyArticleForTheWriter returns drafts and published stories.
func (articleLibrarian *ArticleLibrarian) FindAnyArticleForTheWriter(
	publicWebAddressSlug string,
) (article.Article, error) {
	return articleLibrarian.articleShelf.FindArticleByPublicWebAddressSlug(publicWebAddressSlug)
}

// ListPublishedArticlesForReaders returns public stories, newest first.
func (articleLibrarian *ArticleLibrarian) ListPublishedArticlesForReaders() ([]article.Article, error) {
	return articleLibrarian.articleShelf.ListArticlesVisibleToThePublic()
}

// ListEveryArticleForTheWriter returns every story in the cabinet.
func (articleLibrarian *ArticleLibrarian) ListEveryArticleForTheWriter() ([]article.Article, error) {
	return articleLibrarian.articleShelf.ListEveryArticleForTheWriter()
}

// SearchPublishedArticlesForReaders keeps stories whose title or summary mention the search words.
func (articleLibrarian *ArticleLibrarian) SearchPublishedArticlesForReaders(
	searchWords string,
) ([]article.Article, error) {
	publishedArticles, listError := articleLibrarian.articleShelf.ListArticlesVisibleToThePublic()
	if listError != nil {
		return nil, listError
	}

	cleanedSearchWords := strings.ToLower(strings.TrimSpace(searchWords))
	if cleanedSearchWords == "" {
		return publishedArticles, nil
	}

	matchingArticles := make([]article.Article, 0)
	for _, publishedArticle := range publishedArticles {
		titleAndSummary := strings.ToLower(publishedArticle.ArticleTitle + " " + publishedArticle.ArticleSummary)
		if strings.Contains(titleAndSummary, cleanedSearchWords) {
			matchingArticles = append(matchingArticles, publishedArticle)
		}
	}

	return matchingArticles, nil
}
