package types

import (
	"fmt"
	"strings"
	"time"
)

// enum for sitemap frequency
// https://gobyexample.com/enums
type SiteMapFrequency int

const (
	Always SiteMapFrequency = iota
	Hourly
	Daily
	Weekly
	Monthly
	Yearly
	Never
)

var siteMapFrequencyName = map[SiteMapFrequency]string{
	Always:  "always",
	Hourly:  "hourly",
	Daily:   "daily",
	Weekly:  "weekly",
	Monthly: "monthly",
	Yearly:  "yearly",
	Never:   "never",
}

var nameSiteMapFrequency = map[string]SiteMapFrequency{
	"always":  Always,
	"hourly":  Hourly,
	"daily":   Daily,
	"weekly":  Weekly,
	"monthly": Monthly,
	"yearly":  Yearly,
	"never":   Never,
}

func (ss SiteMapFrequency) String() string {
	return siteMapFrequencyName[ss]
}

// explorePath string, baseURL string, allowList []string, useExecutionTime bool
// based on https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md
type Options struct {
	AllowList        []string
	UseExecutionTime bool
	Frequency        SiteMapFrequency
}

type Option func(*Options)

func WithAllowList(allowList []string) Option {
	return func(args *Options) {
		args.AllowList = allowList
	}
}

// if 'true' it uses the time of execution for <lastmod> instead of the time reported by the filesystem
func WithUseExecutionTime(useExecutionTime bool) Option {
	return func(args *Options) {
		args.UseExecutionTime = useExecutionTime
	}
}

func WithFrequency(f SiteMapFrequency) Option {
	return func(args *Options) {
		args.Frequency = f
	}
}

// accepts values defined by https://www.sitemaps.org/protocol.html#:~:text=both%20sources%20differently.-,%3Cchangefreq%3E,-optional
func WithFrequencyFromString(f string) Option {
	return func(args *Options) {
		freq, ok := nameSiteMapFrequency[strings.ToLower(f)]
		if ok {
			args.Frequency = freq
		} else {
			fmt.Printf("'%s' not accepted 'changefreq' value, defaulting to 'never'\n", f)
			args.Frequency = Never
		}
	}
}

/////////////////// internal representation of file system

type File struct {
	Name    string
	LastMod time.Time
}

// List of file types
type FlattenedFolder struct {
	Files []File
}

// Expected input type for sitemap template
type SiteStructure struct {
	Files      FlattenedFolder
	BaseURL    string
	ChangeFreq string
}
