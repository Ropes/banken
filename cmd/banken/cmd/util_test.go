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

func TestTopN(t *testing.T) {
	tests := []struct {
		m map[string]uint64
		n int
	}{
		{
			m: map[string]uint64{"/ski": uint64(100), "/wat": uint64(5), "/google": uint64(1000), "/": uint64(100000000)},
			n: 10,
		},
		{
			m: map[string]uint64{"/ski": uint64(100), "/wat": uint64(5), "/google": uint64(1000), "/": uint64(100000000)},
			n: 1,
		},
		{
			m: map[string]uint64{"/ski": uint64(100), "/wat": uint64(5), "/google": uint64(1000), "/": uint64(100000000)},
			n: 3,
		},
	}

	for _, test := range tests {
		t.Run("-", func(t *testing.T) {
			sorted := topNRequests(test.m, test.n)
			t.Logf("sorted: %v", sorted)
			if len(sorted) > test.n {
				t.Errorf("number of elements returned greater than n[%d]: %d", test.n, len(sorted))
			}
			if len(sorted) > 2 {
				if sorted[0].C < sorted[1].C {
					t.Errorf("sorted request order is incorrect: %v", sorted)
				}
			}

		})
	}

}
