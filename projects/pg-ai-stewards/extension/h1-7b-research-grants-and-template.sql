-- =====================================================================
-- Batch H.1.7b — research agent tool grants + gather template update
--
-- Two changes:
--   1. Grant 7 new tools to the 'research' agent family:
--      - fs_read, fs_list, fs_search  (substrate-scoped filesystem)
--      - work_item_list, work_item_show  (prior work)
--      - watchman_pass_show, watchman_passes_list  (substrate state)
--
--      study_* tools are already granted (existing wildcard). Escalation
--      write tools (work_item_escalation_*) are NOT granted — they belong
--      to the operator review surface.
--
--   2. Update research-write's gather stage input_template to:
--      - Add a "CONSULT PRIOR WORK FIRST" section before external search
--      - Bump the round budget from 4 to 8 (3-4 prior-work rounds,
--        3-4 external-search rounds; total ≤ 8)
--
-- Idempotent: agent_tool_perms uses ON CONFLICT DO NOTHING (a deny
-- entry would block; we trust the existing deny '*' allowlist
-- pattern). The pipelines update uses jsonb_set with an explicit
-- WHERE on the stage name.
-- =====================================================================

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('research', 'fs_read',              'allow', 'manual'),
  ('research', 'fs_list',              'allow', 'manual'),
  ('research', 'fs_search',            'allow', 'manual'),
  ('research', 'work_item_list',       'allow', 'manual'),
  ('research', 'work_item_show',       'allow', 'manual'),
  ('research', 'watchman_pass_show',   'allow', 'manual'),
  ('research', 'watchman_passes_list', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;

-- Update the gather stage's input_template. We rebuild the stages
-- array with the new template, preserving every other stage as-is.
-- The new template:
--   - Adds "## CONSULT PRIOR WORK FIRST" before "## TOOL GUIDANCE"
--   - Lists the new tools (study_search, fs_search, work_item_list,
--     fs_read) and what each is for
--   - Bumps "Maximum 4 rounds" → "Maximum 8 rounds"
--   - Bumps "Maximum 8 sources" → "Maximum 8 strong sources OR
--     equivalent value via prior-work citation"
--
-- The new template is verbatim below. Stewardship note: this template
-- is one place we may want to fork per-pipeline once H.2 lands (the
-- context-gather stage owns the prior-work-reading job and gather
-- stays focused on external search). For H.1.7 we put it all in
-- gather; H.2 splits it cleanly.

DO $$
DECLARE
    v_new_template text;
    v_new_stages   jsonb;
BEGIN
    v_new_template :=
$T$Binding question: {{input.binding_question}}

## YOUR TASK

Find sources that bear on the binding question, give precedence to what we already know, and produce a sources brief. Then **STOP** and end your turn. Do not keep searching for additional confirmation once you have enough strong sources — the synthesize stage and the review stage handle balance and verification downstream.

## CONSULT PRIOR WORK FIRST

Before external search, check what the substrate already knows. Spend 2-4 rounds here when the topic is one we've worked on; skip to external search when it's genuinely new ground.

- `study_search` — full-text search of the substrate's studies corpus (gospel, research, planning). Returns matching slugs + snippets + ranks.
- `fs_search` — regex search across journals (`.spec/journal/*`), proposals (`.spec/proposals/*`), mind files (`.mind/*`), and docs (`docs/**`). Use this to find prior work mentioning a topic by name.
- `fs_read` — read a journal or proposal in full once `fs_search` surfaces it.
- `work_item_list` / `work_item_show` — list and inspect prior work_items on this binding question or adjacent.
- `study_get` / `study_similar` — read a study by slug, find related studies via embedding edges.

When prior work already answers a part of the binding question, your brief should cite it — point at the existing source instead of over-searching externally for confirmation.

## HARD CONSTRAINTS

- **Maximum 8 strong sources in the final brief.** Sources can be a mix of prior-work citations (substrate studies, journals, proposals) and external sources.
- **Maximum 8 rounds of tool calls total.** Typical shape: 2-4 rounds of prior-work consultation, then 3-4 rounds of external search. If you reach round 8, produce the brief with what you have and end your turn.
- **End-of-turn:** your final message must be the sources brief in markdown. No further tool calls. No "let me also search for..."

## TOOL GUIDANCE — EXTERNAL SEARCH

After consulting prior work, you have `web_search_exa` (Exa neural search), `web_search` (DuckDuckGo), `news_search`, `fetch_url`, `fetch_urls`, `yt_search`, `yt_get`, and others. Use 1-2 search calls per round to cast wide; use `fetch_url` to read a specific high-value source. Parallel tool calls in one round are fine — that's ONE round.

## FOR EACH SOURCE YOU KEEP

- **Title** + **URL or substrate path** + **publication date**
- **One-sentence summary** of what it adds to the binding question
- **Short verbatim quote** (1-3 sentences) you might draw on in synthesis
- **Source type:** prior-work-substrate / primary documentation / news reporting / opinion / vendor blog / academic / etc.
- **Credibility note:** primary source for this claim? secondary? recency vs domain half-life?

## OUTPUT FORMAT

Produce a markdown sources brief: a numbered list of up to 8 sources, each with the five fields above. **No prose intro. No prose outro.** Just the structured list. The synthesize stage drafts the actual research piece from your brief — your job is the brief, not the prose.$T$;

    -- Rebuild stages: take the existing array, find gather, replace its input_template.
    SELECT jsonb_agg(
        CASE
            WHEN s->>'name' = 'gather'
                THEN jsonb_set(s, '{input_template}', to_jsonb(v_new_template))
            ELSE s
        END
        ORDER BY ord
    )
    INTO v_new_stages
    FROM stewards.pipelines p,
         jsonb_array_elements(p.stages) WITH ORDINALITY AS arr(s, ord)
    WHERE p.family = 'research-write';

    UPDATE stewards.pipelines
       SET stages     = v_new_stages,
           updated_at = now()
     WHERE family = 'research-write';
END
$$;

-- Sanity: dump the new gather template length so the operator can spot-check.
SELECT 'gather template length:' AS check_name,
       length(s->>'input_template') AS chars
  FROM stewards.pipelines, jsonb_array_elements(stages) s
 WHERE family='research-write' AND s->>'name'='gather';
