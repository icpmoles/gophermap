package assets

import "embed"

// embedded sitemap.xml template.
// expects a SiteStructure object
//
//go:embed templates/sitemap.tmpl.xml
var Templates embed.FS
