---
date: 2026-06-22
topic: llama-chip hub — a shared GPU-pool coordinator (llama.cpuchip.net), token-based join
lane: pg-ai-stewards
---

# The hub — turning federation into a platform

Same day, second arc. After the peer-federation shipped, Michael asked the natural next thing:
*"Could we extend this to run on llama.cpuchip.net? Grant keys to join a group, pool GPU resources,
give me an admin API key I can use to grant others join tokens… a platform where me and my
co-workers share compute would be AWESOME."* Plus: he's about to deploy a node off-LAN (work /
cpuchip.net), and a side-note to set up a shared email/SMS service.

## The council call: control plane vs data plane

The clarifying frame — and the thing that made this tractable — is that his ask splits into two
layers:

- **Control plane** (the hub): tokens, roster, UI. Small, hosted, fully buildable by me.
- **Data plane**: how inference bytes actually cross LANs. The real fork, and the part he hadn't
  named because it's not obvious.

I surfaced the fork with a recommendation and let him decide. He chose **NetBird self-host** for the
data plane (sovereign mesh, traffic peer-to-peer, never through the hub) and **llama.cpuchip.net,
tokens-now-Google-later** for hosting/auth. The NetBird choice *simplified* the hub: it's pure
control plane, zero relay code. New standing capability → ratified in council via those two answers
before I built (dominion_in_council satisfied).

## What shipped (`cpuchip/llama-chip` `3b82a32`, pushed)

- **`internal/hub`** — token store (JSON-file persisted, sha256 hashes, admin/node/user kinds,
  admin-key env seed) + in-memory TTL roster + HTTP server (bearer auth, admin-gated mint/revoke,
  register→roster) + an embedded admin UI (paste a key → live roster with GPU bars + token mgmt).
- **`cmd/llama-hub`** — the service binary + Dockerfile + docker-compose for Dokploy.
- **`internal/hubclient`** — the node side: heartbeats this node's models + GPU to the hub, applies
  the roster to the federation. A hub blip never clears the last-known roster (local-first holds).
- **fed hub-managed mode** — `config.federation.hub_url`/`hub_token`; `ApplyRoster` builds routes
  from the roster (self-excluded). Either discovery mode (static peers OR hub) feeds the same
  peer-to-peer routing I shipped this morning.
- **`docs/hub.md`** — architecture, token model, Dokploy deploy, co-worker onboarding, the
  private-intent scoping guardrail.

## Decisions worth keeping

- **Tokens are the right primitive, not OAuth.** GPU nodes are headless — they can't do a login
  flow. So API tokens cover the core (node join), and Google/email is a *human-UI* nicety to add
  later. This validated his "start with api token" instinct and de-risked P1.
- **Dependency-free on purpose.** Token store is a JSON file + mutex, roster is in-memory — keeps
  llama-hub a single pure-stdlib binary, matching llama-chip's "no Docker/Python/CGO" ethos. SQLite
  is a later swap if it ever needs scale.
- **The privacy guardrail, surfaced not buried.** A shared pool could route vivint to a co-worker's
  GPU. The token model reserves a `scope` field for pinning private intents to trusted nodes; P1
  doesn't *enforce* it yet, so the doc says plainly: keep private substrate work on a hub group of
  your own nodes until scoping lands. Named it rather than letting it surprise us.

## Oracle-first + live proof

`internal/hub` + `internal/hubclient` suites (mint/verify/revoke, plaintext-never-stored,
persistence across reload, admin gate, register→roster, TTL prune, node↔hub roster application with
self-exclusion). Then a live binary smoke: llama-hub booted, seeded its admin key from env, served
the UI, minted a node token, accepted a node registration (models + GPU → roster); bad token → 401,
node minting → 403. Live rig on :8090 untouched throughout (scratch port :18088, since removed).

## Carry-forward (his hands)

- **Deploy** llama.cpuchip.net on NOCIX Dokploy (compose `cmd/llama-hub/docker-compose.yml`,
  `LLAMA_HUB_ADMIN_KEY` secret, domain → 8088, `/data` volume). Then NetBird mesh for the off-LAN
  node. Runbooks: `docs/hub.md` + `docs/federation.md`.
- **Shared email/SMS service** — filed to general-workspace's inbox; pairs with the hub (emailing a
  "you've been granted compute" token).
- **Next hub slices (deferred):** enforced private-intent scoping (the `scope` field), Google/email
  human auth, federated placement (load-on-demand on a remote node).
