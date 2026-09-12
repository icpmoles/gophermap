package gophermap

import (
	"github.com/icpmoles/gophermap/assets"

	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"
)

// Default allowed extensions
var ExtensionsAllowList = []string{"pdf", "txt", "epub", "md"}

type file struct {
	Name    string
	LastMod time.Time
}

// List of file types
type FlattenedFolder struct {
	Files []file
}

// Expected input type for sitemap template
type SiteStructure struct {
	Files   FlattenedFolder
	BaseURL string
}

/*
Checks if the file has the correct extension.
Expects a list of lowercase extensions
*/
func isAllowedFile(path string, allowList *[]string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	return slices.Contains(*allowList, strings.ToLower(ext))
}

/*
fs.WalkDirFunc doesn't allow for custom arguments, so we use a wrapper that captures the function context
*/
func explorerWrapper(path string, d fs.DirEntry, err error, allowList *[]string, ff *FlattenedFolder) error {
	if err != nil {
		return fmt.Errorf("Walking %q: %w", path, err)
	}

	if !d.IsDir() && isAllowedFile(path, allowList) {
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("Getting file info for %q: %w", path, err)
		}

		file := file{
			Name:    path,
			LastMod: info.ModTime().UTC(),
		}

		ff.Files = append(ff.Files, file)
	}

	return nil
}

/*
Returns list of files relative to path.
*/
func GetFlattenedFolder(explorePath string, allowList []string) (ff FlattenedFolder, err error) {

	err = filepath.WalkDir(explorePath,
		func(path string, d fs.DirEntry, err error) error {
			return explorerWrapper(path, d, err, &allowList, &ff)
		})

	if err != nil {
		return ff, fmt.Errorf("Error exploring target directory %q:\n\t%w", explorePath, err)
	}

	return ff, err
}

/*
CreateSitemap:
- explores the path provided with explorePath
- filters all the files with extension that respect the allowList. (See ExtensionsAllowList for an example)
- calculates the final URL based on baseURL
- writes the resulting XML content to wr
*/
func CreateSitemap(wr io.Writer, explorePath string, baseURL string, allowList []string) error {

	fmt.Println("Exploring ", explorePath)

	lowerAllowList := make([]string, len(allowList))

	for i, word := range allowList {
		lowerAllowList[i] = strings.ToLower(word)
	}
	files, err := GetFlattenedFolder(explorePath, lowerAllowList)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d suitable files\n", len(files.Files))

	site := SiteStructure{
		Files:   files,
		BaseURL: baseURL,
	}

	ts, err := template.ParseFS(assets.Templates, "templates/sitemap.tmpl.xml")
	if err != nil {
		return fmt.Errorf("Error parsing embedded XML template: %w", err)
	}

	err = ts.Execute(wr, site)
	if err != nil {
		return fmt.Errorf("Error executing XML template: %w", err)
	}

	return err
}
