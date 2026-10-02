// Minimal example of using gophermap as a library.
//
// It scans the current directory for documents and prints the sitemap to stdout:
//
//	go run ./example
package main

import (
	"io"
	"log"
	"log/slog"
	"os"

	gophermap "github.com/icpmoles/gophermap/pkg"
	"github.com/icpmoles/gophermap/types"
)

func main() {
	// folders to scan, every matching file inside them ends up in the sitemap
	folders := []string{"."}

	// we use stdout, the library never prints anything else on it
	if err := run(os.Stdout, folders); err != nil {
		log.Fatal(err)
	}
}

// CreateSitemap writes to any io.Writer: a file, a buffer, an HTTP response...
func run(w io.Writer, folders []string) error {
	return gophermap.CreateSitemap(w, folders, "https://example.com",
		// only include markdown and PDF files (default: pdf, txt, epub, md)
		types.WithAllowList([]string{"md", "pdf"}),
		// value of <changefreq> (default: never)
		types.WithFrequency(types.Weekly),
		// use the time of execution for <lastmod> instead of each file's modification time
		types.WithUseExecutionTime(true),
		// progress messages go to stderr (default: no logging)
		types.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
	)
}
