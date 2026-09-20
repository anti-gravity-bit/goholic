// Package sqlitearticleshelf is a cabinet made of one sqlite file.
//
// The live website uses this cabinet so stories survive a restart.
package sqlitearticleshelf

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/anti-gravity-bit/goholic/internal/article"
	"github.com/anti-gravity-bit/goholic/internal/articleshelf"

	_ "modernc.org/sqlite"
)

// SqliteArticleShelf keeps articles in one file on disk.
type SqliteArticleShelf struct {
	openSqliteDatabase *sql.DB
}

// OpenSqliteArticleShelfFile opens or creates the cabinet file.
func OpenSqliteArticleShelfFile(sqliteFilePath string) (*SqliteArticleShelf, error) {
	openSqliteDatabase, openError := sql.Open("sqlite", sqliteFilePath)
	if openError != nil {
		return nil, fmt.Errorf("could not open sqlite file: %w", openError)
	}

	_, pragmeError := openSqliteDatabase.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
	`)
	if pragmeError != nil {
		return nil, fmt.Errorf("could not set sqlite rules: %w", pragmeError)
	}

	_, createTableError := openSqliteDatabase.Exec(`
		CREATE TABLE IF NOT EXISTS articles_on_the_shelf (
			article_number INTEGER PRIMARY KEY AUTOINCREMENT,
			public_web_address_slug TEXT NOT NULL UNIQUE,
			article_title TEXT NOT NULL,
			article_summary TEXT NOT NULL,
			article_body_markdown TEXT NOT NULL,
			is_visible_to_the_public INTEGER NOT NULL,
			moment_the_article_was_written TEXT NOT NULL,
			moment_the_article_last_changed TEXT NOT NULL
		);
	`)
	if createTableError != nil {
		return nil, fmt.Errorf("could not create articles table: %w", createTableError)
	}

	return &SqliteArticleShelf{
		openSqliteDatabase: openSqliteDatabase,
	}, nil
}

// CloseTheCabinet closes the sqlite file.
func (sqliteArticleShelf *SqliteArticleShelf) CloseTheCabinet() error {
	return sqliteArticleShelf.openSqliteDatabase.Close()
}

// PutArticleOnTheShelf stores a new article.
func (sqliteArticleShelf *SqliteArticleShelf) PutArticleOnTheShelf(
	articleToKeep article.Article,
) (article.Article, error) {
	insertResult, insertError := sqliteArticleShelf.openSqliteDatabase.Exec(
		`
			INSERT INTO articles_on_the_shelf (
				public_web_address_slug,
				article_title,
				article_summary,
				article_body_markdown,
				is_visible_to_the_public,
				moment_the_article_was_written,
				moment_the_article_last_changed
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		articleToKeep.PublicWebAddressSlug,
		articleToKeep.ArticleTitle,
		articleToKeep.ArticleSummary,
		articleToKeep.ArticleBodyMarkdown,
		boolAsSqliteInteger(articleToKeep.IsVisibleToThePublic),
		articleToKeep.MomentTheArticleWasWritten.UTC().Format(time.RFC3339Nano),
		articleToKeep.MomentTheArticleLastChanged.UTC().Format(time.RFC3339Nano),
	)
	if insertError != nil {
		return article.Article{}, articleshelf.ErrArticleSlugAlreadyTaken{
			PublicWebAddressSlug: articleToKeep.PublicWebAddressSlug,
		}
	}

	articleNumber, numberError := insertResult.LastInsertId()
	if numberError != nil {
		return article.Article{}, numberError
	}

	articleToKeep.ArticleNumber = articleNumber
	return articleToKeep, nil
}

// ReplaceArticleOnTheShelf overwrites an existing article.
func (sqliteArticleShelf *SqliteArticleShelf) ReplaceArticleOnTheShelf(
	articleToKeep article.Article,
) (article.Article, error) {
	updateResult, updateError := sqliteArticleShelf.openSqliteDatabase.Exec(
		`
			UPDATE articles_on_the_shelf
			SET
				article_title = ?,
				article_summary = ?,
				article_body_markdown = ?,
				is_visible_to_the_public = ?,
				moment_the_article_last_changed = ?
			WHERE public_web_address_slug = ?
		`,
		articleToKeep.ArticleTitle,
		articleToKeep.ArticleSummary,
		articleToKeep.ArticleBodyMarkdown,
		boolAsSqliteInteger(articleToKeep.IsVisibleToThePublic),
		articleToKeep.MomentTheArticleLastChanged.UTC().Format(time.RFC3339Nano),
		articleToKeep.PublicWebAddressSlug,
	)
	if updateError != nil {
		return article.Article{}, updateError
	}

	rowsChanged, rowsError := updateResult.RowsAffected()
	if rowsError != nil {
		return article.Article{}, rowsError
	}

	if rowsChanged == 0 {
		return article.Article{}, articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: articleToKeep.PublicWebAddressSlug,
		}
	}

	return articleToKeep, nil
}

// FindArticleByPublicWebAddressSlug looks up one article by its web name.
func (sqliteArticleShelf *SqliteArticleShelf) FindArticleByPublicWebAddressSlug(
	publicWebAddressSlug string,
) (article.Article, error) {
	rowFromSqlite := sqliteArticleShelf.openSqliteDatabase.QueryRow(
		`
			SELECT
				article_number,
				public_web_address_slug,
				article_title,
				article_summary,
				article_body_markdown,
				is_visible_to_the_public,
				moment_the_article_was_written,
				moment_the_article_last_changed
			FROM articles_on_the_shelf
			WHERE public_web_address_slug = ?
		`,
		publicWebAddressSlug,
	)

	return scanOneArticleFromSqlite(rowFromSqlite, publicWebAddressSlug)
}

