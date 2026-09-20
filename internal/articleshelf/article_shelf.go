// Package articleshelf describes the cabinet where articles live.
//
// This package has no wooden cabinet and no database.
// It only says what a cabinet must be able to do.
//
// Memory cabinets and sqlite cabinets both follow this list of promises.
package articleshelf

import "github.com/anti-gravity-bit/goholic/internal/article"

// ArticleShelf is the promise every storage cabinet must keep.
type ArticleShelf interface {
	PutArticleOnTheShelf(articleToKeep article.Article) (article.Article, error)
	ReplaceArticleOnTheShelf(articleToKeep article.Article) (article.Article, error)
	FindArticleByPublicWebAddressSlug(publicWebAddressSlug string) (article.Article, error)
	FindArticleByArticleNumber(articleNumber int64) (article.Article, error)
	ListArticlesVisibleToThePublic() ([]article.Article, error)
	ListEveryArticleForTheWriter() ([]article.Article, error)
	RemoveArticleByPublicWebAddressSlug(publicWebAddressSlug string) error
}

// ErrArticleWasNotFound means the cabinet looked and found nothing.
type ErrArticleWasNotFound struct {
	PublicWebAddressSlug string
}

// Error explains that the article is missing.
func (articleWasNotFound ErrArticleWasNotFound) Error() string {
	if articleWasNotFound.PublicWebAddressSlug == "" {
		return "the article was not found on the shelf"
	}

	return "the article was not found on the shelf: " + articleWasNotFound.PublicWebAddressSlug
}

// ErrArticleSlugAlreadyTaken means two stories tried to use the same web name.
type ErrArticleSlugAlreadyTaken struct {
	PublicWebAddressSlug string
}

// Error explains that the slug is busy.
func (articleSlugAlreadyTaken ErrArticleSlugAlreadyTaken) Error() string {
	return "another article already uses this public web address slug: " + articleSlugAlreadyTaken.PublicWebAddressSlug
}
