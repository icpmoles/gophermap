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

func CreateSitemap(wr io.Writer, path string) (time.Duration, error) {
	start := time.Now()
	fmt.Println("Exploring ", path)
	files, _ := getFlattenedFolder(path)

	ts, err := template.ParseFS(content, "assets/sitemap.tmpl.xml")
	if err != nil {
		log.Print(err.Error())
	}

	err = ts.Execute(wr, files)

	return time.Since(start), err
}

func main() {
	folder := "."
	if len(os.Args) < 2 {
		fmt.Println("Writing in the same directory")
	} else {
		folder = os.Args[1]
	}

	s_f, err := os.Create("sitemap.xml")
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	execution_time, err := CreateSitemap(s_f, folder)

	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	fmt.Println("Template Executed in: ", execution_time)
}