// FindArticleByArticleNumber looks up one article by its number.
func (sqliteArticleShelf *SqliteArticleShelf) FindArticleByArticleNumber(
	articleNumber int64,
) (article.Article, error) {
	rowFromSqlite := sqliteArticleShelf.openSqliteDatabase.QueryRow(
		`
			SELECT
				article_number,
				public_web_address_slug,
				article_title,
				article_summary,
				article_body_markdown,
				is_visible_to_the_public,
				moment_the_article_was_written,
				moment_the_article_last_changed
			FROM articles_on_the_shelf
			WHERE article_number = ?
		`,
		articleNumber,
	)

	return scanOneArticleFromSqlite(rowFromSqlite, "")
}

// ListArticlesVisibleToThePublic returns published stories, newest first.
func (sqliteArticleShelf *SqliteArticleShelf) ListArticlesVisibleToThePublic() ([]article.Article, error) {
	return sqliteArticleShelf.listArticlesWithOptionalPublicFilter(true)
}

// ListEveryArticleForTheWriter returns drafts and published stories.
func (sqliteArticleShelf *SqliteArticleShelf) ListEveryArticleForTheWriter() ([]article.Article, error) {
	return sqliteArticleShelf.listArticlesWithOptionalPublicFilter(false)
}

// RemoveArticleByPublicWebAddressSlug throws one article away.
func (sqliteArticleShelf *SqliteArticleShelf) RemoveArticleByPublicWebAddressSlug(
	publicWebAddressSlug string,
) error {
	deleteResult, deleteError := sqliteArticleShelf.openSqliteDatabase.Exec(
		`DELETE FROM articles_on_the_shelf WHERE public_web_address_slug = ?`,
		publicWebAddressSlug,
	)
	if deleteError != nil {
		return deleteError
	}

	rowsChanged, rowsError := deleteResult.RowsAffected()
	if rowsError != nil {
		return rowsError
	}

	if rowsChanged == 0 {
		return articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: publicWebAddressSlug,
		}
	}

	return nil
}

func (sqliteArticleShelf *SqliteArticleShelf) listArticlesWithOptionalPublicFilter(
	onlyArticlesVisibleToThePublic bool,
) ([]article.Article, error) {
	sqlQuery := `
		SELECT
			article_number,
			public_web_address_slug,
			article_title,
			article_summary,
			article_body_markdown,
			is_visible_to_the_public,
			moment_the_article_was_written,
			moment_the_article_last_changed
		FROM articles_on_the_shelf
	`
	if onlyArticlesVisibleToThePublic {
		sqlQuery += ` WHERE is_visible_to_the_public = 1 `
	}
	sqlQuery += ` ORDER BY moment_the_article_was_written DESC `

	rowsFromSqlite, queryError := sqliteArticleShelf.openSqliteDatabase.Query(sqlQuery)
	if queryError != nil {
		return nil, queryError
	}
	defer rowsFromSqlite.Close()

	articlesFromSqlite := make([]article.Article, 0)
	for rowsFromSqlite.Next() {
		scannedArticle, scanError := scanArticleColumns(rowsFromSqlite)
		if scanError != nil {
			return nil, scanError
		}

		articlesFromSqlite = append(articlesFromSqlite, scannedArticle)
	}

	return articlesFromSqlite, rowsFromSqlite.Err()
}

type sqliteRowScanner interface {
	Scan(destinationColumns ...any) error
}

func scanOneArticleFromSqlite(
	rowFromSqlite sqliteRowScanner,
	publicWebAddressSlug string,
) (article.Article, error) {
	scannedArticle, scanError := scanArticleColumns(rowFromSqlite)
	if scanError == sql.ErrNoRows {
		return article.Article{}, articleshelf.ErrArticleWasNotFound{
			PublicWebAddressSlug: publicWebAddressSlug,
		}
	}
	if scanError != nil {
		return article.Article{}, scanError
	}

	return scannedArticle, nil
}

func scanArticleColumns(rowFromSqlite sqliteRowScanner) (article.Article, error) {
	var scannedArticle article.Article
	var visibleAsInteger int
	var writtenAtAsText string
	var changedAtAsText string

	scanError := rowFromSqlite.Scan(
		&scannedArticle.ArticleNumber,
		&scannedArticle.PublicWebAddressSlug,
		&scannedArticle.ArticleTitle,
		&scannedArticle.ArticleSummary,
		&scannedArticle.ArticleBodyMarkdown,
		&visibleAsInteger,
		&writtenAtAsText,
		&changedAtAsText,
	)
	if scanError != nil {
		return article.Article{}, scanError
	}

	writtenAt, writtenAtError := time.Parse(time.RFC3339Nano, writtenAtAsText)
	if writtenAtError != nil {
		return article.Article{}, writtenAtError
	}

	changedAt, changedAtError := time.Parse(time.RFC3339Nano, changedAtAsText)
	if changedAtError != nil {
		return article.Article{}, changedAtError
	}

	scannedArticle.IsVisibleToThePublic = visibleAsInteger == 1
	scannedArticle.MomentTheArticleWasWritten = writtenAt
	scannedArticle.MomentTheArticleLastChanged = changedAt
	return scannedArticle, nil
}

func boolAsSqliteInteger(flag bool) int {
	if flag {
		return 1
	}

	return 0
}
