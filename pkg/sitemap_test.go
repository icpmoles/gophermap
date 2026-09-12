package gophermap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"uuid"
)

type pathTable struct {
	path    string
	allowed bool
}

func generateRandomDirectory(size int) []pathTable {

	paths := []pathTable{}

	allowedExtensions := []string{
		"pdf",
		"txt",
		"epub",
		"md",
		"PdF",
		"Txt",
		"Epub",
		"MD",
		"md.txt",
		"txt.pdf",
		"md.pdf",
	}

	ForbiddenExtensions := []string{
		"png",
		"doc",
		"docx",
		"md.png",
		"epub.doc",
		"pdfx",
		"pdf.",
		"pdfx",
		"spdf",
	}

	Folders := []string{
		"assets/A/",
		"assets/B/",

		"pictures/AB/",
		"pictures/B/",

		"foo/AB/",
		"foo/B/",

		"foo/bar/car/fool/AB/",
		"foo/bar/car/fool/B/",

		"random/AB/",
		"random/B/",

		"downloads/AB/",
		"downloads/B/",

		"",
	}

	// var allowedPath bool
	for i := range size {

		// allowedPath = i%2 == 1
		filename := uuid.NewV4().String()

		// if allowedPath {
		allowedPath := pathTable{
			path: "allowed/" +
				Folders[i%len(Folders)] +
				filename + "." +
				allowedExtensions[i%len(allowedExtensions)],
			allowed: true}
		paths = append(paths, allowedPath)

		forbiddenPath := pathTable{
			path: "forbidden/" +
				Folders[i%len(Folders)] +
				filename + "." +
				ForbiddenExtensions[i%len(ForbiddenExtensions)],
			allowed: false}
		paths = append(paths,
			forbiddenPath)

	}

	return paths

}

func TestIsAllowedFileLong(t *testing.T) {
	paths := generateRandomDirectory(100)

	for _, test := range paths {
		if got := isAllowedFile(test.path, &ExtensionsAllowList); got != test.allowed {
			t.Fatalf("isAllowedFile (long) (%q) = %v, want %v", test.path, got, test.allowed)
		}
	}
}

func TestIsAllowedFile(t *testing.T) {
	tests := []pathTable{
		// Correct Extension
		{path: "document.pdf", allowed: true},
		{path: "notes.txt", allowed: true},
		{path: "book.epub", allowed: true},
		{path: "README.md", allowed: true},
		{path: "document.PDF", allowed: true},
		{path: "goo.txt.epub", allowed: true},

		// Incorrect extension
		{path: "image.png", allowed: false},
		{path: "goo.doc", allowed: false},

		// Edge cases
		{path: "goo.pdf.doc", allowed: false}, // nested extension: take the rightmost
		{path: "txt", allowed: false},         // no extension, but one of the allowed one
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if got := isAllowedFile(test.path, &ExtensionsAllowList); got != test.allowed {
				t.Fatalf("isAllowedFile(%q) = %v, want %v", test.path, got, test.allowed)
			}
		})
	}
}

func TestGetFlattenedFolder(t *testing.T) {
	root := t.TempDir()
	paths_accepted := []string{
		"notes/readme.md",
		"books/book.epub"}
	paths_ignored := []string{
		"ignored/image.png",
		"ignored/archive.zip",
	}

	paths := append(paths_accepted, paths_ignored...)

	for _, path := range paths {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(path), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetFlattenedFolder(root, []string{"pdf", "txt", "epub", "md"})
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Files) != 2 {
		t.Fatalf("GetFlattenedFolder() found %d files, want 2", len(got.Files))
	}

	found := make(map[string]time.Time, len(got.Files))
	for _, file := range got.Files {
		found[file.Name] = file.LastMod
	}

	for _, path := range paths_accepted {
		fullPath := filepath.Join(root, path)
		modTime, exists := found[fullPath]
		if !exists {
			t.Errorf("GetFlattenedFolder() did not include %q", fullPath)
			continue
		}
		if modTime.Location() != time.UTC {
			t.Errorf("modification time for %q has location %v, want UTC", fullPath, modTime.Location())
		}
	}

	for _, path := range paths_ignored {
		fullPath := filepath.Join(root, path)
		_, exists := found[fullPath]
		if exists {
			t.Errorf("GetFlattenedFolder() should not include %q", fullPath)
			continue
		}
	}
}

func TestGetFlattenedFolderMissingPath(t *testing.T) {
	// provide an unitialized subdirectory
	_, err := GetFlattenedFolder(filepath.Join(t.TempDir(), "missing"),
		[]string{"pdf"})
	if err == nil {
		t.Fatal("GetFlattenedFolder() returned nil error for a missing path")
	}
	if !strings.Contains(err.Error(), "Error exploring target directory") {
		t.Fatalf("GetFlattenedFolder() error = %q, want exploration context", err)
	}
}

func TestCreateSitemap(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "notes.md")
	if err := os.WriteFile(filePath, []byte("notes"), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := CreateSitemap(&output, root, "https://example.com", ExtensionsAllowList); err != nil {
		t.Fatal(err)
	}

	var sitemap struct {
		URLs []struct {
			Location string `xml:"loc"`
			LastMod  string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(output.Bytes(), &sitemap); err != nil {
		t.Fatalf("CreateSitemap() produced invalid XML: %v", err)
	}

	if len(sitemap.URLs) != 1 {
		t.Fatalf("CreateSitemap() produced %d URLs, want 1", len(sitemap.URLs))
	}
	if got, want := sitemap.URLs[0].Location, "https://example.com/"+filePath; got != want {
		t.Errorf("URL location = %q, want %q", got, want)
	}
	if sitemap.URLs[0].LastMod == "" {
		t.Error("URL lastmod is empty")
	}
}

func TestCreateSitemapReturnsExplorationError(t *testing.T) {
	var output bytes.Buffer
	err := CreateSitemap(&output, filepath.Join(t.TempDir(), "missing"),
		"https://example.com",
		[]string{"pdf", "txt", "epub", "md"})
	if err == nil {
		t.Fatal("CreateSitemap() returned nil error for a missing directory")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("CreateSitemap() error = %v, want an os.ErrNotExist cause", err)
	}
}

func BenchmarkGetFlattenedFolder(b *testing.B) {
	root := b.TempDir()
	paths := generateRandomDirectory(400)

	for _, path := range paths {
		fullPath := filepath.Join(root, path.path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(path.path), 0o644); err != nil {
			b.Fatal(err)
		}
	}

	for b.Loop() {
		_, err := GetFlattenedFolder(root, ExtensionsAllowList)
		if err != nil {
			b.Fatal("GetFlattenedFolder() returned nil error")
		}
	}
}
