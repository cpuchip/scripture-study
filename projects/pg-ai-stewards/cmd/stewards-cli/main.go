// stewards-cli — Go CLI for the pg_ai_stewards extension.
//
// Cross-platform replacement for stewards.ps1 and import-studies.ps1.
// Designed to run identically on Windows dev machines and Linux
// hosting servers; no codepage handling required (pgx is unicode-clean).
//
// Usage:
//
//	stewards-cli import --source <kind>:<dir-or-file> [--source ...]
//	stewards-cli study show <slug> [--sim N] [--cites N] [--verse-chars N]
//	stewards-cli study list [--kind <kind>]
//	stewards-cli study refresh [<slug>]
//
// Connection: STEWARDS_DSN env var; defaults to local docker mapping.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/db"
	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/importer"
	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/show"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// On Windows, the default console code page is cp1252 even when
	// stdout is being piped, so em-dashes from UTF-8 strings render
	// as ΓÇö. Set the active code page to 65001 (UTF-8) for this
	// process's output. No-op on Linux/Mac where stdout is already UTF-8.
	configureUTF8Stdout()
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "import":
		runImport(ctx, os.Args[2:])
	case "study":
		runStudy(ctx, os.Args[2:])
	case "workstream", "ws":
		runWorkstream(ctx, os.Args[2:])
	case "edges":
		runEdges(ctx, os.Args[2:])
	case "todo":
		runTodo(ctx, os.Args[2:])
	case "context", "ctx":
		runContext(ctx, os.Args[2:])
	case "watchman":
		runWatchman(ctx, os.Args[2:])
	case "pipeline":
		runPipeline(ctx, os.Args[2:])
	case "work-item", "wi":
		runWorkItem(ctx, os.Args[2:])
	case "materialize-writes":
		runMaterializeWrites(ctx, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `stewards-cli — pg_ai_stewards CLI

Commands:
  import --source <kind>:<dir-or-file> [--source ...]
      Import documents/agents/skills into the substrate. May be
      repeated. Document kinds (study|doc|proposal|phase-doc|phase|
      journal) target stewards.studies; agents and skills target
      stewards.agents / stewards.skills.
      Examples:
          --source study:study
          --source doc:docs/work-with-ai
          --source proposal:.spec/proposals
          --source phase-doc:projects/pg-ai-stewards/phases.md
          --source journal:.spec/journal
          --source agent:.github/agents
          --source skill:.github/skills

  study show <slug> [--sim N] [--cites N] [--verse-chars N]
      Print a formatted view of a study + resolved citations + similar.

  study list [--kind <kind>]
      List all studies; optionally filtered by kind.

  study refresh [<slug>]
      Re-resolve citations + recompute similarity for one slug, or all.

  workstream list
      List all workstreams + count of declared proposals.

  workstream show <id>
      Show one workstream + its declared proposals (from graph).

  edges <slug>
      Show outbound declared-provenance edges for a slug (Phase 2.6a).

  todo create --parent <kind>:<slug> --title "..." [--body "..."] [--slug X] [--session SID]
      Create a todo attached to a parent vertex (Workstream|Study|Phase|Todo).

  todo done <id-or-slug> [--session SID] [--status done|dropped]
      Mark a todo done (or other terminal status).

  todo list [--parent <kind>:<slug>] [--status open|in_progress|done|dropped]
      List todos with optional parent and status filters.

  todo audit
      Run roll-up audit (parent done with open children, etc.).

  context <slug> [--depth N]
      Walk the graph neighborhood of a slug (Phase 2.6c). Depth
      clamped 1..4. Returns one row per (hop, direction, edge,
      neighbor) with closest-hop-wins dedup.

  watchman queue [--limit N]
  watchman verdict <slug> --status clean|drift|done|superseded|skipped
                          [--reasoning T] [--model M] [--pass-id P]
                          [--tokens-in N] [--tokens-out N]
  watchman finding <slug> --kind drift|synthesis --message T
                          [--severity low|medium|high]
                          [--suggested-action T] [--related slug,slug]
                          [--pass-id P]
  watchman ack <id> [--resolution acted|dismissed|deferred]
  watchman history <slug>
      Watchman substrate (Phase 2.7a). Inspect the dirty queue,
      record verdicts (which reset the dirty-bit) and findings
      (which suppress re-evaluation until acknowledged), and view
      the verdict + finding timeline for a doc.

  watchman pass [--limit N] [--provider P] [--model M] [--timeout S]
                [--max-input-chars N] [--dry-run] [--slug X]
      Phase 3a Go-orchestrator pass. CLI loops over slugs, polls
      work_queue per slug, parses JSON in Go, calls record_verdict.
      Useful for --slug single-doc repro and Go-side log visibility.

  watchman pass-now [--limit N] [--provider P] [--model M] [--budget T]
                    [--actor A] [--timeout S]
      Phase 2.7b.1 trigger-driven pass. SQL function enqueues N
      chats; the AFTER UPDATE trigger on work_queue records each
      verdict transactionally with the chat completion. CLI polls
      stewards.watchman_passes until the row reaches a terminal
      status. Faster, no race window, bgworker stays generic.

  watchman passes [--limit N]
      List recent passes (newest first) with rolled-up verdict counts.

  watchman pass-detail <pass-id>
      Show one pass + every verdict and finding it produced.

  watchman config show
  watchman config set [--schedule S] [--budget T] [--model M]
                      [--provider P] [--agent A] [--dirty-threshold N]
                      [--idle-hours N] [--enabled BOOL]
                      [--min-interval-hours N] [--preferred-dow N]
                      [--preferred-hour N] [--pass-limit N]
                      [--pressure-cooldown-hours N]
                      [--idle-cooldown-hours N]
      View / edit the singleton stewards.watchman_config row.
      Phase 2.7b.2 scheduler fields: --enabled, --min-interval-hours,
      --preferred-dow (-1=any, 0=Sun..6=Sat), --preferred-hour
      (-1=any, 0..23 UTC), --pass-limit, cooldown hours.

  watchman scheduler-status
      Print stewards.watchman_should_fire() decision + the live inputs
      feeding it (dirty count, hours since last pass, in-progress pass,
      etc.). Useful for "why isn't it firing?" debugging.

  watchman active-md
      Render a markdown status report from substrate state via
      stewards.regenerate_active_md(). Sections: In Flight by
      workstream, Open Findings, Open Todos, Recent Watchman
      Activity, Corpus Stats. Phase 2.7b.4.

  pipeline list | show <family>
      List or inspect pipeline definitions (Phase 3c.1).

  work-item create --pipeline P [--slug S] [--input JSON]
                   [--user-input "..."] [--actor A] [--budget T]
  work-item list [--pipeline P] [--status S]
  work-item show <id-or-slug>
  work-item dispatch <id-or-slug> [--user-input "..."]
  work-item advance <id-or-slug> [--output JSON]
  work-item cancel <id-or-slug> [--reason R]
      Phase 3c.1: pipeline instances flowing through stages. dispatch
      enqueues a chat for the current stage; advance records its
      output and transitions to the next stage (or marks completed).
      Auto-advance via trigger comes in Phase 3c.2.

Environment:
  STEWARDS_DSN    Postgres DSN (default: postgres://stewards:stewards@localhost:5432/stewards?sslmode=disable)
`)
}

