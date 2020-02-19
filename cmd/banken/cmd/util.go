package cmd

import (
	"net/url"
	"sort"
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

// ReqCount links URLs to their request occurrence count:C.
type ReqCount struct {
	URL string
	C   uint64
}

func topNRequests(m map[string]uint64, n int) []ReqCount {
	reqs := make([]ReqCount, 0)
	for k, v := range m {
		reqs = append(reqs, ReqCount{URL: k, C: v})
	}
	sort.Slice(reqs, func(i, j int) bool {
		return reqs[i].C > reqs[j].C
	})
	// Cap the length of returned data.
	if len(reqs) > n {
		reqs = reqs[:n]
	}
	return reqs
}
