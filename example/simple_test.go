package main

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	root := t.TempDir()
	included := []string{"notes.md", "docs/manual.pdf"}
	excluded := []string{"readme.txt", "book.epub", "image.png"}

	for _, path := range append(included, excluded...) {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// <lastmod> has a 1 second resolution
	before := time.Now().UTC().Truncate(time.Second)
	var output bytes.Buffer
	if err := run(&output, []string{root}); err != nil {
		t.Fatal(err)
	}
	after := time.Now().UTC()

	var sitemap struct {
		URLs []struct {
			Location   string `xml:"loc"`
			LastMod    string `xml:"lastmod"`
			Changefreq string `xml:"changefreq"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(output.Bytes(), &sitemap); err != nil {
		t.Fatalf("run() produced invalid XML: %v\n%s", err, output.String())
	}

	// only md and pdf files are allowed, txt and epub are excluded even if they are in the default list
	found := make(map[string]bool, len(sitemap.URLs))
	for _, url := range sitemap.URLs {
		found[url.Location] = true

		if url.Changefreq != "weekly" {
			t.Errorf("%s: changefreq = %q, want %q", url.Location, url.Changefreq, "weekly")
		}

		lastMod, err := time.Parse(time.RFC3339, url.LastMod)
		if err != nil {
			t.Errorf("%s: lastmod %q is not a valid timestamp: %v", url.Location, url.LastMod, err)
			continue
		}
		if lastMod.Before(before) || lastMod.After(after) {
			t.Errorf("%s: lastmod = %v, want the execution time (between %v and %v)", url.Location, lastMod, before, after)
		}
	}

	if len(sitemap.URLs) != len(included) {
		t.Errorf("run() produced %d URLs, want %d", len(sitemap.URLs), len(included))
	}
	for _, path := range included {
		if want := "https://example.com/" + filepath.Join(root, path); !found[want] {
			t.Errorf("run() did not include %q", want)
		}
	}
}
