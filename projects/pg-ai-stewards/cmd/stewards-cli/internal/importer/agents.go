package importer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

// AgentDoc is the parsed shape of one Copilot/Claude agent file.
//
//	.github/agents/<name>.agent.md      — Copilot format (YAML list tools)
//	.claude/agents/<name>.md            — Claude format (comma-string tools)
//	.stewards/<model>/<name>.agent.md   — model-tuned variant (declares model_match)
type AgentDoc struct {
	Family      string   // filename without .agent.md or .md
	Description string   // frontmatter description
	Body        string   // markdown body (becomes agents.prompt)
	Tools       []string // tool patterns to allow
	Model       string   // optional preferred model (logged, not stored yet)
	ModelMatch  string   // optional glob for stewards.agents.model_match; '' → '*'
	Handoffs    string   // optional handoffs YAML (logged, not stored yet)
}

// agentFrontmatter is the lenient YAML shape we accept. tools may be
// either a list (Copilot YAML) or a single comma-separated string
// (Claude format); we normalize after parse.
//
// model_match is the new field as of Phase 3c.3.3 — when set, the
// agent row is upserted under (family, model_match) instead of the
// default (family, '*'), enabling per-model prompt variants.
type agentFrontmatter struct {
	Description string      `yaml:"description"`
	Name        string      `yaml:"name"`
	Tools       interface{} `yaml:"tools"`
	Model       string      `yaml:"model"`
	ModelMatch  string      `yaml:"model_match"`
	Handoffs    interface{} `yaml:"handoffs,omitempty"`
}

// parseAgentMarkdown reads one *.agent.md or .md agent file and produces
// an AgentDoc. Returns an error on YAML parse failure; callers may
// continue past individual failures.
func parseAgentMarkdown(absPath string) (*AgentDoc, error) {
	raw, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}

	m := yamlFrontRe.FindStringSubmatchIndex(raw)
	if m == nil {
		return nil, fmt.Errorf("no YAML frontmatter found")
	}
	yamlBlock := raw[m[2]:m[3]]
	body := raw[m[1]:]

	var fm agentFrontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, fmt.Errorf("yaml parse: %w", err)
	}

	// family = filename minus .agent.md (Copilot) or .md (Claude)
	base := filepath.Base(absPath)
	family := strings.TrimSuffix(base, ".md")
	family = strings.TrimSuffix(family, ".agent")

	// Tools: accept either a YAML list ([a, b, 'becoming/*']) or a
	// comma-separated string ("Read, Edit, Write"). After parse,
	// normalize to []string.
	var tools []string
	switch v := fm.Tools.(type) {
	case []interface{}:
		for _, t := range v {
			if s, ok := t.(string); ok {
				if s = strings.TrimSpace(s); s != "" {
					tools = append(tools, s)
				}
			}
		}
	case string:
		for _, p := range strings.Split(v, ",") {
			if p = strings.TrimSpace(p); p != "" {
				tools = append(tools, p)
			}
		}
	case nil:
		// no tools declared → empty list
	default:
		// unknown shape; log and proceed with empty tools
		fmt.Fprintf(os.Stderr,
			"  WARN agent %s: tools field has unexpected shape %T\n",
			family, v)
	}

	// Handoffs: keep verbatim YAML for future use.
	var handoffsYAML string
	if fm.Handoffs != nil {
		if b, err := yaml.Marshal(fm.Handoffs); err == nil {
			handoffsYAML = string(b)
		}
	}

	return &AgentDoc{
		Family:      family,
		Description: strings.TrimSpace(fm.Description),
		Body:        strings.TrimSpace(body),
		Tools:       tools,
		Model:       strings.TrimSpace(fm.Model),
		ModelMatch:  strings.TrimSpace(fm.ModelMatch),
		Handoffs:    handoffsYAML,
	}, nil
}

