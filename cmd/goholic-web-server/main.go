// Command goholic-web-server starts the goholic.in blog.
//
// A five year old map of the program:
// 1. Read the door keys from the environment.
// 2. Open the sqlite cabinet.
// 3. Hire a librarian.
// 4. Open the front desk.
// 5. Listen for browsers.
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anti-gravity-bit/goholic/internal/articlelibrarian"
	"github.com/anti-gravity-bit/goholic/internal/sqlitearticleshelf"
	"github.com/anti-gravity-bit/goholic/internal/wallclock"
	"github.com/anti-gravity-bit/goholic/internal/webpage"
	"github.com/anti-gravity-bit/goholic/internal/websiteidentity"
	"github.com/anti-gravity-bit/goholic/internal/writerpassport"
)

func main() {
	listenAddress := environmentValueOrFallback("GOHOLIC_LISTEN_ADDRESS", ":8080")
	sqliteFilePath := environmentValueOrFallback("GOHOLIC_DATABASE_FILE_PATH", "data/goholic.sqlite")
	writerUsername := environmentValueOrFallback("GOHOLIC_WRITER_USERNAME", "abir")
	writerPassword := environmentValueOrFallback("GOHOLIC_WRITER_PASSWORD", "change-me-now")
	cookieSecret := environmentValueOrFallback("GOHOLIC_SESSION_SECRET", "change-this-secret!")

	sqliteDirectory := filepath.Dir(sqliteFilePath)
	makeDirectoryError := os.MkdirAll(sqliteDirectory, 0o755)
	if makeDirectoryError != nil {
		log.Fatalf("could not create the data folder: %v", makeDirectoryError)
	}

	sqliteArticleShelf, openError := sqlitearticleshelf.OpenSqliteArticleShelfFile(sqliteFilePath)
	if openError != nil {
		log.Fatalf("could not open the article shelf: %v", openError)
	}
	defer sqliteArticleShelf.CloseTheCabinet()

	articleLibrarian := articlelibrarian.NewArticleLibrarian(sqliteArticleShelf, wallclock.RealWallClock{})
	plantError := articleLibrarian.PlantFirstStoriesIfTheShelfIsEmpty()
	if plantError != nil {
		log.Fatalf("could not plant the first stories: %v", plantError)
	}

	writerPassport, passportError := writerpassport.NewWriterPassport(
		writerUsername,
		writerPassword,
		cookieSecret,
		wallclock.RealWallClock{},
	)
	if passportError != nil {
		log.Fatalf("could not build the writer passport: %v", passportError)
	}

	webpageServer, webpageError := webpage.NewWebpageServer(
		websiteidentity.NewGoholicWebsiteIdentity(),
		articleLibrarian,
		writerPassport,
	)
	if webpageError != nil {
		log.Fatalf("could not open the front desk: %v", webpageError)
	}

	webpageServer.BindRoutesToDefaultServeMux()
	log.Printf("goholic is listening on %s", listenAddress)
	listenError := http.ListenAndServe(listenAddress, nil)
	if listenError != nil {
		log.Fatalf("the web server stopped: %v", listenError)
	}
}

func environmentValueOrFallback(environmentName string, fallbackValue string) string {
	foundValue := os.Getenv(environmentName)
	if foundValue == "" {
		return fallbackValue
	}

	return foundValue
}
