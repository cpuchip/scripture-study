---
title: Phase F — Zion / Council
date: 2026-05-11
status: design sub-spec — ready for Phase F implementation (after A–E lived with)
parent: full-agentic-substrate.md (D-F1..F4 ratification, plus 2026-05-11 re-validation)
purpose: >
  Add the multi-agent council primitive to the substrate. Today every
  work_item is one agent walking a maturity ladder (commission). Phase F
  adds councils — multiple agents reasoning together about a single intent,
  with a bishop facilitating voices toward agreement. Modeled on the ward
  council pattern (D&C 102, Mosiah 26, Acts 15), not on generic
  multi-agent orchestration.
---

# Phase F — Zion / Council

## I. Binding problem

The substrate today has one collaboration shape: **commission**. One agent walks a maturity ladder for one work_item; gates ratify between maturities; lessons are extracted at the end. This is the right shape for most work — but it's not the only shape the gospel framework knows.

The other shape is **council**: multiple agents reasoning together about a single intent, with a bishop (D&C 102) facilitating not orchestrating. Modeled in the ward council. Modeled in Mosiah 26:13-14 when Alma sought guidance. Modeled in Acts 15 when the early church convened on the Gentile question.

The substrate needs councils for the questions that no single agent should decide alone:
- "Should we adopt this new pipeline?"
- "Is this study's reading of D&C 130 sound, or are we missing something?"
- "Should we revise the covenant in light of these last six atonements?"

Phase F is hard because it depends on every prior phase being lived with: A (in-flight discipline), B (gate ratification), C (intent + covenant as substrate state), D (lesson accumulation), E (trust earned). A council convened without earned trust is theater. A council on an unstated intent is a meeting. A council without ratified lessons is starting from zero. Phase F is the substrate's culmination, not its centerpiece.

## II. Success criteria

1. **A human convenes a council** via Stewards-UI: picks an intent, selects 2–4 agent_families with roles (proposer / critic / synthesizer), assigns a bishop. Council appears in `/councils` UI within 5s.
2. **All members dispatch** in parallel via the standard work_queue path, each with role-specific framing prepended to their system prompt.
3. **Synthesizer dispatches** after all members respond, receives all responses + intent + covenant in context, produces a proposed resolution.
4. **Bishop** (human, or master-tier agent on the relevant pipeline for low-stakes councils) reviews the synthesizer's proposal and either resolves, requests another round, or dissolves.
5. **Resolution writes to all three destinations** per D-F3: canonical row in `stewards.resolutions`, with optional promotion to a `study/` file (doctrinal questions) or `.mind/decisions.md` (engineering questions). Type-of-question determines destination.
6. **Manual + system-suggested convening** per D-F4: humans convene from UI; watchman pass output and dashboard banner suggest convenings when patterns emerge in atonement lessons (5+ lessons on the same pipeline+stage).
7. **One council at a time** per D-F1 — the substrate refuses to convene a second council while one is `deliberating`. Lift after a month if real demand emerges.

## III. Constraints and boundaries

