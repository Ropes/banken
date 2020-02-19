package cmd

import (
	"net/url"
	"strings"
)

// HTTPURLSlug reduces the path down to only its first element
// iff the path exists. Maintains the base / for all URLs which
// do not contain a section.
func HTTPURLSlug(domain, path string) string {
	slug := strings.Split(path[1:], "/")
	var p string
	if len(slug) >= 2 {
		p = slug[0]
		if p == "" {
			p = "/"
		}
	} else if len(slug) == 1 {
		p = "/"
	} else {
		p = "/"
	}
	u := url.URL{
		Scheme: "http",
		Host:   domain,
		Path:   p,
	}
	return u.String()
}
