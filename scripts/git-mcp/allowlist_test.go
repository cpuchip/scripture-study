package main

import "testing"

func TestValidateAgentBranch(t *testing.T) {
	cases := []struct {
		name    string
		want_ok bool
	}{
		// Valid
		{"agent/study-write/80424ffe-mysteries-of-god", true},
		{"agent/research/abc-some-slug", true},
		{"agent/teaching/0001", true},
		// Empty / wrong shape
		{"", false},
		{"agent/study-write", false},                // missing tail
		{"agent/study-write/", false},               // empty tail
		{"agent//id-slug", false},                   // empty pipeline
		{"study-write/abc", false},                  // missing agent/ prefix
		// Protected branches
		{"main", false},
		{"master", false},
		{"release/2026-04", false},
		// Disallowed chars
		{"agent/study-write/ABC123", false}, // uppercase
		{"agent/study-write/abc def", false},
		{"agent/study-write/abc/123", false},
		{"agent/study-write/abc..", false},
	}
	for _, c := range cases {
		err := validateAgentBranch(c.name)
		got_ok := err == nil
		if got_ok != c.want_ok {
			t.Errorf("validateAgentBranch(%q) = err=%v, want_ok=%v", c.name, err, c.want_ok)
		}
	}
}

func TestBuildAgentBranchName(t *testing.T) {
	cases := []struct {
		pipeline, id, slug string
		want_branch        string
		want_err           bool
	}{
		{"study-write", "80424ffe", "mysteries-of-god",
			"agent/study-write/80424ffe-mysteries-of-god", false},
		{"research", "abc123", "",
			"agent/research/abc123", false},
		{"study-write", "80424ffe", "this-is-a-very-long-slug-that-exceeds-forty-chars-and-needs-trim",
			"agent/study-write/80424ffe-this-is-a-very-long-slug-that-exceeds-fo", false},
		{"", "abc", "slug", "", true},
		{"pipeline", "", "slug", "", true},
		// Pipeline name normalized
		{"Study Write!", "abc", "slug", "agent/study-write/abc-slug", false},
	}
	for _, c := range cases {
		got, err := buildAgentBranchName(c.pipeline, c.id, c.slug)
		if c.want_err {
			if err == nil {
				t.Errorf("buildAgentBranchName(%q,%q,%q) = (%q, nil), want error",
					c.pipeline, c.id, c.slug, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("buildAgentBranchName(%q,%q,%q) = err %v, want %q",
				c.pipeline, c.id, c.slug, err, c.want_branch)
			continue
		}
		if got != c.want_branch {
			t.Errorf("buildAgentBranchName(%q,%q,%q) = %q, want %q",
				c.pipeline, c.id, c.slug, got, c.want_branch)
		}
	}
}

func TestValidateWorkdirID(t *testing.T) {
	cases := []struct {
		id      string
		want_ok bool
	}{
		{"80424ffe-df5d-48f6-b350-a3da355b290e", true},
		{"abc123", true},
		{"", false},
		{"../../../etc/passwd", false},
		{"abc/def", false},
		{"abc def", false},
		{"abc.def", false},
	}
	for _, c := range cases {
		err := validateWorkdirID(c.id)
		got_ok := err == nil
		if got_ok != c.want_ok {
			t.Errorf("validateWorkdirID(%q) = err=%v, want_ok=%v", c.id, err, c.want_ok)
		}
	}
}
