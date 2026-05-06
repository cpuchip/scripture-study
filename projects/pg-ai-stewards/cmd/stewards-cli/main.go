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
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/db"
	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/importer"
	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/show"
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
      Import documents into stewards.studies. May be repeated.
      Examples:
          --source study:study
          --source doc:docs/work-with-ai
          --source proposal:.spec/proposals
          --source phase-doc:projects/pg-ai-stewards/phases.md
          --source journal:.spec/journal

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
      the verdict + finding timeline for a doc. The bgworker that
      automates this is Phase 2.7b.

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
		ok, fail := importer.ImportSource(ctx, pool, src, *limit, *verbose)
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
	default:
		fmt.Fprintf(os.Stderr, "watchman: unknown subcommand %q\n", sub)
		os.Exit(1)
	}
}
