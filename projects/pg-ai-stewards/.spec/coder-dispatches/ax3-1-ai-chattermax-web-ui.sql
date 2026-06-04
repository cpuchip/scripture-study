-- AX3 task 1 — ai-chattermax web UI (Vue SPA + Go serves it).
-- Dispatched by the steward (Claude) on Michael's behalf, 2026-06-04.
-- code-pr pipeline: clone→plan→plan_review(qwen)→implement(kimi)→verify→review(qwen)→pr(DRAFT).
-- Base = main (the real backend); PR head = agent/code-pr/wi-<id>; auto-deploys to chat.ibeco.me on merge.
-- implement model = kimi-k2.6 (default; honest first-attempt on novel Vue territory — escalate on re-run if it flails).
INSERT INTO stewards.work_items (pipeline_family, current_stage, status, intent_id, origin, actor, input)
VALUES (
    'code-pr', 'clone', 'pending',
    '24c681f9-240f-400b-b3ee-5dc3f6bce992', 'human', 'human',
    jsonb_build_object(
        'repo',            'https://github.com/cpuchip/ai-chattermax',
        'base_branch',     'main',
        'max_tokens',      '64000',
        'revise_cap',      '2',
        'plan_revise_cap', '2',
        'binding_question', $bq$Add a web UI to the ai-chattermax Go server (repo root, module github.com/cpuchip/ai-chattermax) so the live backend at chat.ibeco.me has a usable face. Deliver BOTH parts in ONE PR.

PART A — Vue frontend in frontend/: a Vue 3 + Vite + Tailwind single-page app (this repo is standalone; the stack matches the workspace convention used by sibling projects). Screens / behavior:
- A join screen: the user enters a display name and a room name (default "lobby").
- A room view: opens a WebSocket to the EXISTING endpoint GET /ws/{room}?id=<name>, shows incoming messages in a scrolling transcript, and sends typed messages over the socket.
- A roster panel: polls the EXISTING endpoint GET /roster/{room} on an interval (~3s) and lists the participants returned.
Keep components small and idiomatic. No router library is required (a simple view switch is fine); use Tailwind for styling. Put the WS/roster logic in a composable so it can be unit-tested without a real network.

PART B — serve the SPA from Go: in cmd/server, embed the built frontend (frontend/dist) via go:embed and serve it at GET / with SPA history-fallback (any non-API, non-WS path returns index.html), WITHOUT changing the behavior of the existing /healthz, /ws/{room}, or /roster/{room} routes. Update the multi-stage Dockerfile to add a Node build stage (npm ci && npm run build in frontend/) producing frontend/dist BEFORE the Go build stage so the embedded assets are the real build.

IMPORTANT: go:embed requires the embedded directory to exist at `go build` time. Make `go build ./...` compile cleanly — either commit a placeholder frontend/dist/index.html, or ensure the ground-truth command builds the frontend first. The existing Go packages (room, scheduler, transcript, presence) are unchanged.

Ground-truth build+test command (build the frontend FIRST so the embed compiles): cd frontend && npm ci && npm run build && cd .. && go build ./... && go test ./...$bq$,
        'acceptance_criteria', $ac$1. FRONTEND EXISTS: frontend/ is a Vue 3 + Vite project (package.json with vue + vite, Tailwind configured) implementing a join screen (display name + room), a room transcript view, a message composer, and a roster panel.
2. WS WIRING: the room view opens a WebSocket to /ws/{room}?id=<name>, renders received messages, and sends composed messages over the socket. The WS/roster logic lives in a composable and is unit-tested WITHOUT a real network (fake/mocked socket).
3. ROSTER: the roster panel fetches GET /roster/{room} and renders the participants, refreshing on an interval.
4. GO SERVES SPA: cmd/server embeds frontend/dist via go:embed and serves it at GET / with history-fallback to index.html for unknown non-API/non-WS paths. A Go test asserts: GET / returns 200 with Content-Type text/html; an unknown path (e.g. /room/lobby) returns the SPA index; and /healthz still returns its JSON 200.
5. NO REGRESSION: the existing /ws/{room} and /roster/{room} handlers remain registered and behave as before; existing server tests still pass.
6. DOCKERFILE: a Node build stage (npm ci && npm run build) produces frontend/dist; the Go stage embeds that non-empty dist; the final image still runs as non-root with the /healthz HEALTHCHECK.
7. GREEN (ground truth): from the repo root, `cd frontend && npm ci && npm run build && cd .. && go build ./... && go test ./...` exits 0 with everything passing. This exact command is the gate.$ac$
    )
)
RETURNING id, input->>'sandbox' AS sandbox, current_stage, status;
