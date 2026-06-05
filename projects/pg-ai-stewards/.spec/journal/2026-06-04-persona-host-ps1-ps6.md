# persona-host PS.1–PS.6 — the substrate's persona sidecar, built end-to-end

**Date:** 2026-06-04 (cont.)
**Lane:** AX2 / ai-chattermax persona concept #6. Build: mine (substrate-adjacent + security-sensitive, not coder-reachable).
**Outcome:** ✅ COMPLETE. `cmd/persona-host`, 6 gated commits `fd13f56`→`f529242` (not pushed).

## What was built

An **optional Go sidecar** — `cmd/persona-host` — exactly parallel to `cmd/coder-mcp`. The substrate core extension is untouched: persona identity, JWT minting, and the room handshake live in the sidecar, so a general `pg-ai-stewards` install never runs this. One host serves many personas.

- **PS.1** skeleton + `persona_host` schema (personas / persona_rooms / token_issuance / signing_key) applied idempotently on boot.
- **PS.2** Ed25519 keypair generated on first boot, persisted (race-safe `ON CONFLICT DO NOTHING` + re-select), published at `GET /pubkey`. Private key never logged — only a public sha256 fingerprint.
- **PS.3** `MintToken` (EdDSA JWT via golang-jwt v5): iss=pg-ai-stewards, sub=persona id, slug/name/avatar/room, ~15m exp, DB-minted jti recorded in `token_issuance`. `VerifyToken` is the reference impl ai-chattermax mirrors.
- **PS.4** registry list + seed `dm-assistant` / `npc-ally` on the substrate's real `fiction` agent family. `GET /personas` roster.
- **PS.5** `POST /join` handshake (resolve persona → upsert `persona_rooms` → mint) + a README documenting the full connection contract for the verify side.
- **PS.6** security gate: `-verify-token` ops mode; live end-to-end over real HTTP; tampered token rejected; log-leak check.

## Verification (the discipline held)

- **9 Go tests**, every one with the inverse hypothesis baked in: round-trip, **wrong-key rejected, expired rejected, tampered rejected**, /pubkey never leaks private material, unknown-persona → 404.
- **Per-phase smoke** against dev PG before each commit (C–F cadence), each independently confirmed via psql.
- **Live e2e** (PS.6) — the strongest "verify via the real path": started the real server, drove healthz → /pubkey → /personas → /join over HTTP, then verified the minted token against the **fetched** pubkey (not the in-memory key). Tampered token rejected on the live path. **Log-leak grep PASS**: no "PRIVATE KEY", no raw token, no PEM markers in stderr — only a public fingerprint.

## Decisions / notes for next time

- **Sidecar surface = lean HTTP + a `-verify-token` admin/ops mode.** No MCP server (it isn't bridge-spawned; ai-chattermax talks to it directly).
- **DSN secret handling:** the binary reads `STEWARDS_DSN`; for host smokes the dev password is sourced into `PGPASSWORD` via `docker exec printenv` (never echoed). pgx DSN *parse* errors are deliberately not wrapped (they echo the password) — return a redacted error.
- **pgx multi-statement migration:** `pool.Exec(ctx, schemaSQL)` with no args uses the simple protocol → the whole script applies in one call.
- **go.work:** added the module (sibling of the other `cmd/*`); built with `GOWORK=off` (the root go.work shadows otherwise).
- **Roster hygiene:** the early smokes created a `smoke-persona` that showed in `/personas`; reordered the smoke to mint for a seeded persona and cleaned the leftover from dev (scoped delete).

## Carry-forward

- **#7 the turn loop** is the next big piece and is NOT in #6: a long-lived per-(persona, room) connection that self-paces within the room ceiling and dispatches each turn to the substrate (`consult_subagent` / a work_item) for cognition. This is where the sidecar meets the substrate's dispatch.
- **Deployment** (open): its own container next to the substrate; a **scoped DB role** for `persona_host` (not superuser); ai-chattermax configured with the pubkey (env `STEWARDS_PERSONA_PUBKEY` or fetch `/pubkey`).
- **ai-chattermax side:** AX3-2 (#123, WS sender/de-dup) is the nearer-term chat-product fix; the persona verify-side wiring lands when #7 connects them.
