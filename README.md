## Usage

Explores recursivelly the working directory or the specified folder looking for pdf, epub, md, text files and creates a sitemap.xml in the working directory.

```
gophermap 
```

### Positional arguments

- **BaseURL**: https://example.com/ (mandatory) URL of your website

- **path**: (optional) folder to explore 

### Notes

All the paths in the sitemap are referred to the working directory


## Credits

sitemap template taken from:
https://github.com/gohugoio/hugo/blob/master/tpl/tplimpl/embedded/templates/sitemap.xml 