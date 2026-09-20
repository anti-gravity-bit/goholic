// Package article is the heart of the blog.
//
// An Article is one story the writer wants to keep.
// Other packages may store articles or show articles,
// but they do not decide what an article *is*.
package article

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

// Article is one blog story.
type Article struct {
	ArticleNumber               int64
	PublicWebAddressSlug        string
	ArticleTitle                string
	ArticleSummary              string
	ArticleBodyMarkdown         string
	IsVisibleToThePublic        bool
	MomentTheArticleWasWritten  time.Time
	MomentTheArticleLastChanged time.Time
}

// NewDraftArticle builds a draft after checking the writer's words.
func NewDraftArticle(
	publicWebAddressSlug string,
	articleTitle string,
	articleSummary string,
	articleBodyMarkdown string,
	momentTheArticleWasWritten time.Time,
) (Article, error) {
	cleanedSlug, slugError := CleanPublicWebAddressSlug(publicWebAddressSlug)
	if slugError != nil {
		return Article{}, slugError
	}

	cleanedTitle := strings.TrimSpace(articleTitle)
	if cleanedTitle == "" {
		return Article{}, errors.New("an article needs a title")
	}

	cleanedSummary := strings.TrimSpace(articleSummary)
	if cleanedSummary == "" {
		return Article{}, errors.New("an article needs a short summary")
	}

	cleanedBody := strings.TrimSpace(articleBodyMarkdown)
	if cleanedBody == "" {
		return Article{}, errors.New("an article needs a body")
	}

	return Article{
		PublicWebAddressSlug:        cleanedSlug,
		ArticleTitle:                cleanedTitle,
		ArticleSummary:              cleanedSummary,
		ArticleBodyMarkdown:         cleanedBody,
		IsVisibleToThePublic:        false,
		MomentTheArticleWasWritten:  momentTheArticleWasWritten.UTC(),
		MomentTheArticleLastChanged: momentTheArticleWasWritten.UTC(),
	}, nil
}

// CleanPublicWebAddressSlug turns a title-like string into a safe web address piece.
//
// Example: "Hello Goholic World" becomes "hello-goholic-world".
func CleanPublicWebAddressSlug(rawPublicWebAddressSlug string) (string, error) {
	trimmedSlug := strings.TrimSpace(strings.ToLower(rawPublicWebAddressSlug))
	if trimmedSlug == "" {
		return "", errors.New("an article needs a public web address slug")
	}

	var slugBuilder strings.Builder
	lastCharacterWasAHyphen := false

	for _, currentRune := range trimmedSlug {
		if unicode.IsLetter(currentRune) || unicode.IsDigit(currentRune) {
			slugBuilder.WriteRune(currentRune)
			lastCharacterWasAHyphen = false
			continue
		}

		if currentRune == '-' || currentRune == ' ' || currentRune == '_' {
			if lastCharacterWasAHyphen || slugBuilder.Len() == 0 {
				continue
			}

			slugBuilder.WriteByte('-')
			lastCharacterWasAHyphen = true
		}
	}

	cleanedSlug := strings.Trim(slugBuilder.String(), "-")
	if cleanedSlug == "" {
		return "", errors.New("the public web address slug has no useful letters")
	}

	return cleanedSlug, nil
}

// WithUpdatedWriting returns a copy with new words and a new change time.
func (existingArticle Article) WithUpdatedWriting(
	articleTitle string,
	articleSummary string,
	articleBodyMarkdown string,
	momentTheArticleLastChanged time.Time,
) (Article, error) {
	updatedArticle, buildError := NewDraftArticle(
		existingArticle.PublicWebAddressSlug,
		articleTitle,
		articleSummary,
		articleBodyMarkdown,
		existingArticle.MomentTheArticleWasWritten,
	)
	if buildError != nil {
		return Article{}, buildError
	}

	updatedArticle.ArticleNumber = existingArticle.ArticleNumber
	updatedArticle.IsVisibleToThePublic = existingArticle.IsVisibleToThePublic
	updatedArticle.MomentTheArticleLastChanged = momentTheArticleLastChanged.UTC()
	return updatedArticle, nil
}

// ShownToThePublic returns a copy that readers on the internet can see.
func (existingArticle Article) ShownToThePublic(momentTheArticleLastChanged time.Time) Article {
	existingArticle.IsVisibleToThePublic = true
	existingArticle.MomentTheArticleLastChanged = momentTheArticleLastChanged.UTC()
	return existingArticle
}

// HiddenFromThePublic returns a copy that only the writer can see.
func (existingArticle Article) HiddenFromThePublic(momentTheArticleLastChanged time.Time) Article {
	existingArticle.IsVisibleToThePublic = false
	existingArticle.MomentTheArticleLastChanged = momentTheArticleLastChanged.UTC()
	return existingArticle
}
