# Verify-suite report — 2026-06-09-1341

Scratch: `pg-ai-stewards-dev:pg18` as `pg-stewards-verify` (hermetic: no bgworker, no provider secrets). Live: `pg-ai-stewards-dev`.
Replay order: `extension/migration-order.txt` (git first-add; intra-commit lexical).

## 1. Bootstrap replay

| | count |
|---|---|
| replayed ok | 197 |
| failed | 6 |
| not in git (appended last) | 0 |

### Replay failures

**2-7b1-watchman-automation.sql**
```
psql:/migrations/2-7b1-watchman-automation.sql:53: NOTICE:  relation "watchman_passes" already exists, skipping
psql:/migrations/2-7b1-watchman-automation.sql:56: NOTICE:  relation "watchman_passes_started_idx" already exists, skipping
psql:/migrations/2-7b1-watchman-automation.sql:58: NOTICE:  relation "watchman_passes_status_idx" already exists, skipping
psql:/migrations/2-7b1-watchman-automation.sql:81: NOTICE:  relation "watchman_config" already exists, skipping
psql:/migrations/2-7b1-watchman-automation.sql:539: ERROR:  cannot drop columns from view
```

**4c-steward-dispatch.sql**
```
psql:/migrations/4c-steward-dispatch.sql:37: ERROR:  cannot drop function steward_tick() because extension pg_ai_stewards requires it
HINT:  You can drop extension pg_ai_stewards instead.
```

**4d-steward-realign.sql**
```
psql:/migrations/4d-steward-realign.sql:66: ERROR:  cannot drop function steward_tick() because extension pg_ai_stewards requires it
HINT:  You can drop extension pg_ai_stewards instead.
```

**5d5-tools-off-and-templates.sql**
```
psql:/migrations/5d5-tools-off-and-templates.sql:30: ERROR:  check constraint "gate_prompts_id_check" of relation "gate_prompts" is violated by some row
```

**5e2-sabbath.sql**
```
psql:/migrations/5e2-sabbath.sql:18: NOTICE:  constraint "gate_prompts_id_check" of relation "gate_prompts" does not exist, skipping
psql:/migrations/5e2-sabbath.sql:22: ERROR:  check constraint "gate_prompts_id_check" of relation "gate_prompts" is violated by some row
```

**es11-gateway-upstream-cost.sql**
```
psql:/migrations/es11-gateway-upstream-cost.sql:31: ERROR:  cannot drop function record_cost_event(uuid,integer,text,text,integer,integer,integer,integer,text,text) because extension pg_ai_stewards requires it
HINT:  You can drop extension pg_ai_stewards instead.
```

## 2. Parity vs live

| object | live | scratch | missing in scratch | extra in scratch | def mismatch |
|---|---|---|---|---|---|
| functions | 313 | 311 | 3 | 1 | 20 |
| columns | 763 | 763 | 0 | 0 | 0 |
| views | 11 | 11 | 0 | 0 | 0 |
| triggers | 36 | 36 | 0 | 0 | 0 |

### functions detail

Missing in scratch (live-only — never landed in repo files, or replay failure):
- `ct2_echo_tool(p_args jsonb)`
- `record_cost_event(p_work_item_id uuid, p_attempt_seq integer, p_provider text, p_model text, p_input_tokens integer, p_output_tokens integer, p_cache_write_tokens integer, p_cache_read_tokens integer, p_notes text)`
- `record_cost_event(p_work_item_id uuid, p_attempt_seq integer, p_provider text, p_model text, p_input_tokens integer, p_output_tokens integer, p_cache_write_tokens integer, p_cache_read_tokens integer, p_session_id text, p_notes text, p_upstream_micro bigint)`

Extra in scratch (repo-only — never applied live?):
- `record_cost_event(p_work_item_id uuid, p_attempt_seq integer, p_provider text, p_model text, p_input_tokens integer, p_output_tokens integer, p_cache_write_tokens integer, p_cache_read_tokens integer, p_session_id text, p_notes text)`

