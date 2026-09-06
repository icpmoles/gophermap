package gophermap

import (
	"gophermap/assets"

	"fmt"
	"io"
	"io/fs"
	"log"
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

type Folder struct {
	name    string
	files   []File
	folders []Folder
}

type FlattenedFolder struct {
	Files []File
}

type SiteStructure struct {
	Files   FlattenedFolder
	BaseURL string
}

func isAllowedFile(path string) bool {
	allowList := []string{"pdf", "txt", "epub", "md"}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	return slices.Contains(allowList, ext)
}

func getFlattenedFolder(path string) (ff FlattenedFolder, err error) {

	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && isAllowedFile(path) {
			info, err := d.Info()
			if err != nil {
				return err
			}

			file := File{Name: path, LastMod: info.ModTime()}
			// fmt.Println(path, info.ModTime())
			ff.Files = append(ff.Files, file)
		}

		return nil
	})

	return ff, err
}

func CreateSitemap(wr io.Writer, path string, baseurl string) (time.Duration, error) {
	start := time.Now()
	fmt.Println("Exploring ", path)

	files, _ := getFlattenedFolder(path)
	fmt.Printf("Found %d suitable files\n", len(files.Files))

	site := SiteStructure{
		Files:   files,
		BaseURL: baseurl,
	}

	ts, err := template.ParseFS(assets.Templates, "templates/sitemap.tmpl.xml")
	if err != nil {
		log.Print(err.Error())
	}

	err = ts.Execute(wr, site)

	return time.Since(start), err
}