// ---------- import ----------

type sourceFlag []importer.Source

func (s *sourceFlag) String() string {
	parts := make([]string, 0, len(*s))
	for _, src := range *s {
		parts = append(parts, src.Kind+":"+src.Path)
	}
	return strings.Join(parts, ",")
}

func (s *sourceFlag) Set(value string) error {
	idx := strings.Index(value, ":")
	if idx <= 0 || idx == len(value)-1 {
		return fmt.Errorf("--source must be <kind>:<path>, got %q", value)
	}
	*s = append(*s, importer.Source{Kind: value[:idx], Path: value[idx+1:]})
	return nil
}

func runImport(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	var sources sourceFlag
	fs.Var(&sources, "source", "kind:path (repeat for multiple)")
	limit := fs.Int("limit", 0, "max files per source (0 = no limit)")
	verbose := fs.Bool("v", false, "log each file as it imports")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if len(sources) == 0 {
		fmt.Fprintln(os.Stderr, "import: at least one --source required")
		os.Exit(1)
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	totalOK, totalFail := 0, 0
	for _, src := range sources {
		var ok, fail int
		// Dispatch by kind: agents/skills go to specialized importers
		// that target stewards.agents / stewards.skills tables; other
		// kinds (study, doc, proposal, journal, ...) go through
		// stewards.import_study() into stewards.studies.
		switch src.Kind {
		case "agent", "agents":
			ok, fail = importer.ImportAgents(ctx, pool, src, *limit, *verbose)
		case "skill", "skills":
			ok, fail = importer.ImportSkills(ctx, pool, src, *limit, *verbose)
		default:
			ok, fail = importer.ImportSource(ctx, pool, src, *limit, *verbose)
		}
		fmt.Printf("=== %s (%s): ok=%d fail=%d ===\n", src.Kind, src.Path, ok, fail)
		totalOK += ok
		totalFail += fail
	}
	fmt.Printf("\nTotal: ok=%d fail=%d\n", totalOK, totalFail)
	if totalFail > 0 {
		os.Exit(2)
	}
}

// ---------- study ----------

func runStudy(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "study: subcommand required (show|list|refresh)")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch args[0] {
	case "show":
		fs := flag.NewFlagSet("study show", flag.ExitOnError)
		sim := fs.Int("sim", 5, "similarity limit")
		cites := fs.Int("cites", 20, "citation limit")
		verseChars := fs.Int("verse-chars", 140, "max chars per verse line")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "study show: <slug> required")
			os.Exit(1)
		}
		if err := show.Study(ctx, pool, fs.Arg(0), *sim, *cites, *verseChars); err != nil {
			fmt.Fprintf(os.Stderr, "show: %v\n", err)
			os.Exit(1)
		}
	case "list":
		fs := flag.NewFlagSet("study list", flag.ExitOnError)
		kind := fs.String("kind", "", "filter by kind")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if err := show.List(ctx, pool, *kind); err != nil {
			fmt.Fprintf(os.Stderr, "list: %v\n", err)
			os.Exit(1)
		}
	case "refresh":
		slug := ""
		if len(args) > 1 {
			slug = args[1]
		}
		if err := show.Refresh(ctx, pool, slug); err != nil {
			fmt.Fprintf(os.Stderr, "refresh: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "study: unknown subcommand %q (show|list|refresh)\n", args[0])
		os.Exit(1)
	}
}

