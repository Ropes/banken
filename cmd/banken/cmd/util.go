package cmd

import (
	"net/url"
	"strings"
)

// HTTPURLSlug reduces the path down to only its first element
// iff the path exists.
func HTTPURLSlug(domain, path string) string {
	slug := strings.Split(path[1:], "/")
	var p string
	if len(slug) > 0 {
		p = slug[0]
	}
	u := url.URL{
		Scheme: "http",
		Host:   domain,
		Path:   p,
	}
	return u.String()
}