Definition mismatch (replay produced a different definition than live):
- `acknowledge_finding(p_finding_id bigint, p_resolution text, p_actor text)`
- `complete_todo(p_ref text, p_session text, p_status text)`
- `context_for_hop(p_seed_slug text)`
- `context_for(p_slug text, p_depth integer)`
- `create_todo(p_parent_kind text, p_parent_slug text, p_title text, p_body text, p_slug text, p_session text)`
- `declared_edges(p_slug text)`
- `dry_run_chat(p_agent_family text, p_model text, p_session_id text, p_user_input text)`
- `evaluate_gate(p_work_item_id uuid)`
- `generate_scenarios(p_work_item_id uuid)`
- `import_workstream(p_id text, p_name text, p_description text, p_status text)`
- `link_declared_edges(p_slug text, p_frontmatter jsonb)`
- `link_phase_to_doc(p_phase_slug text, p_parent_doc_slug text)`
- `list_todos(p_parent_kind text, p_parent_slug text, p_status text)`
- `record_finding(p_slug text, p_kind text, p_message text, p_severity text, p_suggested_action text, p_related_slugs text[], p_pass_id text, p_actor text)`
- `record_verdict(p_slug text, p_verdict text, p_reasoning text, p_model text, p_tokens_in integer, p_tokens_out integer, p_pass_id text, p_actor text)`
- `study_history(p_slug text)`
- `todo_rollup_audit()`
- `touch_todo()`
- `verify_work_item(p_work_item_id uuid)`
- `workstream_proposals(p_ws_id text)`

## 3. Verify/smoke files (hermetic scratch)

Failures here are NOT all regressions: files that assume live data, a running
bgworker, or provider HTTP will fail on a hermetic scratch by design. Triage column is the point.

| file | result |
|---|---|
| `verify-1-6-1-reaper-check.sql` | PASS |
| `verify-1-6-1-reaper-setup.sql` | PASS |
| `verify-1-6-1.sql` | PASS |
| `verify-2-1.sql` | PASS |
| `verify-2-2.sql` | FAIL |
| `verify-2-3.sql` | FAIL |
| `verify-2-7b1-inverse.sql` | PASS |
| `verify-2-7b2-decision.sql` | PASS |
| `verify-2-7b3-budget.sql` | PASS |
| `verify-3c2-inverse.sql` | FAIL |
| `verify-3e2-2.sql` | FAIL |
| `verify-4a-steward.sql` | PASS |
| `verify-4a.sql` | PASS |
| `verify-4b.sql` | FAIL |
| `verify-4c.sql` | FAIL |
| `verify-5a.sql` | FAIL |
| `verify-loop.sql` | PASS |
| `test-gate-e2e.sql` | FAIL |
| `smoke/g45-cleanup.sql` | PASS |
| `smoke/g45-lesson-resolution-triggers.sql` | FAIL |
| `smoke/h1-0-overrides.sql` | FAIL |
| `smoke/h1-2-pipeline-and-dispatch.sql` | FAIL |
| `smoke/h1-4-first-real-run.sql` | FAIL |
| `smoke/h1-4-halt.sql` | PASS |
| `smoke/h1-4-recovery.sql` | FAIL |
| `smoke/h1-5a-soft-fail.sql` | FAIL |
| `smoke/h1-5d-materialize.sql` | FAIL |
| `smoke/h1-5d-retry.sql` | FAIL |
| `smoke/h1-6-1-maturity-hook.sql` | FAIL |
| `smoke/h1-6-2-verified-trigger.sql` | FAIL |
| `smoke/h1-6-5-full-auto-e2e.sql` | FAIL |
| `smoke/h1-6-6-strip-prefix.sql` | PASS |
| `smoke/h1-7-validate-substrate-reflection.sql` | FAIL |
| `smoke/h2-validate-context-gather.sql` | FAIL |
| `smoke/h3-5-enqueue-proposed-work-items.sql` | FAIL |
| `smoke/h3-6-planning-pipeline-e2e.sql` | FAIL |
| `smoke/h3-followup-3-sc-ai-literacy-mvp.sql` | FAIL |
| `smoke/i2-fk-smoke.sql` | PASS |
| `smoke/i4-agent-proposal-smoke.sql` | PASS |
| `smoke/i4-real-e2e.sql` | FAIL |
| `smoke/i6-schema-migration-smoke.sql` | PASS |
| `smoke/j10-smoke-dispatch.sql` | FAIL |
| `smoke/j11-smoke-gate.sql` | FAIL |
| `smoke/j11-smoke-gemini-cost.sql` | FAIL |
| `smoke/j12-smoke-preflight.sql` | PASS |
| `smoke/j8-smoke-cleanup.sql` | FAIL |
| `smoke/j8-smoke-object.sql` | FAIL |
| `smoke/j8-smoke-override.sql` | FAIL |
| `smoke/j9-smoke-subset.sql` | FAIL |
| `smoke/j9-smoke-unknown-lens.sql` | PASS |
| `smoke/m2-smoke-substitution.sql` | PASS |
| `smoke/smoke-es7-gating.sql` | FAIL |
| `smoke/smoke-es7.sql` | PASS |
| `smoke/smoke-es8.sql` | PASS |

### Failure tails