// ---------- workstream (Phase 2.6a) ----------

func runWorkstream(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "workstream: subcommand required (list|show)")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch args[0] {
	case "list":
		if err := show.WorkstreamList(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "list: %v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "workstream show: <id> required (e.g. WS5)")
			os.Exit(1)
		}
		if err := show.WorkstreamShow(ctx, pool, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "show: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "workstream: unknown subcommand %q (list|show)\n", args[0])
		os.Exit(1)
	}
}

// ---------- edges (Phase 2.6a) ----------

func runEdges(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "edges: <slug> required")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := show.DeclaredEdges(ctx, pool, args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "edges: %v\n", err)
		os.Exit(1)
	}
}

// ---------- todo (Phase 2.6b) ----------

func runTodo(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "todo: subcommand required (create|done|list|audit)")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("todo create", flag.ExitOnError)
		parent := fs.String("parent", "", "<kind>:<slug> (e.g. Workstream:WS5, Study:proposal-token-efficiency)")
		title := fs.String("title", "", "todo title (required)")
		body := fs.String("body", "", "todo body")
		slug := fs.String("slug", "", "optional human-friendly slug (must be unique)")
		session := fs.String("session", "", "creating session id (free-form)")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if *parent == "" || *title == "" {
			fmt.Fprintln(os.Stderr, "todo create: --parent and --title required")
			os.Exit(1)
		}
		idx := strings.Index(*parent, ":")
		if idx <= 0 || idx == len(*parent)-1 {
			fmt.Fprintf(os.Stderr, "todo create: --parent must be <kind>:<slug>, got %q\n", *parent)
			os.Exit(1)
		}
		if err := show.TodoCreate(ctx, pool, (*parent)[:idx], (*parent)[idx+1:], *title, *body, *slug, *session); err != nil {
			fmt.Fprintf(os.Stderr, "create: %v\n", err)
			os.Exit(1)
		}
	case "done":
		fs := flag.NewFlagSet("todo done", flag.ExitOnError)
		session := fs.String("session", "", "completing session id")
		status := fs.String("status", "done", "terminal status (done|dropped|in_progress|open)")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "todo done: <id-or-slug> required")
			os.Exit(1)
		}
		if err := show.TodoComplete(ctx, pool, fs.Arg(0), *session, *status); err != nil {
			fmt.Fprintf(os.Stderr, "done: %v\n", err)
			os.Exit(1)
		}
	case "list":
		fs := flag.NewFlagSet("todo list", flag.ExitOnError)
		parent := fs.String("parent", "", "<kind>:<slug> filter")
		status := fs.String("status", "", "status filter")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		pk, ps := "", ""
		if *parent != "" {
			idx := strings.Index(*parent, ":")
			if idx <= 0 || idx == len(*parent)-1 {
				fmt.Fprintf(os.Stderr, "todo list: --parent must be <kind>:<slug>, got %q\n", *parent)
				os.Exit(1)
			}
			pk, ps = (*parent)[:idx], (*parent)[idx+1:]
		}
		if err := show.TodoList(ctx, pool, pk, ps, *status); err != nil {
			fmt.Fprintf(os.Stderr, "list: %v\n", err)
			os.Exit(1)
		}
	case "audit":
		if err := show.TodoAudit(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "audit: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "todo: unknown subcommand %q (create|done|list|audit)\n", args[0])
		os.Exit(1)
	}
}