// upsertAgent inserts/updates an agent and rebuilds its tool perms.
// The perm rebuild is "delete-then-insert" so reimports are idempotent
// and removed tools actually go away.
//
// Tool perm pattern (mirrors the Phase 1.5 stewards-explore seed):
//   ('agent', '*', 'deny')                     — explicit deny by default
//   ('agent', <tool>, 'allow') for each tool   — declared allow list
//   ('agent', 'skill', 'allow')                — so the agent can load skills
func upsertAgent(ctx context.Context, pool *pgxpool.Pool, a *AgentDoc) error {
	// Resolve model_match. Empty → '*' (the default variant). Per-model
	// variants like 'kimi-*' are upserted side-by-side with the default,
	// because the agents PK is (family, model_match).
	modelMatch := a.ModelMatch
	if modelMatch == "" {
		modelMatch = "*"
	}

	// 1. Upsert agents row. steps defaults to 8 (the Phase 1.5
	//    substrate default; agents.steps is NOT NULL). Live DB has
	//    a separate UPDATE that bumps non-watchman agents to 50 — see
	//    3c.3.1 trigger bugfixes. New variants land at 8 and inherit
	//    the bump on the next migration apply.
	if _, err := pool.Exec(ctx,
		`INSERT INTO stewards.agents
		    (family, model_match, description, mode, prompt,
		     temperature, top_p, response_format, steps)
		 VALUES ($1, $2, $3, 'primary', $4, NULL, NULL, NULL, 8)
		 ON CONFLICT (family, model_match) DO UPDATE
		 SET description = EXCLUDED.description,
		     mode        = EXCLUDED.mode,
		     prompt      = EXCLUDED.prompt`,
		a.Family, modelMatch, a.Description, a.Body,
	); err != nil {
		return fmt.Errorf("upsert agent %s/%s: %w", a.Family, modelMatch, err)
	}

	// 2. Tool perms are keyed by (agent_family, tool_pattern), not by
	//    model_match — they're shared across all variants of a family.
	//    So we ONLY rebuild perms when importing the default ('*') variant.
	//    Variant imports (e.g. kimi-*) leave the existing perms alone;
	//    they inherit whatever the default declared.
	if modelMatch != "*" {
		return nil
	}

	// 3. Clear existing FRONTMATTER tool perms for this family so removed
	//    tools actually disappear on reimport. Critically: do NOT touch
	//    rows where source='broadcast' or source='manual' — those are
	//    substrate-internal grants (e.g. 3c.2.5's blanket study_*: allow)
	//    that the importer doesn't own. See 3c3-3-agent-tool-perms-provenance.sql
	//    for the column rationale.
	if _, err := pool.Exec(ctx,
		`DELETE FROM stewards.agent_tool_perms
		  WHERE agent_family = $1 AND source = 'frontmatter'`,
		a.Family,
	); err != nil {
		return fmt.Errorf("clear tool perms %s: %w", a.Family, err)
	}

	// 4. Insert deny-* + allow-skill + one allow per declared tool.
	//    All rows from this path are tagged source='frontmatter'.
	rules := [][2]string{{"*", "deny"}, {"skill", "allow"}}
	for _, t := range a.Tools {
		// Skip 'skill' if the agent already declared it; we always
		// add allow-skill above, and dupe (family, pattern) would
		// hit the PK.
		if t == "skill" {
			continue
		}
		rules = append(rules, [2]string{t, "allow"})
	}
	for _, r := range rules {
		// ON CONFLICT updates `action` and `source` so a previously
		// broadcast/manual row that's now declared in frontmatter gets
		// re-tagged as frontmatter (and will be cleaned up properly on
		// the next import).
		if _, err := pool.Exec(ctx,
			`INSERT INTO stewards.agent_tool_perms
			    (agent_family, tool_pattern, action, source)
			 VALUES ($1, $2, $3, 'frontmatter')
			 ON CONFLICT (agent_family, tool_pattern) DO UPDATE
			 SET action = EXCLUDED.action,
			     source = 'frontmatter'`,
			a.Family, r[0], r[1],
		); err != nil {
			return fmt.Errorf("insert tool perm %s/%s: %w", a.Family, r[0], err)
		}
	}

	return nil
}

// ImportAgents walks a directory of *.agent.md (or .md) files,
// parses each, and upserts into stewards.agents + stewards.agent_tool_perms.
// Returns (ok, fail) counts.
func ImportAgents(ctx context.Context, pool *pgxpool.Pool,
	src Source, limit int, verbose bool,
) (int, int) {
	absRoot, err := filepath.Abs(src.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agents: resolve %s: %v\n", src.Path, err)
		return 0, 1
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agents: stat %s: %v\n", absRoot, err)
		return 0, 1
	}

	var files []string
	if !info.IsDir() {
		files = []string{absRoot}
	} else {
		err = filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(p), ".md") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "agents: walk %s: %v\n", absRoot, err)
			return 0, 1
		}
	}

	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}

	ok, fail := 0, 0
	for _, abs := range files {
		a, err := parseAgentMarkdown(abs)
		if err != nil {
			// Files without YAML frontmatter (README.md, NOTES.md,
			// etc.) get silently skipped rather than counted as a
			// failure — agent directories often include free-text docs
			// alongside the canonical *.agent.md files.
			if strings.Contains(err.Error(), "no YAML frontmatter") {
				if verbose {
					fmt.Printf("  skip: %s (no frontmatter)\n",
						filepath.Base(abs))
				}
				continue
			}
			fmt.Fprintf(os.Stderr, "  PARSE FAIL: %s: %v\n",
				filepath.Base(abs), err)
			fail++
			continue
		}
		if err := upsertAgent(ctx, pool, a); err != nil {
			fmt.Fprintf(os.Stderr, "  IMPORT FAIL: %s: %v\n", a.Family, err)
			fail++
			continue
		}
		if verbose {
			variant := a.ModelMatch
			if variant == "" {
				variant = "*"
			}
			fmt.Printf("  ok: agent %s/%s (%d tools)\n",
				a.Family, variant, len(a.Tools))
		}
		ok++
	}
	return ok, fail
}