**verify-2-2.sql**
```

=== Test 6: study_citations_resolved end-to-end ===
 cited_uri | anchor_text | verse_count | first_ref | first_text_preview 
-----------+-------------+-------------+-----------+--------------------
(0 rows)

=== Test 7: corpus state (404s cached, no retries) ===
psql:/migrations/verify-2-2.sql:86: ERROR:  division by zero
```

**verify-2-3.sql**
```

 edges_after_refresh 
---------------------
                   0
(1 row)

psql:/migrations/verify-2-3.sql:30: ERROR:  unhandled cypher(cstring) function call
DETAIL:  stewards_graph
```

**verify-3c2-inverse.sql**
```
CREATE FUNCTION
DELETE 0
DELETE 0
DELETE 0

=== TRIAL 1: trigger PRESENT ΓÇö auto-advance, rollup ===
psql:/migrations/verify-3c2-inverse.sql:104: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
```

**verify-3e2-2.sql**
```
 pg_sleep 
----------
 
(1 row)
psql:/migrations/verify-3e2-2.sql:44: ERROR:  syntax error at or near ":"

LINE 7:  WHERE id = :enqueued_id;
                    ^
```

**verify-4b.sql**
```


=== G. NON-REGRESSION: synthetic dispatch test on a pending work_item ===
    Setup: create work_item, dispatch via existing 2-arg signature (omit p_allow_failed_status)
    Expected: returns work_id, work_queue gets a chat row with provider opencode_go
psql:/migrations/verify-4b.sql:83: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
PL/pgSQL function inline_code_block line 9 at assignment
```

**verify-4c.sql**
```
       0
(1 row)


=== B. Setup: create a study-write work_item in failed state ===
psql:/migrations/verify-4c.sql:35: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
PL/pgSQL function inline_code_block line 5 at assignment
```

**verify-5a.sql**
```
 Hello steward, you are gate!
(1 row)


=== B. Setup synthetic work_item with simulated stage output ===
psql:/migrations/verify-5a.sql:32: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
PL/pgSQL function inline_code_block line 5 at assignment
```

**test-gate-e2e.sql**
```
DELETE 0
psql:/migrations/test-gate-e2e.sql:37: ERROR:  null value in column "intent_id" of relation "work_items" violates not-null constraint
DETAIL:  Failing row contains (653e35ab-ff33-4cdf-afe6-cf9ed997f24f, gate-test-e2e-1, study-write, outline, in_progress, {"topic": "D&C 130:18-19 ΓÇö intelligence in the resurrection"}, {"outline": {"output": "I. Opening hook ΓÇö what does it mean fo..., {}, null, 0, 0, gate-test, null, 2026-06-09 18:43:29.298601+00, 2026-06-09 18:43:29.298601+00, null, 0, null, null, null, normal, null, null, null, 0, 0, null, null, null, null, null, raw, [], 0, null, null, null, null, null, null, null, null, null, human, null, null, null, null).
```

**smoke/g45-lesson-resolution-triggers.sql**
```
psql:/migrations/smoke/g45-lesson-resolution-triggers.sql:44: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
SQL statement "SELECT stewards.work_item_create('study-write',
        '{"binding_question":"smoke G45"}'::jsonb,
        'smoke-g45', 'human', NULL, NULL)"
PL/pgSQL function inline_code_block line 10 at SQL statement
```

**smoke/h1-0-overrides.sql**
```
psql:/migrations/smoke/h1-0-overrides.sql:77: NOTICE:  study-write pipeline defaults: sabbath=t atonement=f
psql:/migrations/smoke/h1-0-overrides.sql:77: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
SQL statement "SELECT stewards.work_item_create('study-write',
        '{"binding_question":"H.1.0 smoke A ΓÇö inherits"}'::jsonb,
        'h10-smoke-a', 'human', NULL, NULL)"
PL/pgSQL function inline_code_block line 17 at SQL statement
```

**smoke/h1-2-pipeline-and-dispatch.sql**
```
psql:/migrations/smoke/h1-2-pipeline-and-dispatch.sql:97: NOTICE:  pipeline: family=research-write stages=4 sabbath=t atone=t template=research/<slug>.md
psql:/migrations/smoke/h1-2-pipeline-and-dispatch.sql:97: NOTICE:  gather tools_disabled=false   synthesize tools_disabled=false   review tools_disabled=true
psql:/migrations/smoke/h1-2-pipeline-and-dispatch.sql:97: NOTICE:  work_item created: 9654a7c1-fff9-4f01-a5b2-deb2d22d8a07
psql:/migrations/smoke/h1-2-pipeline-and-dispatch.sql:97: ERROR:  no agent variant resolved: family=research model=qwen3.6-plus
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 131 at assignment
PL/pgSQL function inline_code_block line 52 at assignment
```

