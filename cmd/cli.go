package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gophermap "github.com/icpmoles/gophermap/pkg"
)

func print_usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"gophermap - Generate sitemap of directory\n\n",
	)

	fmt.Fprintf(flag.CommandLine.Output(),
		"Usage: %s [-directory <folder> -output <output_file> -allow <ext>] <base_url>\n\n\t",
		os.Args[0])

	fmt.Fprintf(flag.CommandLine.Output(), "Where <base_url> should be in the form 'https://example.com'\n\n")

	fmt.Fprintln(flag.CommandLine.Output(), "Options:")
	flag.PrintDefaults()
}

type cliStringList []string

func (s *cliStringList) String() string {
	return strings.Join(*s, ",")
}

func (s *cliStringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	/* Parse CLI parameters*/
	folderPtr := flag.String("directory", ".", "Directory to analyze")
	sitemap := flag.String("output", "sitemap.xml", "output file")

	// we allow multiple allowed extensions
	var allowed cliStringList
	flag.Var(&allowed, "allow", "allowed extension to include (can be specified multiple times) (default 'md','pdf','txt','epub')")

	flag.Usage = print_usage
	flag.Parse()

	if len(allowed) == 0 {
		// default values
		allowed = gophermap.ExtensionsAllowList
	}

	args := flag.Args()
	if len(args) != 1 {
		print_usage()
		log.Fatal("Please provide <base_url> argument")
	}
	url := args[0]

	s_f, err := os.Create(*sitemap)
	if err != nil {
		log.Fatal("Error creating output file:\n\t", err.Error())
	}

	start := time.Now()
	err = gophermap.CreateSitemap(s_f, *folderPtr, url, allowed)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Template Executed in: ", time.Since(start))
}
