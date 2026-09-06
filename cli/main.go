package main

import (
	"embed"
	// "fmt"
	"io"
	"log"
	"os"
	"text/template"
)

//go:embed assets/*
var content embed.FS

type Folder struct {
	files   []string
	folders []Folder
}

func getSitemap(wr io.Writer, path string) error {

	ts, err := template.ParseFS(content, "assets/sitemap.tmpl.xml")
	if err != nil {
		log.Print(err.Error())
	}

	err = ts.Execute(wr, nil)

	return err
}

func main() {
	s_f, err := os.Create("sitemap.xml")
	if err != nil {
		log.Print(err.Error())
	}

	err = getSitemap(s_f, "./target")
	if err != nil {
		log.Print(err.Error())
	}
	// fmt.Println("Hello, World!")
}
