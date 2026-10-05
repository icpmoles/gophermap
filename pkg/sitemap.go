package gophermap

import (
	"github.com/icpmoles/gophermap/assets"
	"github.com/icpmoles/gophermap/types"

	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"text/template"
	"time"
)

// Default allowed extensions
var ExtensionsAllowList = []string{"pdf", "txt", "epub", "md"}

// Maximum number of URLs allowed in a single sitemap, see https://www.sitemaps.org/protocol.html
// It's a variable so tests can lower it instead of creating 50k files
var maxSitemapURLs = 50_000

/*
Checks if the file has the correct extension.
Expects a list of lowercase extensions
*/
func isAllowedFile(path string, allowList *[]string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	return slices.Contains(*allowList, strings.ToLower(ext))
}

/*
Escapes the characters that aren't allowed in XML text (&, <, >, quotes, ...).
text/template doesn't do it for us
*/
func escapeXML(s string) string {
	if !needsXMLEscape(s) {
		// most file names have nothing to escape: return them as-is without allocating
		return s
	}

	var b strings.Builder
	// writing to a strings.Builder never fails
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

/*
Reports whether xml.EscapeText could change s.
Only printable ASCII without &, <, >, " and ' is considered safe, anything else
(control characters, non-ASCII, invalid UTF-8) is left to xml.EscapeText
*/
func needsXMLEscape(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7e {
			return true
		}
		switch c {
		case '&', '<', '>', '"', '\'':
			return true
		}
	}
	return false
}

/*
fs.WalkDirFunc doesn't allow for custom arguments, so we use a wrapper that captures the function context
*/
func explorerWrapper(path string, d fs.DirEntry, err error, allowList *[]string, ff *types.FlattenedFolder, timestamp *time.Time) error {
	if err != nil {
		return fmt.Errorf("Walking %q: %w", path, err)
	}

	if !d.IsDir() && isAllowedFile(path, allowList) {
		var LastModTimestamp time.Time
		if timestamp == nil {
			// only stat the file when we actually need its modification time
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("Getting file info for %q: %w", path, err)
			}
			LastModTimestamp = info.ModTime().UTC()
		} else {
			LastModTimestamp = *timestamp
		}

		file := types.File{
			Name:    filepath.ToSlash(escapeXML(path)),
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

- explores the path(s) provided with explorePath

- filters all the files with extension that respect the allowList. (See ExtensionsAllowList for an example)

- calculates the final URL based on baseURL

- writes the resulting XML content to wr
*/
func CreateSitemap(wr io.Writer, explorePath []string, baseURL string, setters ...types.Option) error {

	// Default Options
	args := &types.Options{
		AllowList:        ExtensionsAllowList,
		UseExecutionTime: false,
		Frequency:        types.Never,
		Logger:           slog.New(slog.DiscardHandler),
	}

	for _, setter := range setters {
		setter(args)
	}

	args.Logger.Info("exploring", "paths", explorePath)

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

	// use go routines, one for each path
	results := make([]types.FlattenedFolder, len(explorePath))
	errs := make([]error, len(explorePath))
	var wg sync.WaitGroup
	for i, v := range explorePath {
		wg.Go(func() {
			results[i], errs[i] = GetFlattenedFolder(v, lowerAllowList, timestamp)
		})
	}
	wg.Wait()

	// reconsruct the results and checks for errors
	total := 0
	for i := range explorePath {
		if errs[i] != nil {
			return errs[i]
		}
		total += len(results[i].Files)
	}

	// allocate once instead of growing the slice folder after folder
	files := types.FlattenedFolder{Files: make([]types.File, 0, total)}
	for i := range results {
		files.Files = append(files.Files, results[i].Files...)
	}

	total_files := len(files.Files)
	args.Logger.Info("found suitable files", "count", total_files)

	if total_files > maxSitemapURLs {
		args.Logger.Error("too many files for sitemap", "count", total_files, "max", maxSitemapURLs)
	}

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
