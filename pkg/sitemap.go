package gophermap

import (
	"gophermap/assets"

	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"
)

type File struct {
	Name    string
	LastMod time.Time
}

type FlattenedFolder struct {
	Files []File
}

type SiteStructure struct {
	Files   FlattenedFolder
	BaseURL string
}

/*
Expects a list of lowercase extensions
*/
func isAllowedFile(path string, allowList *[]string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	return slices.Contains(*allowList, strings.ToLower(ext))
}

/*
fs.WalkDirFunc doesn't allow for custom arguments, so we use a wrapper that captures the function context
*/
func ExplorerWrapper(path string, d fs.DirEntry, err error, allowList *[]string, ff *FlattenedFolder) error {
	if err != nil {
		return fmt.Errorf("Walking %q: %w", path, err)
	}

	if !d.IsDir() && isAllowedFile(path, allowList) {
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("Getting file info for %q: %w", path, err)
		}

		file := File{
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
func getFlattenedFolder(explorePath string, allowList []string) (ff FlattenedFolder, err error) {

	err = filepath.WalkDir(explorePath,
		func(path string, d fs.DirEntry, err error) error {
			return ExplorerWrapper(path, d, err, &allowList, &ff)
		})

	if err != nil {
		return ff, fmt.Errorf("Error exploring target directory %q:\n\t%w", explorePath, err)
	}

	return ff, err
}

func CreateSitemap(wr io.Writer, explorePath string, baseURL string, allowList []string) error {

	fmt.Println("Exploring ", explorePath)

	lowerAllowList := make([]string, len(allowList))

	for i, word := range allowList {
		lowerAllowList[i] = strings.ToLower(word)
	}
	files, err := getFlattenedFolder(explorePath, lowerAllowList)
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
