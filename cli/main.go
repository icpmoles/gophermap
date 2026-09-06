package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"
)

//go:embed assets/*
var content embed.FS

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

	ts, err := template.ParseFS(content, "assets/sitemap.tmpl.xml")
	if err != nil {
		log.Print(err.Error())
	}

	err = ts.Execute(wr, site)

	return time.Since(start), err
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("BaseURL not provided!!")
		os.Exit(1)
	}
	baseurl := os.Args[1]
	folder := "."
	if len(os.Args) < 3 {
		fmt.Println("Writing in the same directory")
	} else {
		folder = os.Args[2]
	}

	s_f, err := os.Create("sitemap.xml")
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	execution_time, err := CreateSitemap(s_f, folder, baseurl)

	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	fmt.Println("Template Executed in: ", execution_time)
}