// ---------- context (Phase 2.6c) ----------

func runContext(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("context", flag.ExitOnError)
	depth := fs.Int("depth", 2, "hops to walk (clamped 1..4)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "context: <slug> required")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := show.Context(ctx, pool, fs.Arg(0), *depth); err != nil {
		fmt.Fprintf(os.Stderr, "context: %v\n", err)
		os.Exit(1)
	}
}

// ---------- watchman (Phase 2.7a) ----------

func runWatchman(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "watchman: subcommand required (queue|verdict|finding|ack|history)")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	sub := args[0]
	rest := args[1:]
	switch sub {
	case "queue":
		fs := flag.NewFlagSet("watchman queue", flag.ExitOnError)
		limit := fs.Int("limit", 50, "max rows")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if err := show.WatchmanQueue(ctx, pool, *limit); err != nil {
			fmt.Fprintf(os.Stderr, "watchman queue: %v\n", err)
			os.Exit(1)
		}
	case "verdict":
		fs := flag.NewFlagSet("watchman verdict", flag.ExitOnError)
		status := fs.String("status", "", "clean|drift|done|superseded|skipped (required)")
		reasoning := fs.String("reasoning", "", "why this verdict")
		model := fs.String("model", "", "model used (NULL for human)")
		passID := fs.String("pass-id", "", "groups verdicts in one pass")
		tokensIn := fs.Int("tokens-in", 0, "tokens consumed (input)")
		tokensOut := fs.Int("tokens-out", 0, "tokens consumed (output)")
		actor := fs.String("actor", "human", "actor recording the verdict")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 || *status == "" {
			fmt.Fprintln(os.Stderr, "watchman verdict: <slug> + --status required")
			os.Exit(1)
		}
		if err := show.WatchmanVerdict(ctx, pool, fs.Arg(0), *status,
			*reasoning, *model, *passID, *actor, *tokensIn, *tokensOut); err != nil {
			fmt.Fprintf(os.Stderr, "watchman verdict: %v\n", err)
			os.Exit(1)
		}
	case "finding":
		fs := flag.NewFlagSet("watchman finding", flag.ExitOnError)
		kind := fs.String("kind", "drift", "drift|synthesis")
		message := fs.String("message", "", "finding message (required)")
		severity := fs.String("severity", "medium", "low|medium|high")
		action := fs.String("suggested-action", "", "what to do about it")
		related := fs.String("related", "", "comma-separated related slugs")
		passID := fs.String("pass-id", "", "groups findings in one pass")
		actor := fs.String("actor", "human", "actor recording the finding")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 || *message == "" {
			fmt.Fprintln(os.Stderr, "watchman finding: <slug> + --message required")
			os.Exit(1)
		}
		var relatedSlugs []string
		if *related != "" {
			for _, s := range strings.Split(*related, ",") {
				if t := strings.TrimSpace(s); t != "" {
					relatedSlugs = append(relatedSlugs, t)
				}
			}
		}
		if err := show.WatchmanFinding(ctx, pool, fs.Arg(0), *kind, *message,
			*severity, *action, *passID, *actor, relatedSlugs); err != nil {
			fmt.Fprintf(os.Stderr, "watchman finding: %v\n", err)
			os.Exit(1)
		}
	case "ack":
		fs := flag.NewFlagSet("watchman ack", flag.ExitOnError)
		resolution := fs.String("resolution", "acted", "acted|dismissed|deferred")
		actor := fs.String("actor", "human", "actor acknowledging")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "watchman ack: <finding-id> required")
			os.Exit(1)
		}
		id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "watchman ack: invalid id %q\n", fs.Arg(0))
			os.Exit(1)
		}
		if err := show.WatchmanAcknowledge(ctx, pool, id, *resolution, *actor); err != nil {
			fmt.Fprintf(os.Stderr, "watchman ack: %v\n", err)
			os.Exit(1)
		}
	case "history":
		fs := flag.NewFlagSet("watchman history", flag.ExitOnError)
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "watchman history: <slug> required")
			os.Exit(1)
		}
		if err := show.WatchmanHistory(ctx, pool, fs.Arg(0)); err != nil {
			fmt.Fprintf(os.Stderr, "watchman history: %v\n", err)
			os.Exit(1)
		}
	case "pass":
		// Phase 3a: model-driven consolidation pass.
		// Defaults to opencode_go + kimi-k2.6 (proven cheap+fast path
		// from Phase 1.6). Local alternative: --provider lm_studio
		// --model qwen/qwen3.6-27b.
		fs := flag.NewFlagSet("watchman pass", flag.ExitOnError)
		provider := fs.String("provider", "opencode_go", "lm_studio|opencode_go|ollama")
		model := fs.String("model", "kimi-k2.6", "model id (provider-specific)")
		agentFamily := fs.String("agent", "watchman-consolidator", "agent family from stewards.agents")
		limit := fs.Int("limit", 5, "max docs to consolidate this pass")
		timeoutSec := fs.Int("timeout", 660, "per-item poll timeout in seconds (bgworker chat default is 600s; raise both for very large inputs or use --max-input-chars)")
		dryRun := fs.Bool("dry-run", false, "print verdicts but do NOT call record_verdict/record_finding")
		slug := fs.String("slug", "", "if set, run only this slug (bypasses dirty_queue) — useful for repro")
		maxInputChars := fs.Int("max-input-chars", 0, "if >0, truncate doc input to this many chars (head/tail with elision marker). 0 = no limit. Use ~30000 for big docs to fit chat timeout.")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if err := show.WatchmanPass(ctx, pool,
			*provider, *model, *agentFamily,
			*limit, time.Duration(*timeoutSec)*time.Second,
			*dryRun, *slug, *maxInputChars,
		); err != nil {
			fmt.Fprintf(os.Stderr, "watchman pass: %v\n", err)
			os.Exit(1)
		}
	case "pass-now":
		// Phase 2.7b.1 trigger-driven pass. Calls
		// stewards.watchman_pass_start, polls watchman_passes until
		// terminal. Defaults are NULL so the SQL function falls
		// through to watchman_config defaults; pass an explicit
		// flag value to override.
		fs := flag.NewFlagSet("watchman pass-now", flag.ExitOnError)
		limit := fs.Int("limit", 5, "max docs to consolidate this pass")
		provider := fs.String("provider", "", "override default_provider (empty = use config)")
		model := fs.String("model", "", "override default_model (empty = use config)")
		agent := fs.String("agent", "", "override default_agent_family (empty = use config)")
		actor := fs.String("actor", "watchman", "actor recording verdicts (audit trail)")
		budget := fs.Int("budget", 0, "override token_budget (0 = use config)")
		timeoutSec := fs.Int("timeout", 1200, "max seconds to wait for the whole pass")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if err := show.WatchmanPassNow(ctx, pool,
			*limit, *provider, *model, *agent, *actor,
			*budget, time.Duration(*timeoutSec)*time.Second,
		); err != nil {
			fmt.Fprintf(os.Stderr, "watchman pass-now: %v\n", err)
			os.Exit(1)
		}
	case "passes":
		fs := flag.NewFlagSet("watchman passes", flag.ExitOnError)
		limit := fs.Int("limit", 20, "max passes to list (newest first)")
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if err := show.WatchmanPasses(ctx, pool, *limit); err != nil {
			fmt.Fprintf(os.Stderr, "watchman passes: %v\n", err)
			os.Exit(1)
		}
	case "pass-detail":
		fs := flag.NewFlagSet("watchman pass-detail", flag.ExitOnError)
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "watchman pass-detail: <pass-id> required")
			os.Exit(1)
		}
		if err := show.WatchmanPassDetail(ctx, pool, fs.Arg(0)); err != nil {
			fmt.Fprintf(os.Stderr, "watchman pass-detail: %v\n", err)
			os.Exit(1)
		}
	case "config":
		// `watchman config show` | `watchman config set --field ...`
		if len(rest) == 0 {
			fmt.Fprintln(os.Stderr, "watchman config: subcommand required (show|set)")
			os.Exit(1)
		}
		switch rest[0] {
		case "show":
			if err := show.WatchmanConfigShow(ctx, pool); err != nil {
				fmt.Fprintf(os.Stderr, "watchman config show: %v\n", err)
				os.Exit(1)
			}
		case "set":
			fs := flag.NewFlagSet("watchman config set", flag.ExitOnError)
			schedule := fs.String("schedule", "", "schedule_cron (human label)")
			provider := fs.String("provider", "", "default_provider")
			model := fs.String("model", "", "default_model")
			agent := fs.String("agent", "", "default_agent_family")
			budget := fs.Int("budget", 0, "token_budget")
			dirtyThreshold := fs.Int("dirty-threshold", 0, "dirty_threshold (pressure trigger)")
			idleHours := fs.Int("idle-hours", 0, "idle_threshold_hours (0=off)")
			// Phase 2.7b.2 scheduler fields
			enabled := fs.Bool("enabled", false, "schedule_enabled (master switch)")
			minInterval := fs.Int("min-interval-hours", 0, "schedule_min_interval_hours (cron gap)")
			preferredDOW := fs.Int("preferred-dow", -1, "schedule_preferred_dow_utc (0=Sun..6=Sat, -1=any)")
			preferredHour := fs.Int("preferred-hour", -1, "schedule_preferred_hour_utc (0..23, -1=any)")
			passLimit := fs.Int("pass-limit", 0, "schedule_pass_limit")
			pressureCooldown := fs.Int("pressure-cooldown-hours", 0, "schedule_pressure_cooldown_hours")
			idleCooldown := fs.Int("idle-cooldown-hours", 0, "schedule_idle_cooldown_hours")
			if err := fs.Parse(rest[1:]); err != nil {
				os.Exit(1)
			}
			seen := map[string]bool{}
			fs.Visit(func(f *flag.Flag) { seen[f.Name] = true })
			fields := []show.WatchmanConfigSetField{}
			if seen["schedule"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_cron", Value: *schedule})
			}
			if seen["provider"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "default_provider", Value: *provider})
			}
			if seen["model"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "default_model", Value: *model})
			}
			if seen["agent"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "default_agent_family", Value: *agent})
			}
			if seen["budget"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "token_budget", Value: *budget})
			}
			if seen["dirty-threshold"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "dirty_threshold", Value: *dirtyThreshold})
			}
			if seen["idle-hours"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "idle_threshold_hours", Value: *idleHours})
			}
			if seen["enabled"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_enabled", Value: *enabled})
			}
			if seen["min-interval-hours"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_min_interval_hours", Value: *minInterval})
			}
			if seen["preferred-dow"] {
				// -1 → NULL (any day); 0..6 → that day; outside range → error
				if *preferredDOW == -1 {
					fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_preferred_dow_utc", Value: nil})
				} else if *preferredDOW >= 0 && *preferredDOW <= 6 {
					fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_preferred_dow_utc", Value: *preferredDOW})
				} else {
					fmt.Fprintf(os.Stderr, "watchman config set: --preferred-dow must be -1..6, got %d\n", *preferredDOW)
					os.Exit(1)
				}
			}
			if seen["preferred-hour"] {
				if *preferredHour == -1 {
					fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_preferred_hour_utc", Value: nil})
				} else if *preferredHour >= 0 && *preferredHour <= 23 {
					fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_preferred_hour_utc", Value: *preferredHour})
				} else {
					fmt.Fprintf(os.Stderr, "watchman config set: --preferred-hour must be -1..23, got %d\n", *preferredHour)
					os.Exit(1)
				}
			}
			if seen["pass-limit"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_pass_limit", Value: *passLimit})
			}
			if seen["pressure-cooldown-hours"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_pressure_cooldown_hours", Value: *pressureCooldown})
			}
			if seen["idle-cooldown-hours"] {
				fields = append(fields, show.WatchmanConfigSetField{Column: "schedule_idle_cooldown_hours", Value: *idleCooldown})
			}
			if err := show.WatchmanConfigSet(ctx, pool, fields); err != nil {
				fmt.Fprintf(os.Stderr, "watchman config set: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "watchman config: unknown subcommand %q (show|set)\n", rest[0])
			os.Exit(1)
		}
	case "scheduler-status":
		// Phase 2.7b.2 — print should_fire() decision + the inputs
		// feeding it. Useful when "the scheduler isn't firing, why?"
		fs := flag.NewFlagSet("watchman scheduler-status", flag.ExitOnError)
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if err := show.WatchmanSchedulerStatus(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "watchman scheduler-status: %v\n", err)
			os.Exit(1)
		}
	case "active-md":
		// Phase 2.7b.4 — print the substrate-derived active.md report.
		fs := flag.NewFlagSet("watchman active-md", flag.ExitOnError)
		if err := fs.Parse(rest); err != nil {
			os.Exit(1)
		}
		if err := show.WatchmanActiveMD(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "watchman active-md: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "watchman: unknown subcommand %q\n", sub)
		os.Exit(1)
	}
}

