-- =====================================================================
-- Phase 3e.2 follow-up — broaden agent grants over the bridge surface
--
-- 3e.2.d shipped 12 grants for a v1 surface (study/lesson/talk gospel
-- triple, journal/review gospel_search, research web_search_exa,
-- research fetch-md). This migration opens the broader catalog now
-- that the bridge has all 8 servers responding (47 cached tools).
--
-- Discipline: READ-ONLY tools only. Substrate agents do not get
-- brain_create/update/delete, create_note/practice/task, log_practice,
-- review_card, etc. — those should remain Claude Code / operator-
-- controlled. Anything that mutates personal data (becoming brain
-- entries, journal practices) is left deny-by-default.
--
-- Bare tool names match what auto-promote (3e.2.d) produces.
-- ON CONFLICT DO NOTHING preserves the existing 12 grants.
-- =====================================================================

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  -- ============ study (already has gospel_search/get + webster_define) ============
  ('study', 'byu_citations',                'allow', 'manual'),
  ('study', 'byu_citations_books',          'allow', 'manual'),
  ('study', 'byu_citations_bulk',           'allow', 'manual'),
  ('study', 'yt_search',                    'allow', 'manual'),
  ('study', 'yt_get',                       'allow', 'manual'),
  ('study', 'webster_search',               'allow', 'manual'),
  ('study', 'webster_search_definitions',   'allow', 'manual'),
  ('study', 'modern_define',                'allow', 'manual'),
  ('study', 'brain_search',                 'allow', 'manual'),
  ('study', 'brain_recent',                 'allow', 'manual'),
  ('study', 'brain_get',                    'allow', 'manual'),

  -- ============ lesson ============
  ('lesson', 'byu_citations',                'allow', 'manual'),
  ('lesson', 'byu_citations_books',          'allow', 'manual'),
  ('lesson', 'yt_search',                    'allow', 'manual'),
  ('lesson', 'yt_get',                       'allow', 'manual'),
  ('lesson', 'webster_search',               'allow', 'manual'),
  ('lesson', 'webster_search_definitions',   'allow', 'manual'),
  ('lesson', 'modern_define',                'allow', 'manual'),

  -- ============ talk ============
  ('talk', 'byu_citations',                'allow', 'manual'),
  ('talk', 'byu_citations_books',          'allow', 'manual'),
  ('talk', 'yt_search',                    'allow', 'manual'),
  ('talk', 'yt_get',                       'allow', 'manual'),
  ('talk', 'webster_search',               'allow', 'manual'),
  ('talk', 'webster_search_definitions',   'allow', 'manual'),
  ('talk', 'modern_define',                'allow', 'manual'),

  -- ============ journal (already has gospel_search) ============
  ('journal', 'webster_define',  'allow', 'manual'),
  ('journal', 'webster_search',  'allow', 'manual'),
  ('journal', 'brain_search',    'allow', 'manual'),
  ('journal', 'brain_recent',    'allow', 'manual'),
  ('journal', 'brain_get',       'allow', 'manual'),

  -- ============ review (already has gospel_search) ============
  ('review', 'webster_define',  'allow', 'manual'),
  ('review', 'webster_search',  'allow', 'manual'),

  -- ============ research (already has web_search_exa + fetch-md*) ============
  ('research', 'web_search',     'allow', 'manual'),
  ('research', 'news_search',    'allow', 'manual'),
  ('research', 'instant_answer', 'allow', 'manual'),
  ('research', 'brain_search',   'allow', 'manual'),
  ('research', 'brain_recent',   'allow', 'manual'),
  ('research', 'brain_get',      'allow', 'manual'),

  -- ============ yt-gospel (no manual grants yet; full YT + scripture surface) ============
  ('yt-gospel', 'yt_search',     'allow', 'manual'),
  ('yt-gospel', 'yt_get',        'allow', 'manual'),
  ('yt-gospel', 'yt_list',       'allow', 'manual'),
  ('yt-gospel', 'yt_download',   'allow', 'manual'),
  ('yt-gospel', 'gospel_search', 'allow', 'manual'),
  ('yt-gospel', 'gospel_get',    'allow', 'manual'),
  ('yt-gospel', 'byu_citations', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