**smoke/h1-4-first-real-run.sql**
```
psql:/migrations/smoke/h1-4-first-real-run.sql:39: NOTICE:  work_item created: e37f2892-ca6d-450d-9384-4c12ce9f42d5 (slug=ai-tools-weekly-2026-05-11)
psql:/migrations/smoke/h1-4-first-real-run.sql:39: ERROR:  no agent variant resolved: family=research model=qwen3.6-plus
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 131 at assignment
PL/pgSQL function inline_code_block line 34 at assignment
```

**smoke/h1-4-recovery.sql**
```
DELETE 0
DELETE 0
DELETE 0
COMMIT
psql:/migrations/smoke/h1-4-recovery.sql:45: ERROR:  no agent variant resolved: family=research model=qwen3.6-plus
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 131 at assignment
PL/pgSQL function inline_code_block line 17 at assignment
```

**smoke/h1-5a-soft-fail.sql**
```
psql:/migrations/smoke/h1-5a-soft-fail.sql:22: NOTICE:  disabled-server enqueue returned: 21
psql:/migrations/smoke/h1-5a-soft-fail.sql:22: ERROR:  expected NULL, got 21
CONTEXT:  PL/pgSQL function inline_code_block line 7 at RAISE
```

**smoke/h1-5d-materialize.sql**
```
        'research/ai-tools-weekly-2026-05-11.md',
        'create',
        v_clean_md,
        v_wi_id::text,
        'work_item'
    )
    RETURNING id"
PL/pgSQL function inline_code_block line 24 at SQL statement
```

**smoke/h1-5d-retry.sql**
```
psql:/migrations/smoke/h1-5d-retry.sql:29: ERROR:  no agent variant resolved: family=research model=qwen3.6-plus
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 131 at assignment
PL/pgSQL function inline_code_block line 25 at assignment
```

**smoke/h1-6-1-maturity-hook.sql**
```
psql:/migrations/smoke/h1-6-1-maturity-hook.sql:53: NOTICE:  [start] maturity = raw (expect raw)
psql:/migrations/smoke/h1-6-1-maturity-hook.sql:53: NOTICE:  [after gather] maturity = raw (expect researched)
psql:/migrations/smoke/h1-6-1-maturity-hook.sql:53: ERROR:  expected researched, got raw
CONTEXT:  PL/pgSQL function inline_code_block line 22 at RAISE
```

**smoke/h1-6-2-verified-trigger.sql**
```
psql:/migrations/smoke/h1-6-2-verified-trigger.sql:134: NOTICE:  on_maturity_verified: enqueue_work_item_file pwid=2 for work_item=ede4bb81-d860-4857-96d0-41eba26a0bdf
psql:/migrations/smoke/h1-6-2-verified-trigger.sql:134: NOTICE:  [case 1: auto_mat ON, sabbath OFF] pending_file_writes count: 0 ΓåÆ 1
psql:/migrations/smoke/h1-6-2-verified-trigger.sql:134: NOTICE:  [case 2: auto_mat OFF (default)] pending_file_writes count: 0 ΓåÆ 0 (expect no change)
psql:/migrations/smoke/h1-6-2-verified-trigger.sql:134: NOTICE:  on_maturity_verified: sabbath_dispatch failed: no agent variant resolved: family=plan model=qwen3.6-plus
psql:/migrations/smoke/h1-6-2-verified-trigger.sql:134: NOTICE:  [case 3: sabbath ON (default), auto_mat OFF] sabbath work_queue count: 0 ΓåÆ 0
psql:/migrations/smoke/h1-6-2-verified-trigger.sql:134: ERROR:  expected new sabbath work_queue row
CONTEXT:  PL/pgSQL function inline_code_block line 96 at RAISE
```

**smoke/h1-6-5-full-auto-e2e.sql**
```
UPDATE 1
psql:/migrations/smoke/h1-6-5-full-auto-e2e.sql:37: ERROR:  no agent variant resolved: family=research model=qwen3.6-plus
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 131 at assignment
PL/pgSQL function inline_code_block line 24 at assignment
```

**smoke/h1-7-validate-substrate-reflection.sql**
```
 6f178b9c-be85-4eed-b623-9ec2a80dc993 | h1-7-validation-substrate-reflection-2
(1 row)

INSERT 0 1
psql:/migrations/smoke/h1-7-validate-substrate-reflection.sql:56: ERROR:  resolve_template_path: path stage_results.context_gather.output not resolvable; stopped at context_gather
CONTEXT:  PL/pgSQL function resolve_template_path(jsonb,jsonb,text) line 28 at RAISE
PL/pgSQL function render_stage_input(uuid) line 34 at assignment
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 119 at assignment
```