// ---------- pipeline (Phase 3c.1) ----------

func runPipeline(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "pipeline: subcommand required (list|show)")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch args[0] {
	case "list":
		if err := show.PipelineList(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "pipeline list: %v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "pipeline show: <family> required")
			os.Exit(1)
		}
		if err := show.PipelineShow(ctx, pool, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "pipeline show: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "pipeline: unknown subcommand %q (list|show)\n", args[0])
		os.Exit(1)
	}
}

// ---------- work-item (Phase 3c.1) ----------

func runWorkItem(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "work-item: subcommand required (create|list|show|dispatch|advance|cancel)")
		os.Exit(1)
	}
	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("work-item create", flag.ExitOnError)
		pipeline := fs.String("pipeline", "", "pipeline family (required)")
		slug := fs.String("slug", "", "optional human-readable slug (must be unique)")
		actor := fs.String("actor", "human", "who initiated this work_item")
		inputJSON := fs.String("input", "{}", "input as JSON object")
		userInput := fs.String("user-input", "", "shorthand for --input '{\"user_input\":\"...\"}'")
		budget := fs.Int("budget", 0, "token budget across all stages (0 = unbounded)")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if *pipeline == "" {
			fmt.Fprintln(os.Stderr, "work-item create: --pipeline required")
			os.Exit(1)
		}
		input := *inputJSON
		if *userInput != "" {
			b, err := json.Marshal(map[string]any{"user_input": *userInput})
			if err != nil {
				fmt.Fprintf(os.Stderr, "work-item create: encode user_input: %v\n", err)
				os.Exit(1)
			}
			input = string(b)
		}
		if err := show.WorkItemCreate(ctx, pool, *pipeline, *slug, *actor,
			[]byte(input), *budget); err != nil {
			fmt.Fprintf(os.Stderr, "work-item create: %v\n", err)
			os.Exit(1)
		}
	case "list":
		fs := flag.NewFlagSet("work-item list", flag.ExitOnError)
		pipeline := fs.String("pipeline", "", "filter by pipeline family")
		status := fs.String("status", "", "filter by status")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if err := show.WorkItemList(ctx, pool, *pipeline, *status); err != nil {
			fmt.Fprintf(os.Stderr, "work-item list: %v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "work-item show: <id-or-slug> required")
			os.Exit(1)
		}
		if err := show.WorkItemShow(ctx, pool, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "work-item show: %v\n", err)
			os.Exit(1)
		}
	case "dispatch":
		fs := flag.NewFlagSet("work-item dispatch", flag.ExitOnError)
		userInput := fs.String("user-input", "", "override user input for the current stage")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "work-item dispatch: <id-or-slug> required")
			os.Exit(1)
		}
		if err := show.WorkItemDispatch(ctx, pool, fs.Arg(0), *userInput); err != nil {
			fmt.Fprintf(os.Stderr, "work-item dispatch: %v\n", err)
			os.Exit(1)
		}
	case "advance":
		fs := flag.NewFlagSet("work-item advance", flag.ExitOnError)
		output := fs.String("output", "{}", "stage output as JSON object")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "work-item advance: <id-or-slug> required")
			os.Exit(1)
		}
		if err := show.WorkItemAdvance(ctx, pool, fs.Arg(0), []byte(*output)); err != nil {
			fmt.Fprintf(os.Stderr, "work-item advance: %v\n", err)
			os.Exit(1)
		}
	case "cancel":
		fs := flag.NewFlagSet("work-item cancel", flag.ExitOnError)
		reason := fs.String("reason", "", "optional cancellation reason")
		if err := fs.Parse(args[1:]); err != nil {
			os.Exit(1)
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "work-item cancel: <id-or-slug> required")
			os.Exit(1)
		}
		if err := show.WorkItemCancel(ctx, pool, fs.Arg(0), *reason); err != nil {
			fmt.Fprintf(os.Stderr, "work-item cancel: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "work-item: unknown subcommand %q\n", args[0])
		os.Exit(1)
	}
}

