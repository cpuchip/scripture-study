package auth

import "testing"

func TestIsAllowedRedirect(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"/today", true},
		{"/practices", true},
		{"//evil.com/phish", false},
		{"https://ibeco.me/", true},
		{"https://ibeco.me/today", true},
		{"https://1828.ibeco.me/", true},
		{"https://sub.deep.ibeco.me/", true},
		{"http://ibeco.me/insecure-but-allowed", true},
		{"https://evil.com/phish", false},
		{"https://ibeco.me.evil.com/", false},
		{"https://notibeco.me/", false},
		{"javascript:alert(1)", false},
		{"ftp://ibeco.me/", false},
	}
	for _, c := range cases {
		if got := isAllowedRedirect(c.in); got != c.want {
			t.Errorf("isAllowedRedirect(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