**In scope:**
- `stewards.councils` + `stewards.council_members` + `stewards.resolutions` tables
- `stewards.convene_council(intent_id, members jsonb, bishop text)` SQL function — creates council, dispatches members in parallel
- `stewards.synthesize_council(council_id)` SQL function — fires after all members respond
- `stewards.resolve_council(council_id, resolution text, destination text)` SQL function — bishop's resolution path
- bgworker auto-fire extension: detect when all members of a council have responded, auto-fire synthesize
- Role-specific prompt prepending in `compose_system_prompt` (when session belongs to a council member)
- Watchman pass extension: detect lesson clusters and emit convening suggestions
- Stewards-UI Council view: live deliberation visible (each member's contribution streams in), synthesizer's proposal at the bottom, bishop's resolution form

**Out of scope (explicitly):**
- More than one concurrent council (D-F1 ratified)
- Auto-convening (D-F4 ratified system-suggested-with-human-convene; not auto-fire)
- Multi-round deliberation chains beyond bishop's "request another round" (one round of revisions max for F1; tune later)
- Cross-substrate councils (the substrate doesn't talk to other substrates — no federation in F1)
- council_authority as a separate trust dimension (F2 nuance noted: ships with master-on-pipeline rule; council_authority is a future evolution path with debug agent as candidate first cultivator)

## IV. Prior art

- **Phase 5b's parallel chat dispatch** — multiple chats can be in-flight simultaneously via the work_queue. Council members reuse this; nothing new at the dispatch layer.
- **Phase 5b's bgworker auto-fire** — `_gate_eval` / `_scenarios_gen` / `_verify` / `_sabbath` / `_atonement` markers. Phase F adds `_council_member` (per-member dispatch) and `_council_synthesize` (the synthesizer dispatch).
- **Phase C's compose_system_prompt extension** — already injects intent + covenant. Phase F adds role-specific framing for council members.
- **Phase D's lessons table** — watchman scans this for convening signals.
- **Phase E's trust_scores** — bishop eligibility check for D-F2 (master-on-pipeline-of-intent).
- **Ward council pattern** — D&C 102 (high council), Mosiah 26:13-14 (Alma seeking guidance), Acts 15 (Jerusalem council). Member roles draw from these. Substrate is specifically modeled on ward council, not generic multi-agent.
- **Brain v3** — has none of this. Phase F is the substrate's most distinctive ground.

## V. Proposed approach

### V.1 Schema

```sql
CREATE TABLE stewards.councils (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id       uuid NOT NULL REFERENCES stewards.intents(id),
    binding_question text NOT NULL,           -- the single question the council convenes on
    convened_at     timestamptz NOT NULL DEFAULT now(),
    convened_by     text NOT NULL,             -- human name or 'watchman' for system-suggested
    bishop          text NOT NULL,             -- 'human:michael' | 'agent:debug:study-write:master'
    status          text NOT NULL DEFAULT 'deliberating'
                      CHECK (status IN ('deliberating', 'synthesizing', 'awaiting_bishop',
                                         'resolved', 'dissolved')),
    resolution_id   uuid REFERENCES stewards.resolutions(id),
    dissolved_reason text,
    resolved_at     timestamptz
);

-- One row per (council, member). Each member gets dispatched as its own
-- chat in the work_queue.
CREATE TABLE stewards.council_members (
    council_id      uuid NOT NULL REFERENCES stewards.councils(id) ON DELETE CASCADE,
    agent_family    text NOT NULL,
    role            text NOT NULL CHECK (role IN ('proposer', 'critic', 'synthesizer')),
    work_id         bigint,                    -- the work_queue id of this member's dispatch
    response        text,                      -- assistant content when complete
    completed_at    timestamptz,
    PRIMARY KEY (council_id, agent_family, role)
);

CREATE INDEX council_members_council ON stewards.council_members (council_id);

-- Resolutions canonical (D-F3 — three destinations; this is the canonical one)
CREATE TABLE stewards.resolutions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    council_id      uuid REFERENCES stewards.councils(id),
    resolved_at     timestamptz NOT NULL DEFAULT now(),
    resolved_by     text NOT NULL,             -- human name or agent identifier
    text            text NOT NULL,             -- the resolution itself
    promoted_to     text,                      -- 'study/<slug>.md' | '.mind/decisions.md' | NULL
    promoted_at     timestamptz
);

CREATE INDEX resolutions_council ON stewards.resolutions (council_id);

-- One-council-at-a-time enforcement (partial unique index)
CREATE UNIQUE INDEX one_active_council
    ON stewards.councils ((1))
    WHERE status IN ('deliberating', 'synthesizing', 'awaiting_bishop');
```

### V.2 Convene

```sql
CREATE OR REPLACE FUNCTION stewards.convene_council(
    p_intent_id     uuid,
    p_binding_question text,
    p_members       jsonb,              -- [{"agent_family": "...", "role": "..."}, ...]
    p_bishop        text,
    p_convened_by   text DEFAULT 'human'
) RETURNS uuid
LANGUAGE plpgsql AS $$
DECLARE
    v_council_id uuid;
    v_member jsonb;
    v_work_id bigint;
    v_session_id text;
BEGIN
    -- D-F1 enforcement
    IF EXISTS (SELECT 1 FROM stewards.councils
                WHERE status IN ('deliberating', 'synthesizing', 'awaiting_bishop')) THEN
        RAISE EXCEPTION 'one council at a time — resolve or dissolve the active council first';
    END IF;

    INSERT INTO stewards.councils (intent_id, binding_question, convened_by, bishop)
    VALUES (p_intent_id, p_binding_question, p_convened_by, p_bishop)
    RETURNING id INTO v_council_id;

    -- Dispatch each member in parallel
    FOR v_member IN SELECT * FROM jsonb_array_elements(p_members) LOOP
        v_session_id := 'council--' || substring(v_council_id::text FROM 1 FOR 8) ||
                        '--' || (v_member->>'role') ||
                        '--' || (v_member->>'agent_family');

        INSERT INTO stewards.sessions (id, label, kind)
        VALUES (v_session_id,
                format('council %s role=%s agent=%s', v_council_id,
                       v_member->>'role', v_member->>'agent_family'),
                'council')
        ON CONFLICT (id) DO NOTHING;

        -- compose role-specific user message; render template council_<role>
        -- (proposer / critic / synthesizer — the latter for the synthesize step)
        ... compose body via compose_system_prompt + dry_run_chat ...

        INSERT INTO stewards.work_queue (kind, provider, payload)
        VALUES ('chat', 'opencode_go', jsonb_build_object(
            ...,
            '_council_id', v_council_id::text,
            '_council_member', true,
            '_council_role', v_member->>'role',
            'tools_disabled', false  -- members get tools (unlike gates) — they need them to reason
        ))
        RETURNING id INTO v_work_id;

        INSERT INTO stewards.council_members (council_id, agent_family, role, work_id)
        VALUES (v_council_id, v_member->>'agent_family', v_member->>'role', v_work_id);
    END LOOP;

    -- Extend sessions_kind_check to include 'council'
    -- (separate migration, mirrors 5c's pattern for adding 'gate')

    RETURN v_council_id;
END;
$$;
```

### V.3 Synthesize

The bgworker auto-fire detects when all members of a council have responded (all `council_members.response IS NOT NULL`). Fires `synthesize_council`:

```sql
CREATE OR REPLACE FUNCTION stewards.synthesize_council(p_council_id uuid)
RETURNS bigint
LANGUAGE plpgsql AS $$
DECLARE
    v_synth_session_id text;
    v_payload jsonb;
    v_work_id bigint;
BEGIN
    UPDATE stewards.councils SET status = 'synthesizing' WHERE id = p_council_id;

    -- Compose synthesizer prompt: intent + covenant + binding question +
    -- all member responses (formatted as labeled blocks).
    -- Template: gate_prompts.council_synthesize.

    -- Dispatch as another work_queue chat with marker _council_synthesize=true.
    -- Synthesizer doesn't get tools (output is a structured proposed resolution).

    RETURN v_work_id;
END;
$$;
```

The synthesizer's response (apply step) sets `councils.status = 'awaiting_bishop'` and stores the proposed resolution text on a draft row in `resolutions` (NULL `resolved_by`, NULL `promoted_to`).

### V.4 Resolve

```sql
CREATE OR REPLACE FUNCTION stewards.resolve_council(
    p_council_id uuid,
    p_action text,                        -- 'accept' | 'request_revision' | 'dissolve'
    p_resolution_text text,               -- bishop's final text (may differ from synth proposal)
    p_destination text,                   -- 'study/<slug>.md' | '.mind/decisions.md' | NULL
    p_resolved_by text                    -- human name or master-agent identifier
) RETURNS uuid
LANGUAGE plpgsql AS $$
... handles three cases ...
$$;
```

`accept` writes to `resolutions` (canonical), sets `councils.status='resolved'`, records `promoted_to`. The actual file write to `study/` or `.mind/decisions.md` follows the same pattern as Phase D's lesson promotion (substrate emits a "pending file write" record, sidecar/next commit materializes — keeps substrate stateless on FS).

`request_revision` re-dispatches all member chats with the synthesizer's proposal in context as "Bishop requests revision because: <reason>." Members respond again; synthesizer fires again. One round of revision allowed for F1.

`dissolve` sets `councils.status='dissolved'`, records `dissolved_reason`. Resolved/dissolved state allows another council to convene.

### V.5 Bishop eligibility (F2)

```sql
CREATE OR REPLACE FUNCTION stewards.bishop_eligible(
    p_bishop text,                        -- 'human:michael' or 'agent:<family>:<pipeline>:master'
    p_intent_id uuid
) RETURNS boolean
LANGUAGE plpgsql STABLE AS $$
DECLARE
    v_intent stewards.intents%ROWTYPE;
    v_agent_family text;
    v_pipeline text;
BEGIN
    -- Humans always eligible
    IF p_bishop LIKE 'human:%' THEN
        RETURN true;
    END IF;

    SELECT * INTO v_intent FROM stewards.intents WHERE id = p_intent_id;

    -- Low-stakes check (per 2026-05-11 ratification: technical/factual intents only)
    -- Heuristic: intent.scripture_anchor IS NULL AND intent.values_hierarchy
    -- doesn't contain 'doctrinal' or 'spiritual' as values keys.
    IF v_intent.scripture_anchor IS NOT NULL
       OR v_intent.values_hierarchy::text ~* '(doctrinal|spiritual|discernment)' THEN
        RETURN false;  -- High-stakes; human bishop only
    END IF;

    -- Parse 'agent:<family>:<pipeline>:master'
    -- Verify agent is master-tier on that pipeline (any model)
    -- Recommend Phase F1 just trusts the encoded string; Stewards-UI bishop
    -- picker pre-filters to eligible agents only.

    RETURN ...;
END;
$$;
```

**F2 future evolution path** (noted by Michael 2026-05-11): introduce `council_authority` as a separate trust dimension. Debug agent is candidate first cultivator because its skills are designed to get at the root — well-suited to the bishop's facilitation role. When `council_authority` exists, `bishop_eligible` becomes "agent has earned council_authority master tier." For F1, master-on-pipeline is the rule.

### V.6 System-suggested convening (F4)

Watchman pass (Phase 2.7) extension: at each pass, check `stewards.lessons` for clusters:
- 5+ ratified lessons on the same `(pipeline_family, current_stage)` not yet referenced in any council
- Or 3+ lessons on the same theme (TBD how to detect — initially keyword-match on lesson.content)

When a cluster is found, emit a watchman finding of kind `council_suggested` with the proposed intent + binding question. Stewards-UI dashboard pulls watchman findings and shows a banner: "Watchman suggests convening a council on: <binding question>. <Convene> <Dismiss>"

Convene from the banner pre-fills the intent and binding question; human still picks members + bishop.

### V.7 Stewards-UI Council view

New top-level route `/councils` (and `/councils/:id`).

`/councils` — list view: active council (if any) prominently, history below.

`/councils/:id` — live deliberation:
- Header: intent purpose + binding question + bishop
- Members section: each member's role + agent + status (dispatched / responding / done) + their full response when complete (markdown rendered)
- Synthesizer section: proposed resolution when synthesizer completes
- Bishop section: resolution form (textarea pre-filled with synthesizer proposal, edit freely + destination dropdown + Accept / Request revision / Dissolve buttons)

All sections auto-refresh (5s polling like Dashboard). The UI feels like a *room* — that's the explicit design goal per the proposal.

API:
- `POST /api/councils/convene` — wraps `convene_council` SQL fn
- `GET /api/councils/list` — recent + active
- `GET /api/councils/get?id=` — full council detail with members + synthesizer + draft resolution
- `POST /api/councils/resolve` — wraps `resolve_council` SQL fn
- `POST /api/councils/request-revision` — same fn, action='request_revision'
- `POST /api/councils/dissolve` — same fn, action='dissolve'

## VI. Open questions / follow-ups

- **Member count bounds.** D-F1 says "1 concurrent council." Members per council? Recommend 2–4 (proposer + critic + optional second proposer + synthesizer; or 3 proposers + 1 critic + 1 synthesizer for a 5-member council). Limit at 5 for F1; lift if needed.
- **Bishop dispatch path for agent bishops.** When `bishop` is an agent identifier, who dispatches the bishop's resolution decision? A separate `_council_bishop_dispatch=true` chat? Or does the synthesizer's output land directly as resolution and the agent-bishop is just a metadata tag? Recommend the former — explicit dispatch keeps the gate-style audit trail.
- **Concurrency lift criterion.** D-F1 ratified "1 at a time, lift if real demand emerges." Define "real demand": measure how often the substrate refuses a second convene with the unique-index error. If >5 refusals/week, lift to 3 concurrent.
- **System-suggested specificity.** Watchman finds clusters; does it also suggest the binding question? Recommend yes — the proposed binding question is part of the suggestion. Human edits if wrong.
- **Resolution promotion-to-study.** Same file-write mechanism question as Phase D's lesson promotion. Sidecar / pending-write pattern recommended; substrate stays FS-stateless.
- **F-Phase ordering with E.** F's bishop_eligible reads trust_scores. F's synthesizer + members earn (or lose) trust like commission agents. Order locked: E ships before F.
- **Council kind audit.** `sessions_kind_check` extended to add 'council' (mirror 5c pattern). Migration file `Xa-sessions-council-kind.sql` ships with F.
- **What if all members error?** If 2 of 3 members error and 1 responds, synthesizer fires anyway with 1 response? Or wait + retry? Recommend: synthesizer fires with whatever members succeeded; bishop sees which members errored and decides.

## VII. Estimated programming time

- V.1 schema + V.2 convene + V.3 synthesize + V.4 resolve: 1–2 sessions
- V.5 bishop_eligible + V.6 watchman convening suggestions + bgworker auto-fire extension: 1 session
- V.7 Stewards-UI Council view (the heaviest UI in the substrate; feels-like-a-room is the design bar): 1–2 sessions
- Prompt engineering (proposer / critic / synthesizer / bishop templates): 1 session

**Total: 4–5 sessions.** Matches the proposal's 3-4 estimate plus a session for the room-feel UI.

## VIII. Acceptance scenarios

- A human convenes a council via Stewards-UI: picks an existing intent ("evaluate three MCP server implementations"), enters binding question, selects 3 agent_families (debug as proposer, plan as critic, debug as synthesizer), assigns themself as bishop. Council appears at `/councils/:id` within 5s.
- All 3 members dispatched in parallel via work_queue. Each responds within 60s. Members section shows responses live.
- bgworker detects all members complete; synthesizer fires automatically. Synthesizer's proposed resolution appears in `/councils/:id` synthesizer section within 90s.
- Bishop edits the resolution text, picks destination `.mind/decisions.md`, clicks Accept. `resolutions` row written; council status `resolved`; pending-file-write record emitted; next git commit materializes the file write.
- A second `convene_council` call while the first is `deliberating` raises `one council at a time`.
- After 5 ratified atonement lessons accumulate on study-write outline stage, watchman emits a finding `council_suggested`. Dashboard banner appears: "Watchman suggests convening a council on: 'Should the study-write outline stage's prompt be revised?'"
- A council convened on a doctrinal intent (intent.scripture_anchor IS NOT NULL) refuses an agent-bishop selection. UI bishop picker shows only humans for that intent.
- Phase E trust_scores update after the council resolves — proposer / critic / synthesizer get successful_completions += 1 if the bishop accepted unmodified, with full weight; if the bishop modified the resolution before accepting, that counts as a partial override (TBD weighting; revisit after first month of councils).
