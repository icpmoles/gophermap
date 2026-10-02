// Example of serving a sitemap over HTTP with gophermap.
//
// The sitemap is generated on every request, so it always reflects the current
// content of the folder:
//
//	go run ./example/server -dir ./public -base https://example.com
//	curl http://localhost:8080/sitemap.xml
//
// This only an example. Don't run this in production

package main

import (
	"bytes"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	gophermap "github.com/icpmoles/gophermap/pkg"
	"github.com/icpmoles/gophermap/types"
)

// returns a handler that writes the sitemap of folders, with URLs starting with baseURL
func sitemapHandler(folders []string, baseURL string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		// build the sitemap in memory first: if something goes wrong we can still
		// answer with a 500 instead of sending half a document with a 200
		var buf bytes.Buffer
		err := gophermap.CreateSitemap(&buf, folders, baseURL,
			types.WithFrequency(types.Weekly),
			types.WithLogger(logger),
		)
		if err != nil {
			logger.Error("creating sitemap", "err", err)
			http.Error(w, "could not generate sitemap", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		if _, err := buf.WriteTo(w); err != nil {
			logger.Error("writing response", "err", err)
		}

		logger.Info("sitemap served", "duration", time.Since(start))

	}
}

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	dir := flag.String("dir", ".", "directory to include in the sitemap")
	baseURL := flag.String("base", "http://localhost:8080", "base URL of the files in the sitemap")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	mux := http.NewServeMux()
	// only GET (and HEAD) requests, anything else gets a 405
	mux.HandleFunc("GET /sitemap.xml", sitemapHandler([]string{*dir}, *baseURL, logger))

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("serving sitemap", "url", "http://localhost"+*addr+"/sitemap.xml")
	log.Fatal(server.ListenAndServe())
}
