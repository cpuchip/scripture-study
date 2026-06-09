package main

import "testing"

// normalizeRepoURL owns the allow-list-match contract: the sandbox wants a
// full clone URL, and the R10 smoke proved a bare name gets rejected there
// (and the cheap researcher fumbles instead of recovering).
func TestNormalizeRepoURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"ai-chattermax", "https://github.com/cpuchip/ai-chattermax"},
		{"ai-chattermax.git", "https://github.com/cpuchip/ai-chattermax"},
		{"cpuchip/ai-chattermax", "https://github.com/cpuchip/ai-chattermax"},
		{"someorg/somerepo", "https://github.com/someorg/somerepo"},
		{"someorg/somerepo.git", "https://github.com/someorg/somerepo"},
		{"https://github.com/cpuchip/ai-chattermax", "https://github.com/cpuchip/ai-chattermax"},
		// full URLs pass through verbatim — including .git, which is a valid clone URL
		{"https://github.com/cpuchip/ai-chattermax.git", "https://github.com/cpuchip/ai-chattermax.git"},
		{"http://gitea.local/o/r", "http://gitea.local/o/r"},
		{"  spin  ", "https://github.com/cpuchip/spin"},
		{"", ""},
	}
	for _, c := range cases {
		if got := normalizeRepoURL(c.in); got != c.want {
			t.Errorf("normalizeRepoURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
