---
date: 2026-06-23
topic: sub_work_item (Delegate) + rich-docs council/ratify + multimodal P0 (local vision serves)
lane: pg-ai-stewards
---

# Chat that delegates, and a rig that can see

A long continuation after Stewdio P3/P4 shipped. Three things landed, each
proven: the chat can now spawn linked work, the rich-document vision is
ratified into a phased plan, and the local rig serves vision end to end.

## sub_work_item — the chat can Delegate (OSS `47046ad`, chain → 46)

The Ask-vs-Delegate split. The chat's Ask mode (45) answers from retrieval; this
adds Delegate: ask it to GO DO substantial work and it spawns a real work_item
running a pipeline, linked back to the parent, watched in the cockpit.

`extension/46-chat-tasks.sql` = a `start_task` tool (`chat_start_task_tool`) →
`work_item_create` + dispatch. The linkage is RELIABLE, not model-guessed:
Michael's ask was "it needs to link back to the parent," so the parent comes
from the **session id** — a work-item-grounded chat is `stewdio-<uuid>`, so the
tool pulls the parent uuid server-side and sets `parent_work_item_id`. A doc /
empty chat spawns top-level. Verified `check_and_dispatch_fanout_aggregator`
no-ops for a non-fanout parent, so the link is safe. The chat still can't call
`work_item_create` directly — `start_task` is the one controlled write path; the
agent prompt was softened from "read-only" to the Ask/Delegate split.
**Proven twice on live:** the tool created+linked+dispatched a child; and the
*agent itself* chose `start_task` when asked, spawned a child nested under the
parent, and replied "started it — watch it in the cockpit." This also closed the
deferred NL-launcher (#224).

## Rich documents in chat — council → ratified → spec

Michael's vision, widened by him in council: an attachment is **injectable
subject material** you can drop into ANY chat (empty or work-item-grounded); the
chat holds work_item + media + a corpus/project *lens* together, reasons over the
combination, renders the media inline, and `start_task` spawns work from the
whole. Ratified 100% with one addition (UI renders rich media inline).

De-risk findings that shaped it: the :8090 router already forwards image parts
untouched (no router work); the substrate assumes string `content` so a parallel
`content_parts jsonb` column + a `compose_messages` passthrough is the
load-bearing change (the OpenAI dispatch path forwards arrays verbatim — the
bright spot); **Gemini paid / Vertex = private-safe** (no train-on-data) so the
privacy fork dissolves into a *preference*; the rig models are vision-capable and
the mmproj projectors are on disk. Phasing P0→P4, 6 decisions, spec at
`projects/pg-ai-stewards-oss/.spec/proposals/rich-docs-in-chat.md` (`911d9d8`).

## P0 — local vision serves (llama-chip `ab4c9de`) — PROVEN

llama.cpp does multimodal via the model GGUF + an `--mmproj` projector; llama-chip
never passed it. Now: a `mmproj`/`no_mmproj` config field, `models.Discover`
pairs each model with its co-located `mmproj-*.gguf` (+ a `FindMMProj` helper +
a `supports_vision` flag), and `rig.args()` appends `--mmproj` (explicit wins,
else auto-detect). Unit-tested (projector found + paired + never leaks into the
model list); build/vet/test green.

**Live proof (Michael: "you own the rig for now"):** stopped the rig, rebuilt,
relaunched on the new binary, loaded gemma-4-26b-a4b via a `vision` profile
(auto-mmproj), and sent it a First Orbit game screenshot. It described the scene
**with every telemetry value correct** — "space flight simulation... TERRA, First
Flight... ALT 0m, Periapsis -600.0 km... Stage 1 of 2, 4.5k fuel... TWR 0.00...
LANDED on Terra." It *read the pixels*. Local, $0, private. Then brought
`dance-moe` (both substrate models) up on the new binary — fits WITH the vision
towers (24106/24104 MiB, very tight) and both health-probe ok, so the substrate
is whole again + now vision-capable.

## Rig state (I own it for now; Michael takes it under his terminal later)

The rig runs as my background process on the new binary, serving `dance-moe`
(qwen3.6-35b-a3b + gemma-4-26b-a4b, both vision-capable). VRAM is at ~99.7% — a
real vision turn (big image + context) could OOM, so the small follow-up is a
`dance-vision` profile with trimmed contexts to leave room for the encoder +
image tokens. The `vision` profile (gemma alone, headroom) is the safe fallback.

## Carry-forward

- **P1 next — the load-bearing substrate slice:** `content_parts jsonb` on
  `messages` + `compose_messages` passthrough for array rows + a `vision` role
  alias + inject an image part into a chat turn → vision model → grounded answer,
  end to end. Then P2 (attachments + UI inline rendering), P3 (docs + corpus
  lens), P4 (spawn from the combination).
- VRAM-trim a `dance-vision` profile (small).
- Autonomy stays PAUSED until more core pg-ai-stewards work lands (Michael's call
  — so we don't pay the spin-down-between-jobs tax).
- The digester-reads-our-repos council item stays deferred until it can be
  sandboxed away from the keys.

## Presiding / accounting

The chat-can-write capability (start_task) and the rich-doc plan both crossed
into new standing capability → `dominion_in_council` satisfied by Michael's
explicit ratification ("lets get sub work_item working"; "100% agree"). The rig
restart was on his explicit grant ("you own it for now"). All commits tested +
pushed; the test work_items were cancelled; the rig left in a working state.
