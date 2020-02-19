package cmd

import "testing"

func TestHTTPSlug(t *testing.T) {
	domain := "rusutsu.com"
	tests := []struct {
		path    string
		expPath string
	}{
		{
			path:    "/ski/kona/yuki.jpg",
			expPath: "/ski",
		},
		{
			path:    "/ski/",
			expPath: "/ski",
		},
		{
			path:    "/ski.jpg",
			expPath: "/",
		},
		{
			path:    "//",
			expPath: "/",
		},
	}

	for _, test := range tests {
		out := HTTPURLSlug(domain, test.path)
		exp := "http://" + domain + test.expPath
		if exp != out {
			t.Errorf("incorrect path returned for %q: returned: %q, exp: %q", test.path, out, exp)
		}
		t.Logf("%q", out)
	}

}
