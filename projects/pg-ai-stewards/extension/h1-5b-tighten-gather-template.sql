-- =====================================================================
-- Batch H.1.5b — tighten research-write gather stage input_template
--
-- Surfaced 2026-05-11 during H.1.4: kimi-k2.6 ran 9 chat rounds + 30+
-- tool calls + $0.42 without converging on a sources brief. The
-- original template said "6-12 credible sources" (soft guidance). The
-- model with active intent values (cross-reference, recency-matters)
-- internalized those as license to keep searching past the threshold.
--
-- Fix: rewrite the gather template with an explicit hard stop and
-- explicit end-of-turn instruction. Test whether kimi honors directive
-- constraints when they are unambiguous.
-- =====================================================================

UPDATE stewards.pipelines
   SET stages = jsonb_set(
       stages,
       '{0,input_template}',
       to_jsonb(
           'Binding question: {{input.binding_question}}' || E'\n\n' ||
           '## YOUR TASK' || E'\n\n' ||
           'Find **8 strong sources** that bear on the binding question. Then **STOP**, produce the sources brief, and end your turn. Do not keep searching for additional confirmation or dissenting voices once you have 8 strong sources — the synthesize stage and the review stage handle balance and verification downstream.' || E'\n\n' ||
           '## HARD CONSTRAINTS' || E'\n\n' ||
           '- **Maximum 8 sources.** Once you have 8 strong sources, STOP searching.' || E'\n' ||
           '- **Maximum 4 rounds of tool calls.** If you reach round 4 and don''t yet have 8 strong sources, produce the brief with what you have and end your turn.' || E'\n' ||
           '- **End-of-turn:** when you finish, your final message must be the sources brief in markdown. No further tool calls. No "let me also search for..."' || E'\n\n' ||
           '## TOOL GUIDANCE' || E'\n\n' ||
           'You have web_search_exa (Exa neural search), web_search (DuckDuckGo), news_search, fetch_url, fetch_urls, yt_search, yt_get, and others. Use 1-2 search calls per round to cast wide; use fetch_url to read a specific high-value source. Parallel tool calls in one round are fine — that''s ONE round.' || E'\n\n' ||
           '## FOR EACH SOURCE YOU KEEP' || E'\n\n' ||
           '- **Title** + **URL** + **publication date**' || E'\n' ||
           '- **One-sentence summary** of what it adds to the binding question' || E'\n' ||
           '- **Short verbatim quote** (1-3 sentences) you might draw on in synthesis' || E'\n' ||
           '- **Source type:** primary documentation / news reporting / opinion / vendor blog / academic / etc.' || E'\n' ||
           '- **Credibility note:** primary source for this claim? secondary? recency vs domain half-life?' || E'\n\n' ||
           '## OUTPUT FORMAT' || E'\n\n' ||
           'Produce a markdown sources brief: a numbered list of 8 sources, each with the five fields above. **No prose intro. No prose outro.** Just the structured list. The synthesize stage drafts the actual research piece from your brief — your job is the brief, not the prose.'
       )
   )
 WHERE family = 'research-write';

-- Verify
SELECT family, jsonb_array_length(stages) AS n_stages,
       substring(stages->0->>'input_template' FROM 1 FOR 200) AS gather_template_head
  FROM stewards.pipelines
 WHERE family = 'research-write';
