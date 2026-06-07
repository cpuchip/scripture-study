Dave,

I had my AI do a deep read of your workflow framework — the procedures / agents / skills
repo — and compare it against the system I've been building (a Postgres-backed agent
substrate plus a pile of study and dev workflows). I want to tell you what we found,
because it was more interesting than I expected.

## The big thing: you arrived at the same hard lessons independently

Not the obvious stuff — the subtle stuff:

- Your **Agent Task Suitability** note in `DESIGN.md` — cold-start analytical subagents
  produce "structurally correct but factually invented" output, so prefer inline review by
  the context-holder — is exactly a lesson I learned the painful way. The full-context
  orchestrator catches what a freshly-spawned reviewer fabricates. You named it cleaner
  than I had.
- Your **invariant-traceability** rule in `Plan.md` — every invariant must trace to a
  work-item statement, with that invented-concurrency-constraint as the bad example — is
  the same discipline I enforce against invented quotes. You forbid invented *constraints*
  the way I forbid invented *citations*. Same anti-confabulation instinct, different
  artifact.
- Your **PlanReview** gate — external, tiered findings, revise-up-to-three-then-escalate,
  "uncertain defaults to Blocking when no human is available" — matches my critic stage
  almost beat for beat. So does "Reviewer fixes minor issues directly."
- Even "NEVER narrate yourself" and your Reflect → specific-recommendations loop line up
  with rules I'd written separately.

Two people building from different starting points and landing on the same shape — that's
the strongest evidence I've seen that the pattern is real, and not just something I talked
myself into.

## What I'm borrowing from you (with credit)

1. The explicit **"AI Freedom"** section in your plans — naming what's *intentionally*
   unconstrained. I over-specify; yours is better.
2. **Invariant-traceability** as a formal plan-section rule.
3. Your **SideQuest** lightweight lane. I had the judgment for "this doesn't need the full
   pipeline" but never named the lane.
4. The **ODD / SRE depth** in your debug skill (observability-driven, high-cardinality
   search, compare a good trace against a bad one). My debug methodology is thinner on
   production diagnostics.
5. **"File-first / avoid double-dipping"** — don't explain in chat what belongs in the doc.
   Crisp, and I needed the reminder.

## A few things that might help you back

1. **Your repo was library-only.** The `internal/markdown` package and tests are there, but
   there's no `main.go` and no MCP SDK — so `go build -o md-mcp .` had nothing to build. I
   had my AI build the server entrypoint (wiring all your tools over the official
   `modelcontextprotocol/go-sdk`) plus three new tools — `md-section-append`,
   `md-section-move`, `md-frontmatter-set` — with tests, and opened a PR:
   **github.com/happydave/md-mcp#1**. Take it or leave it, but now it runs, and I'm using it.
2. **Make your cold-start caution enforceable.** You tell reviewers to "be specific." I'd go
   one step further: require every finding to cite a real `file:line`. An ungrounded review
   *can't* cite real lines — so a citation-less finding is a fabrication smell you can
   actually gate on, instead of trusting the reviewer to behave. I'm building that as a hard
   check.
3. **The runtime-vs-document split.** Your framework is meta-instructions an actor follows;
   mine also *runs* — the gates are enforced in code, so an agent can't quietly skip
   plan-review under pressure. Your risk is instruction-drift (the actor cuts a corner when
   it's tired or late); a thin runtime closes it. Happy to show you what that looks like if
   it's useful.

For full transparency: this came out of a faith-framed project — I build these workflows
partly as a way of practicing stewardship and council patterns I care about. What struck me
is that you got to the same place optimizing purely for engineering quality. The discipline
seems to be the discipline, wherever you come at it from.

Thanks for putting the framework out there. It sharpened mine.

— Michael