**smoke/h2-validate-context-gather.sql**
```
           work_item_create           
--------------------------------------
 4430d30d-044c-44ef-a0ec-69b6e0a69c9b
(1 row)

psql:/migrations/smoke/h2-validate-context-gather.sql:29: ERROR:  no agent variant resolved: family=research model=qwen3.6-plus
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function work_item_dispatch_stage(uuid,text,boolean) line 131 at assignment
```

**smoke/h3-5-enqueue-proposed-work-items.sql**
```
]$JSON$
            )
        ),
        'planned',
        'pg-ai-stewards'  -- project_association to test inheritance
    )
    RETURNING id"
PL/pgSQL function inline_code_block line 11 at SQL statement
```

**smoke/h3-6-planning-pipeline-e2e.sql**
```
           work_item_create           
--------------------------------------
 25665fd7-a7e6-4967-9fb2-8173a48cc2be
(1 row)

psql:/migrations/smoke/h3-6-planning-pipeline-e2e.sql:39: ERROR:  insert or update on table "work_items" violates foreign key constraint "work_items_project_association_fkey"
DETAIL:  Key (project_association)=(pg-ai-stewards) is not present in table "projects".
```

**smoke/h3-followup-3-sc-ai-literacy-mvp.sql**
```
           work_item_create           
--------------------------------------
 6139f866-6433-4c48-bb5f-9f2b3fd8bbbb
(1 row)

psql:/migrations/smoke/h3-followup-3-sc-ai-literacy-mvp.sql:49: ERROR:  insert or update on table "work_items" violates foreign key constraint "work_items_project_association_fkey"
DETAIL:  Key (project_association)=(space-center) is not present in table "projects".
```

**smoke/i4-real-e2e.sql**
```
        ('i4-real-e2e-smoke-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         jsonb_build_object('draft', v_out),
         jsonb_build_object('validate', jsonb_build_object('output', v_out::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id"
PL/pgSQL function inline_code_block line 16 at SQL statement
```

**smoke/j10-smoke-dispatch.sql**
```
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id"
PL/pgSQL function start_brainstorm(text,text,text,text,text,bigint,jsonb,text[]) line 109 at SQL statement
```

**smoke/j11-smoke-gate.sql**
```
 gemini_exceeded | opencode_exceeded 
-----------------+-------------------
 f               | f
(1 row)

psql:/migrations/smoke/j11-smoke-gate.sql:32: ERROR:  work_item_create: no intent_id supplied and no scripture-study intent seeded
CONTEXT:  PL/pgSQL function work_item_create(text,jsonb,text,text,integer,uuid) line 22 at RAISE
PL/pgSQL function inline_code_block line 4 at assignment
```

**smoke/j11-smoke-gemini-cost.sql**
```
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id"
PL/pgSQL function start_brainstorm(text,text,text,text,text,bigint,jsonb,text[]) line 109 at SQL statement
```

**smoke/j8-smoke-cleanup.sql**
```
 id | slug | status 
----+------+--------
(0 rows)

UPDATE 0
psql:/migrations/smoke/j8-smoke-cleanup.sql:35: ERROR:  column wi.cost_micro does not exist
LINE 4:        wi.cost_micro,
               ^
```

**smoke/j8-smoke-object.sql**
```
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id"
PL/pgSQL function start_brainstorm(text,text,text,text,text,bigint,jsonb,text[]) line 109 at SQL statement
```

**smoke/j8-smoke-override.sql**
```
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id"
PL/pgSQL function start_brainstorm(text,text,text,text,text,bigint,jsonb,text[]) line 109 at SQL statement
```

**smoke/j9-smoke-subset.sql**
```
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id"
PL/pgSQL function start_brainstorm(text,text,text,text,text,bigint,jsonb,text[]) line 109 at SQL statement
```

**smoke/smoke-es7-gating.sql**
```
psql:/migrations/smoke/smoke-es7-gating.sql:34: ERROR:  no agent variant resolved: family=research model=kimi-k2.6
CONTEXT:  PL/pgSQL function dry_run_chat(text,text,text,text) line 8 at RAISE
PL/pgSQL function dry_run_chat(text,text,text,text,text) line 5 at RETURN
PL/pgSQL function chat_post_internal(text,text,text,text) line 75 at assignment
SQL statement "SELECT stewards.chat_post_internal(
            parent_family, parent_model, parent_session, parent_provider
        )"
PL/pgSQL function tool_dispatch_complete_waiting() line 113 at SQL statement
```
