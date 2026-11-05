package gophermap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/icpmoles/gophermap/types"
)

// builds a SiteStructure with n files, every 10th name needs escaping
func makeSiteStructure(n int) types.SiteStructure {
	ts := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	files := make([]types.File, n)
	for i := range files {
		name := fmt.Sprintf("dir%d/sub/file_%d.md", i%17, i)
		if i%10 == 0 {
			name = fmt.Sprintf("dir%d/Q&A <%d>.pdf", i%17, i)
		}
		files[i] = types.File{Name: escapeXML(name), LastMod: ts.Add(time.Duration(i) * time.Second)}
	}
	return types.SiteStructure{
		Files:      types.FlattenedFolder{Files: files},
		BaseURL:    "https://example.com",
		ChangeFreq: types.Weekly.String(),
	}
}

type parsedSitemap struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []struct {
		Loc        string `xml:"loc"`
		LastMod    string `xml:"lastmod"`
		ChangeFreq string `xml:"changefreq"`
	} `xml:"url"`
}

func TestWriteSitemapXMLMatchesTemplate(t *testing.T) {
	for _, n := range []int{0, 1, 25} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			site := makeSiteStructure(n)

			var tmplOut, xmlOut bytes.Buffer
			if err := writeSitemapTemplate(&tmplOut, site); err != nil {
				t.Fatal(err)
			}
			if err := writeSitemapXML(&xmlOut, site); err != nil {
				t.Fatal(err)
			}

			var fromTmpl, fromXML parsedSitemap
			if err := xml.Unmarshal(tmplOut.Bytes(), &fromTmpl); err != nil {
				t.Fatalf("template produced invalid XML: %v", err)
			}
			if err := xml.Unmarshal(xmlOut.Bytes(), &fromXML); err != nil {
				t.Fatalf("encoding/xml produced invalid XML: %v\n%s", err, xmlOut.String())
			}
			if !reflect.DeepEqual(fromTmpl, fromXML) {
				t.Fatalf("outputs differ\ntemplate:\n%s\nencoding/xml:\n%s", tmplOut.String(), xmlOut.String())
			}
			if len(fromXML.URLs) != n {
				t.Fatalf("got %d URLs, want %d", len(fromXML.URLs), n)
			}
			if n > 0 && fromXML.URLs[0].Loc != "https://example.com/dir0/Q&A <0>.pdf" {
				t.Errorf("escaped name decoded to %q", fromXML.URLs[0].Loc)
			}
		})
	}
}

func BenchmarkWriteSitemap(b *testing.B) {
	writers := []struct {
		name  string
		write func(io.Writer, types.SiteStructure) error
	}{
		{"template", writeSitemapTemplate},
		{"encoding_xml", writeSitemapXML},
	}

	for _, n := range []int{10, 1_000, 50_000} {
		site := makeSiteStructure(n)
		for _, w := range writers {
			b.Run(fmt.Sprintf("%s/files=%d", w.name, n), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if err := w.write(io.Discard, site); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
