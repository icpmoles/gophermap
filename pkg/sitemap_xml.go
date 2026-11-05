package gophermap

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/icpmoles/gophermap/types"
)

// same declaration emitted by the template
const sitemapXMLHeader = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>` + "\n"

// same layout used by the template for <lastmod>
const sitemapTimeLayout = "2006-01-02T15:04:05-07:00"

/*
File names in types.SiteStructure are already escaped by escapeXML, encoding them as
chardata would escape them twice (& -> &amp;amp;). innerxml writes them verbatim,
exactly like the template does
*/
type rawXML struct {
	Value string `xml:",innerxml"`
}

type sitemapURL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        rawXML   `xml:"loc"`
	LastMod    string   `xml:"lastmod"`
	ChangeFreq string   `xml:"changefreq"`
}

/*
Renders site with encoding/xml instead of text/template.
The <url> elements are streamed one at a time, so the whole document is never held in memory.
The output is semantically equivalent to writeSitemapTemplate (whitespace may differ)
*/
func writeSitemapXML(wr io.Writer, site types.SiteStructure) error {
	if _, err := io.WriteString(wr, sitemapXMLHeader); err != nil {
		return fmt.Errorf("Error writing XML header: %w", err)
	}

	enc := xml.NewEncoder(wr)
	enc.Indent("", "  ")

	urlset := xml.StartElement{
		Name: xml.Name{Local: "urlset"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns"}, Value: "http://www.sitemaps.org/schemas/sitemap/0.9"},
			{Name: xml.Name{Local: "xmlns:xhtml"}, Value: "http://www.w3.org/1999/xhtml"},
		},
	}
	if err := enc.EncodeToken(urlset); err != nil {
		return fmt.Errorf("Error encoding <urlset>: %w", err)
	}

	// reuse the same value for every entry, ChangeFreq never changes
	entry := sitemapURL{ChangeFreq: site.ChangeFreq}
	for _, f := range site.Files.Files {
		entry.Loc.Value = site.BaseURL + "/" + f.Name
		entry.LastMod = f.LastMod.Format(sitemapTimeLayout)
		if err := enc.Encode(&entry); err != nil {
			return fmt.Errorf("Error encoding <url> for %q: %w", f.Name, err)
		}
	}

	if err := enc.EncodeToken(urlset.End()); err != nil {
		return fmt.Errorf("Error encoding </urlset>: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("Error flushing XML encoder: %w", err)
	}
	return nil
}
