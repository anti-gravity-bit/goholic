// Package memoryarticleshelf is a cabinet that lives only in computer memory.
//
// Tests love this cabinet because it is fast and forgets everything when the test ends.
package memoryarticleshelf

import (
	"sync"

	"github.com/anti-gravity-bit/goholic/internal/article"
	"github.com/anti-gravity-bit/goholic/internal/articleshelf"
)

// MemoryArticleShelf keeps articles in a map.
type MemoryArticleShelf struct {
	guardThatStopsTwoHandsAtOnce   sync.Mutex
	nextArticleNumber              int64
	articlesByPublicWebAddressSlug map[string]article.Article
}

// NewMemoryArticleShelf builds an empty memory cabinet.
func NewMemoryArticleShelf() *MemoryArticleShelf {
	return &MemoryArticleShelf{
		nextArticleNumber:              1,
		articlesByPublicWebAddressSlug: map[string]article.Article{},
	}
}

// PutArticleOnTheShelf stores a new article.
func (memoryArticleShelf *MemoryArticleShelf) PutArticleOnTheShelf(
	articleToKeep article.Article,
) (article.Article, error) {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	if _, articleAlreadyExists := memoryArticleShelf.articlesByPublicWebAddressSlug[articleToKeep.PublicWebAddressSlug]; articleAlreadyExists {
		return article.Article{}, articleshelf.ErrArticleSlugAlreadyTaken{
			PublicWebAddressSlug: articleToKeep.PublicWebAddressSlug,
		}
	}

	articleToKeep.ArticleNumber = memoryArticleShelf.nextArticleNumber
	memoryArticleShelf.nextArticleNumber++
	memoryArticleShelf.articlesByPublicWebAddressSlug[articleToKeep.PublicWebAddressSlug] = articleToKeep
	return articleToKeep, nil
}

// ReplaceArticleOnTheShelf overwrites an article that already has a number.
func (memoryArticleShelf *MemoryArticleShelf) ReplaceArticleOnTheShelf(
	articleToKeep article.Article,
) (article.Article, error) {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	_, articleAlreadyExists := memoryArticleShelf.articlesByPublicWebAddressSlug[articleToKeep.PublicWebAddressSlug]
	if !articleAlreadyExists {
		return article.Article{}, articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: articleToKeep.PublicWebAddressSlug,
		}
	}

	memoryArticleShelf.articlesByPublicWebAddressSlug[articleToKeep.PublicWebAddressSlug] = articleToKeep
	return articleToKeep, nil
}

// FindArticleByPublicWebAddressSlug looks up one article by its web name.
func (memoryArticleShelf *MemoryArticleShelf) FindArticleByPublicWebAddressSlug(
	publicWebAddressSlug string,
) (article.Article, error) {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	foundArticle, articleAlreadyExists := memoryArticleShelf.articlesByPublicWebAddressSlug[publicWebAddressSlug]
	if !articleAlreadyExists {
		return article.Article{}, articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: publicWebAddressSlug,
		}
	}

	return foundArticle, nil
}

// FindArticleByArticleNumber looks up one article by its number.
func (memoryArticleShelf *MemoryArticleShelf) FindArticleByArticleNumber(
	articleNumber int64,
) (article.Article, error) {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	for _, foundArticle := range memoryArticleShelf.articlesByPublicWebAddressSlug {
		if foundArticle.ArticleNumber == articleNumber {
			return foundArticle, nil
		}
	}

	return article.Article{}, articleshelf.ErrArticleWasNotFound{}
}

// ListArticlesVisibleToThePublic returns only published stories.
func (memoryArticleShelf *MemoryArticleShelf) ListArticlesVisibleToThePublic() ([]article.Article, error) {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	visibleArticles := make([]article.Article, 0)
	for _, foundArticle := range memoryArticleShelf.articlesByPublicWebAddressSlug {
		if foundArticle.IsVisibleToThePublic {
			visibleArticles = append(visibleArticles, foundArticle)
		}
	}

	return sortArticlesNewestFirst(visibleArticles), nil
}

// ListEveryArticleForTheWriter returns drafts and published stories.
func (memoryArticleShelf *MemoryArticleShelf) ListEveryArticleForTheWriter() ([]article.Article, error) {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	everyArticle := make([]article.Article, 0, len(memoryArticleShelf.articlesByPublicWebAddressSlug))
	for _, foundArticle := range memoryArticleShelf.articlesByPublicWebAddressSlug {
		everyArticle = append(everyArticle, foundArticle)
	}

	return sortArticlesNewestFirst(everyArticle), nil
}

// RemoveArticleByPublicWebAddressSlug throws one article away.
func (memoryArticleShelf *MemoryArticleShelf) RemoveArticleByPublicWebAddressSlug(
	publicWebAddressSlug string,
) error {
	memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Lock()
	defer memoryArticleShelf.guardThatStopsTwoHandsAtOnce.Unlock()

	if _, articleAlreadyExists := memoryArticleShelf.articlesByPublicWebAddressSlug[publicWebAddressSlug]; !articleAlreadyExists {
		return articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: publicWebAddressSlug,
		}
	}

	delete(memoryArticleShelf.articlesByPublicWebAddressSlug, publicWebAddressSlug)
	return nil
}

func sortArticlesNewestFirst(articlesToSort []article.Article) []article.Article {
	sortedArticles := append([]article.Article{}, articlesToSort...)
	for firstIndex := 0; firstIndex < len(sortedArticles); firstIndex++ {
		for secondIndex := firstIndex + 1; secondIndex < len(sortedArticles); secondIndex++ {
			firstTime := sortedArticles[firstIndex].MomentTheArticleWasWritten
			secondTime := sortedArticles[secondIndex].MomentTheArticleWasWritten
			if secondTime.After(firstTime) {
				sortedArticles[firstIndex], sortedArticles[secondIndex] = sortedArticles[secondIndex], sortedArticles[firstIndex]
			}
		}
	}

	return sortedArticles
}
