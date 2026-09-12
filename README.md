[![Build GopherMap](https://github.com/icpmoles/gophermap/actions/workflows/build.yaml/badge.svg?branch=main&event=push)](https://github.com/icpmoles/gophermap/actions/workflows/build.yaml)
[![Release](https://img.shields.io/github/release/icpmoles/gophermap.svg?style=flat-square)](https://github.com/icpmoles/gophermap/releases/latest)
[![Casual Maintenance Intended](https://casuallymaintained.tech/badge.svg)](https://casuallymaintained.tech/)


## Usage

Explores recursivelly the working directory or the specified folder looking for pdf, epub, md, txt files and creates a sitemap.xml in the working directory.

```terminal
gophermap https://example.com
gophermap -folder assets -o public/sitemap.xml https://example.com
```

```terminal
% gophermap -h

gophermap - Generate sitemap of directory

Usage: gophermap [-directory <folder> -output <output_file> -allow <ext>] <base_url>

        Where <base_url> should be in the form 'https://example.com'

Options:
  -allow value
        allowed extension (can be specified multiple times) (default 'md','pdf','txt','md')
  -directory string
        Directory to analyze (default ".")
  -output string
        output file (default "sitemap.xml")
```

### Arguments

- **base_url**: https://example.com/ (mandatory) Base URL of your website defined as 

  The absolute URL of your published site including the protocol, host, path, and a trailing slash.

- **folder**: (optional) directory to explore 

- **output**: (optional) output file

- **allow**: (optional) extension to inclue

## Benchmarks

Tested on my own personal *Downloads* folder with approximately 6786 directories & 64515 files

```
Exploring  /home/icpmoles/Downloads
Found 4635 suitable files
Template Executed in:  1.133993069s
```

## Builds

Builds are provided for Linux, Windows and MacOS (darwin)

The CPU architecture include x64 & ARM64 for all OS.

Specifically for Linux there are additional ARMv5/ARMv6/ARMv7 builds meant for low memory devices like Raspberry Pi running on 32 bit distros.


## Credits

- Sitemap template taken from [Hugo CMS](https://github.com/gohugoio/hugo/blob/master/tpl/tplimpl/embedded/templates/sitemap.xml) (Apache-2)
