# Codewright goes live in chat + repo-awareness (Layer A)

**Date:** 2026-06-09 (evening, same Fable-5 day) · **Mode:** dev
**Trigger:** Michael minted a `chattercode` persona key + granted rooms, handed it to me
to wire. Then live-tested and surfaced a real gap.

## Wiring codewright into the live persona-host

The room-join I'd flagged as "Michael's ops step" — he did the Hinge (mint + grant), I
did the mechanical wiring:
1. Seeded `persona_host.personas`: slug=`chattercode`, display=`Codewright`,
   agent_family=`codewright`, **pipeline=`persona-turn-code`** (the crux — without the
   right pipeline it'd run the default chatter, not the tool-using code persona),
   + a character brief.
2. Appended `chattercode=<key>@<room>` to `CHATTERMAX_PERSONAS` in the extension `.env`
   (the slug→key map persona-host reads; the @room suffix is tolerated/ignored — the
   persona joins all rooms its key grants).
3. `docker compose up -d --force-recreate persona-host` → all 3 personas reconnected
   (pg-starlet, chip-assistant, **chattercode**); codewright joined 10-forward,
   Engineering, Library on chat.ibeco.me.

**The slug→pipeline mapping lives in the persona-host's LOCAL `persona_host` schema, not
the ai-chattermax platform** — minting the key on chat.ibeco.me is necessary but not
sufficient; the host needs its own persona row to know which pipeline drives the turns.

## Live feedback → Layer A (b92e805)

Michael chatted with it in Engineering: "it can't really see what it has access to. did
we design it to be fully open any public repo?" Investigation:
- **Not open — the opposite.** `CODER_REPO_ALLOWLIST` was unset → defaulted to
  `github.com/cpuchip/ai-chattermax` ONLY. One repo.
- **Couldn't see its scope** because the allow-list was a bridge env var, invisible to
  the model and absent from its prompt.

Built Layer A of the roadmap (below):
- **`list_repos` tool** (stewards-mcp, mcp_proxy) reads the SAME
  `CODER_REPO_ALLOWLIST`/`CODER_REPO_DENYLIST` the coder sandbox enforces, so what the
  persona reports == what it can actually clone. codewright granted it + re-prompted.
- **★ deny-beats-allow** in coder-mcp `repoAllowed` (new `CODER_REPO_DENYLIST` check
  first). Load-bearing: the bridge clones with a `GITHUB_TOKEN` that CAN reach PRIVATE
  repos, so a broad allow substring (`github.com/cpuchip/`) would otherwise expose
  `private-study` (Michael's job search). Scope set: **allow `github.com/cpuchip/`
  (all public repos), deny `private-study`.** Future private repos → denylist.
- r14 (register + grant + re-prompt) + bridge rebuilt (list_repos + denylist) →
  refresh-tools = 31 tools cached.

**Proven live:** codewright asked "what repos can you look at?" → called list_repos →
"I can look at any public repo under github.com/cpuchip/. The only thing off-limits is
anything matching private-study." ~10s, no fabrication, exact scope.

## Roadmap captured (codewright-workspace-roadmap.md, design-only)

Michael's A→B→C vision:
- **A — repo awareness (DONE):** honest about its reach.
- **B — work inside a repo+env:** clone/build/test/edit. Already 80% built (the coder
  MCP + code-pr cascade). Rec: **B v1 = codewright dispatches a code-pr work_item from
  chat** (reuse, don't duplicate; human Hinge on merge). The persistent-writable-
  workspace is a separate, later optimization — and a security INVERSION (kimi reading
  untrusted repos in a standing writable container = the injection profile the
  Google-MCP vet flagged). Flagged, not built.
- **C — orchestrate other model-CLIs:** `agy -p` / `opencode` / `claude -p` in a
  container, codewright as conductor (the [[project_council_review_beats_gift_matching]]
  idea made concrete: many doers + one critic/composer). **The crux Michael named is
  AUTH** — solved by a **long-lived container with a pre-provisioned secrets mount**
  (env + credential files injected at run, never in the image), plus cost governance
  across THREE billing pools (Anthropic credits, opencode sub, Gemini). Its own ratified
  project + security pass.

## Commits / state
Root `b92e805` (Layer A: sandbox.go denylist + heavyweight_tools list_repos + r14 +
roadmap spec) UNPUSHED. persona_host row + `.env` edit are live/local (gitignored).
Bridge image rebuilt twice this session (now carries list_repos + denylist). Soak
stayed running (no pg rebuild). Spend $1.96/$12.

## Carry-forward
- **Michael:** push root; try codewright live again now that it knows its repos
  (the fix is live); decide B v1 (code-pr-from-chat) when ready; C is its own project.
- Local `bin/stewards-mcp.exe` couldn't rebuild (Claude Code holds it open) — list_repos
  reaches Claude Code on my next reconnect; the BRIDGE binary (the one codewright uses)
  is fresh.
- Smoke work_items left (codewright-*-test, ct2ab-*) — harmless historical records.
