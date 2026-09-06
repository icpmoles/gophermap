package main

import (
	"fmt"
	"gophermap/gophermap"
	"log"
	"os"
)

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

	execution_time, err := gophermap.CreateSitemap(s_f, folder, baseurl)

	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	fmt.Println("Template Executed in: ", execution_time)
}
