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
			path:    "/ski",
			expPath: "/ski",
		},
		{
			path:    "//",
			expPath: "",
		},
	}

	for _, test := range tests {
		out := HTTPURLSlug(domain, test.path)
		exp := "http://" + domain + test.expPath
		if exp != out {
			t.Errorf("incorrect path returned: %q, exp: %q", out, exp)
		}
	}

}
