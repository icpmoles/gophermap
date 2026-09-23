package gophermap

import (
	"github.com/icpmoles/gophermap/assets"
	"github.com/icpmoles/gophermap/types"

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
func explorerWrapper(path string, d fs.DirEntry, err error, allowList *[]string, ff *types.FlattenedFolder, timestamp *time.Time) error {
	if err != nil {
		return fmt.Errorf("Walking %q: %w", path, err)
	}

	if !d.IsDir() && isAllowedFile(path, allowList) {
		info, err := d.Info()

		if err != nil {
			return fmt.Errorf("Getting file info for %q: %w", path, err)
		}

		var LastModTimestamp time.Time
		if timestamp == nil {
			LastModTimestamp = info.ModTime().UTC()
		} else {
			LastModTimestamp = *timestamp
		}

		file := types.File{
			Name:    path,
			LastMod: LastModTimestamp,
		}

		ff.Files = append(ff.Files, file)
	}

	return nil
}

/*
Returns list of files relative to path.
  - timestamp: arbitrary timestamp to use for <lastmod> field. If equal to nil means use the modifiedTimestamp
    from the filesystem
*/
func GetFlattenedFolder(explorePath string, allowList []string, timestamp *time.Time) (ff types.FlattenedFolder, err error) {

	err = filepath.WalkDir(explorePath,
		func(path string, d fs.DirEntry, err error) error {
			return explorerWrapper(path, d, err, &allowList, &ff, timestamp)
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
func CreateSitemap(wr io.Writer, explorePath string, baseURL string, setters ...types.Option) error {

	// Default Options
	args := &types.Options{
		AllowList:        ExtensionsAllowList,
		UseExecutionTime: false,
		Frequency:        types.Never,
	}

	for _, setter := range setters {
		setter(args)
	}

	fmt.Println("Exploring ", explorePath)

	lowerAllowList := make([]string, len(args.AllowList))

	for i, word := range args.AllowList {
		lowerAllowList[i] = strings.ToLower(word)
	}

	var timestamp *time.Time

	if args.UseExecutionTime {
		t := time.Now().UTC()
		timestamp = &t
	} else {
		timestamp = nil
	}

	files, err := GetFlattenedFolder(explorePath, lowerAllowList, timestamp)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d suitable files\n", len(files.Files))

	site := types.SiteStructure{
		Files:      files,
		BaseURL:    baseURL,
		ChangeFreq: args.Frequency.String(),
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
