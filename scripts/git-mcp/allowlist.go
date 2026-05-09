// Allow-list helpers — branch namespace, slug normalization, protected
// branch list. Pure functions, no side effects. Easy to unit-test.

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// gitConfig is shared across tool handlers. Token is NOT stored here —
// looked up from env at exec time so rotation without restart works.
type gitConfig struct {
	WorkdirRoot   string
	GitBin        string
	GhBin         string
	CoAuthorEmail string
}

// Anchored regex for the substrate-driven branch namespace. Matches
// strings like:
//   agent/study-write/80424ffe-mysteries-of-god
//   agent/research/abc123-some-slug
//   agent/teaching/0001-no-slug
// All segments are lowercase letters, digits, hyphens. The trailing
// "-<slug>" is optional (work-item-id alone is valid).
var branchNamespaceRE = regexp.MustCompile(
	`^agent/[a-z0-9-]+/[a-z0-9]+(-[a-z0-9-]+)?$`,
)

// Protected branches the agent must never touch directly. Push to
// these returns a tool error.
var protectedBranches = map[string]bool{
	"main":   true,
	"master": true,
}

// protectedPatterns are prefix matches (e.g. "release/" matches
// "release/2026-04"). Used in addition to protectedBranches.
var protectedPatterns = []string{
	"release/",
}

// validateAgentBranch enforces the namespace + protected-branch rules
// for any branch name an agent tool wants to use. Returns a non-nil
// error if the name is invalid; caller turns the error into a tool
// reply.
func validateAgentBranch(name string) error {
	if name == "" {
		return fmt.Errorf("branch name is empty")
	}
	if protectedBranches[name] {
		return fmt.Errorf("branch %q is protected; agents cannot push or modify it", name)
	}
	for _, pfx := range protectedPatterns {
		if strings.HasPrefix(name, pfx) {
			return fmt.Errorf("branch %q matches protected pattern %q", name, pfx)
		}
	}
	if !branchNamespaceRE.MatchString(name) {
		return fmt.Errorf(
			"branch %q does not match agent namespace regex %s",
			name, branchNamespaceRE,
		)
	}
	return nil
}

// buildAgentBranchName composes the canonical name for a substrate
// pipeline run. Caller passes pipeline_family + work_item_id (UUID
// or short id) + slug (work_item.slug, possibly long). Slug is
// truncated to ~40 chars and normalized to lowercase-hyphen.
func buildAgentBranchName(pipeline, workItemID, slug string) (string, error) {
	pipeline = normalizeSegment(pipeline)
	workItemID = normalizeSegment(workItemID)
	if pipeline == "" || workItemID == "" {
		return "", fmt.Errorf(
			"pipeline and work_item_id are required (got pipeline=%q, work_item_id=%q)",
			pipeline, workItemID,
		)
	}
	tail := workItemID
	if slug = normalizeSegment(slug); slug != "" {
		if len(slug) > 40 {
			slug = slug[:40]
			// Trim trailing hyphen from truncation
			slug = strings.TrimRight(slug, "-")
		}
		tail = workItemID + "-" + slug
	}
	name := "agent/" + pipeline + "/" + tail
	if err := validateAgentBranch(name); err != nil {
		return "", err
	}
	return name, nil
}

// normalizeSegment lowercases, strips disallowed chars, collapses runs
// of hyphens. Used for pipeline / id / slug components before joining.
var segmentDisallowed = regexp.MustCompile(`[^a-z0-9-]+`)
var segmentMultiHyphen = regexp.MustCompile(`-+`)

func normalizeSegment(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = segmentDisallowed.ReplaceAllString(s, "-")
	s = segmentMultiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// shortenWorkItemID truncates a UUID to its first 8 chars for
// inclusion in branch names. Pass-through if already short.
func shortenWorkItemID(id string) string {
	id = normalizeSegment(id)
	if len(id) > 8 && strings.Count(id, "-") >= 4 {
		return strings.SplitN(id, "-", 2)[0]
	}
	return id
}

// validateWorkdirID — UUID or short id. The work_item_id arrives from
// the agent as a tool arg, and we use it to derive the workdir path.
// Anchor + alphanumeric+hyphen to prevent path injection.
var workdirIDRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateWorkdirID(id string) error {
	if id == "" {
		return fmt.Errorf("work_item_id is required")
	}
	if !workdirIDRE.MatchString(strings.ToLower(id)) {
		return fmt.Errorf(
			"work_item_id %q must match %s (alphanumeric + hyphen)",
			id, workdirIDRE,
		)
	}
	return nil
}
