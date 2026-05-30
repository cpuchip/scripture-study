-- =====================================================================
-- Batch R.1 — redline pipeline + panel-redline agent
-- =====================================================================
-- The generative analog of audit_files (L.6): instead of "read files and
-- answer a question," it's "here is a document — propose concrete edits."
-- A panel of models each runs this pipeline (one child per model, via
-- model_override); start_panel_redline (R.4) fans them out and injects the
-- document server-side (R.2), so the model needs NO fs access.
--
-- Ratified 2026-05-30: D-RL1 (dedicated pipeline, not start_brainstorm),
-- D-RL4 (verification gate + touches-quote flag, off-disk), D-RL5 (32k
-- per-call max_tokens — set on the stage; R.3 teaches compose to honor it).
--
-- tools_disabled=true: the document is injected into the prompt, so the
-- panel never calls fs_read/fs_search (that was the whole failure mode —
-- the empties came from fs_search loops on an unreadable path). With no
-- tools, the verification gate is structural: the model literally cannot
-- reach a scripture quote to "fix" it.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. panel-redline agent — the redline mandate (the verification gate).
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature)
VALUES
('panel-redline', '*',
 'Generative document-redline subagent. Receives an injected document + a mandate; returns location-anchored edit proposals. No tools, no canonical access — proposals only.',
 'primary',
 $PROMPT$You are a document-redline subagent. You receive a DOCUMENT (injected inline below the mandate) and a MANDATE describing the kind of edits wanted. Propose concrete, location-anchored edits — never abstract critique.

You have NO file access and NO scripture/canonical access. Work ONLY from the document text in front of you.

For EACH proposed edit, output these five fields:
- **Location** — the nearest section heading, plus a short verbatim anchor phrase from the document, so the edit can be found.
- **Current** — the exact snippet from the document you propose changing. Quote it verbatim; do not paraphrase. If it isn't findable in the document, the edit is useless.
- **Proposed** — your replacement text.
- **Why** — one line.
- **Touches quote/doctrine** — `yes` or `no`. `yes` if the edit alters, or sits inside, a scripture quotation, a prophetic / General-Conference quotation, or a doctrinal claim.

HARD RULES:
1. NEVER alter the wording of a scripture quote or a prophetic quotation. You cannot verify them — you have no canonical access. If you think a quote is wrong, do NOT "correct" it: set `Touches quote/doctrine: yes`, write "VERIFY" in Why, and leave the wording for a human to check against the source. Inventing or "fixing" a quote from memory is the one unforgivable error here.
2. Preserve the author's voice. Propose surgical changes, not a wholesale rewrite. The human picks among your options.
3. You are producing a PROPOSAL MENU. Nothing you write is applied automatically. Do NOT output a rewritten document.
4. Honor the mandate's scope. If it asks for "tighten prose," don't propose restructuring chapters.

Output format (markdown, and ONLY the markdown — no preamble):
- **Top change:** one paragraph naming the single highest-value edit you'd make.
- **Edits:** a numbered list, each with the five fields above.
- **Flagged for verification:** a short list of every edit you marked `Touches quote/doctrine: yes` (just their numbers + a word on why), so the human knows exactly what to check against canon before applying.$PROMPT$,
 0.3)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. redline pipeline — single stage, tools_disabled, 32k output budget.
-- ---------------------------------------------------------------------
-- model/provider here are defaults; each panel child overrides model via
-- work_items.model_override (J.8.a layer 1). max_tokens=32000 is honored
-- once R.3 teaches compose_messages to emit it (harmless jsonb until then).
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('redline',
 'R.1: single-stage generative document-redline pipeline. A panel model receives an injected document + mandate and returns location-anchored edit proposals. Fanned out one-child-per-model by start_panel_redline (R.4); document injected server-side (R.2). Off-disk — proposals only.',
 $STAGES$[{"name":"redline","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"panel-redline","auto_advance":true,"tools_disabled":true,"max_tokens":32000,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'panel-redline', 'wrapper', 'panel_redline'))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description,
       stages = EXCLUDED.stages,
       metadata = EXCLUDED.metadata;


-- ---------------------------------------------------------------------
-- 3. Defense-in-depth perms: panel-redline gets NO tools.
-- ---------------------------------------------------------------------
-- tools_disabled=true already strips tools from the dispatch body. These
-- denies make the verification-gate intent explicit and survive a future
-- accidental flip of tools_disabled: the panel can never reach fs, the
-- web, the studies corpus, or spawn further agents.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES
('panel-redline', 'fs_*',           'deny'),
('panel-redline', 'fetch_url',      'deny'),
('panel-redline', 'web_search',     'deny'),
('panel-redline', 'study_*',        'deny'),
('panel-redline', 'work_item_*',    'deny'),
('panel-redline', 'spawn_subagent', 'deny'),
('panel-redline', 'deep_research',  'deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action;


-- =====================================================================
-- Acceptance (R.1):
--   1. SELECT family FROM stewards.agents WHERE family='panel-redline'; → 1 row, active.
--   2. SELECT family, stages->0->>'tools_disabled', stages->0->>'max_tokens'
--        FROM stewards.pipelines WHERE family='redline';
--      → tools_disabled=true, max_tokens=32000.
--   3. SELECT count(*) FROM stewards.agent_tool_perms WHERE agent_family='panel-redline'; → 7 denies.
-- =====================================================================