// ---------- materialize-writes (Batch G.4.2) ----------
//
// Consumer for stewards.pending_file_writes. Drains unmaterialized rows
// and performs the file write (append or create). Idempotent — already-
// materialized rows are skipped.
//
// Usage:
//
//	stewards-cli materialize-writes [--dry-run] [--limit N] [--repo-root PATH]
//
// target_path on each pending row is resolved against --repo-root (default:
// current working directory). 'create' mode refuses to overwrite an existing
// file; 'append' mode creates if missing.

func runMaterializeWrites(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("materialize-writes", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "print what would be written without touching the filesystem")
	limit := fs.Int("limit", 100, "max rows to process per invocation")
	repoRoot := fs.String("repo-root", ".", "directory target_path values are resolved against")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	rootAbs, err := filepath.Abs(*repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "materialize-writes: --repo-root: %v\n", err)
		os.Exit(1)
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, `
		SELECT id, requested_at, requested_by, target_path, write_mode,
		       content, coalesce(source_kind,''), coalesce(source_id,'')
		  FROM stewards.pending_file_writes
		 WHERE materialized_at IS NULL
		 ORDER BY requested_at ASC
		 LIMIT $1`, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "materialize-writes query: %v\n", err)
		os.Exit(1)
	}
	type pending struct {
		id          int64
		requestedAt time.Time
		requestedBy string
		targetPath  string
		writeMode   string
		content     string
		sourceKind  string
		sourceID    string
	}
	var batch []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.requestedAt, &p.requestedBy, &p.targetPath,
			&p.writeMode, &p.content, &p.sourceKind, &p.sourceID); err != nil {
			fmt.Fprintf(os.Stderr, "scan: %v\n", err)
			continue
		}
		batch = append(batch, p)
	}
	rows.Close()

	if len(batch) == 0 {
		fmt.Println("materialize-writes: nothing pending")
		return
	}

	var ok, failed, skipped int
	for _, p := range batch {
		// Resolve target relative to repo-root. Refuse paths that
		// escape the root (no leading "/", no "..").
		clean := filepath.Clean(p.targetPath)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			recordError(ctx, pool, p.id, fmt.Sprintf("path escape: %s", p.targetPath))
			fmt.Fprintf(os.Stderr, "skip #%d: path escape: %s\n", p.id, p.targetPath)
			failed++
			continue
		}
		full := filepath.Join(rootAbs, clean)

		// Tag with dry-run
		if *dryRun {
			info := fmt.Sprintf("DRY-RUN #%d (%s, %s) → %s [%d bytes]",
				p.id, p.requestedBy, p.writeMode, full, len(p.content))
			fmt.Println(info)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			recordError(ctx, pool, p.id, "mkdir: "+err.Error())
			fmt.Fprintf(os.Stderr, "skip #%d: mkdir %s: %v\n", p.id, filepath.Dir(full), err)
			failed++
			continue
		}

		switch p.writeMode {
		case "create":
			// Refuse to overwrite. If the file exists, mark skipped.
			if _, err := os.Stat(full); err == nil {
				recordError(ctx, pool, p.id, "create: file exists; refusing to overwrite")
				fmt.Fprintf(os.Stderr, "skip #%d: %s already exists (create mode won't overwrite)\n", p.id, full)
				skipped++
				continue
			}
			if err := os.WriteFile(full, []byte(p.content), 0o644); err != nil {
				recordError(ctx, pool, p.id, "write: "+err.Error())
				fmt.Fprintf(os.Stderr, "fail #%d: write %s: %v\n", p.id, full, err)
				failed++
				continue
			}
		case "append":
			f, err := os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				recordError(ctx, pool, p.id, "open: "+err.Error())
				fmt.Fprintf(os.Stderr, "fail #%d: open %s: %v\n", p.id, full, err)
				failed++
				continue
			}
			if _, err := f.WriteString(p.content); err != nil {
				f.Close()
				recordError(ctx, pool, p.id, "append: "+err.Error())
				fmt.Fprintf(os.Stderr, "fail #%d: append %s: %v\n", p.id, full, err)
				failed++
				continue
			}
			f.Close()
		default:
			recordError(ctx, pool, p.id, "unknown write_mode: "+p.writeMode)
			fmt.Fprintf(os.Stderr, "fail #%d: unknown write_mode %q\n", p.id, p.writeMode)
			failed++
			continue
		}

		// Mark materialized
		_, err := pool.Exec(ctx,
			`UPDATE stewards.pending_file_writes
			    SET materialized_at = now(), materialized_by = 'cli'
			  WHERE id = $1`, p.id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn #%d: wrote file but UPDATE failed: %v\n", p.id, err)
			// File is written but DB not updated; next run will re-attempt
			// the create which will then fail with "file exists." That's
			// fine — manual intervention then.
		}

		// If sourceKind='work_item', also update the studies row's
		// file_path if the work_item promoted to one. The promotion
		// path inserts into stewards.studies; we set the file_path
		// here once the file actually exists.
		if p.sourceKind == "work_item" && strings.HasSuffix(p.targetPath, ".md") {
			_, _ = pool.Exec(ctx,
				`UPDATE stewards.studies
				    SET file_path = $1
				  WHERE slug = 'substrate--' || (
				          SELECT slug FROM stewards.work_items WHERE id = $2::uuid
				        )`, p.targetPath, p.sourceID)
		}

		fmt.Printf("ok #%d (%s, %s) → %s\n", p.id, p.requestedBy, p.writeMode, full)
		ok++
	}

	fmt.Printf("\nmaterialize-writes: ok=%d skipped=%d failed=%d (total=%d)\n",
		ok, skipped, failed, len(batch))
	if failed > 0 {
		os.Exit(1)
	}
}

func recordError(ctx context.Context, pool *pgxpool.Pool, id int64, msg string) {
	// best-effort log into materialized_by; don't fail if this fails
	_, _ = pool.Exec(ctx,
		`UPDATE stewards.pending_file_writes
		    SET materialized_by = 'error:' || $1
		  WHERE id = $2 AND materialized_at IS NULL`,
		msg, id)
}
