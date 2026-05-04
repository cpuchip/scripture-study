//! pg_ai_stewards — Phase 1, step 2.
//!
//! Scope of this revision:
//!   1. Bgworker registered via `shared_preload_libraries`.
//!   2. `stewards.work_queue` table for asynchronous work.
//!   3. `stewards.enqueue(kind, provider, payload)` — produces work.
//!   4. Bgworker polls every 500ms, claims one row at a time using
//!      `FOR UPDATE SKIP LOCKED`, runs a stub "echo" provider,
//!      writes result back, `NOTIFY stewards_done '<id>'`.
//!   5. Provider registry parsed from `STEWARDS_PROVIDER_*` env vars
//!      at worker startup. Visible (without secrets) via
//!      `stewards.providers_loaded()`.
//!
//! Out of scope:
//!   - Real HTTP provider calls (tokio + reqwest land in step 6/7).
//!   - LISTEN-driven wake-up (we poll; NOTIFY on completion still works).
//!   - Brain schema (step 3).

use pgrx::bgworkers::*;
use pgrx::prelude::*;
use std::sync::OnceLock;
use std::time::Duration;

::pgrx::pg_module_magic!();

// ---------------------------------------------------------------------------
// Schema bootstrap (creates table on `CREATE EXTENSION`)
// ---------------------------------------------------------------------------

extension_sql!(
    r#"
    -- The `stewards` schema is declared in pg_ai_stewards.control;
    -- pgrx creates it automatically when the extension is installed.

    CREATE TABLE stewards.work_queue (
        id          bigserial PRIMARY KEY,
        kind        text NOT NULL,
        provider    text NOT NULL,
        status      text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'in_progress', 'done', 'error')),
        payload     jsonb NOT NULL DEFAULT '{}'::jsonb,
        result      jsonb,
        error       text,
        created_at  timestamptz NOT NULL DEFAULT now(),
        claimed_at  timestamptz,
        done_at     timestamptz
    );

    -- Index supporting the worker's claim query.
    CREATE INDEX work_queue_pending_idx
        ON stewards.work_queue (created_at)
        WHERE status = 'pending';
    "#,
    name = "create_work_queue",
);

extension_sql!(
    r#"
    -- ============================================================
    -- Step 3: brain replacement schema.
    --
    -- Single brain_entries table with a category enum + jsonb props,
    -- chosen over six per-category tables because it matches how
    -- chromem-go stores them today and keeps the migrator simple.
    -- Category-specific fields (name, follow_ups, status, due_date,
    -- mood, gratitude, ...) all live in `props`.
    --
    -- Categories enumerated in the CHECK constraint below come from
    -- scripts/brain/internal/classifier/classifier.go (the six the
    -- LLM classifier emits) plus 'inbox' (the unclassified default
    -- set by classifier.go and web/server.go). Read from code per
    -- the data-safety checklist; do NOT add categories from memory.
    -- ============================================================

    CREATE TABLE stewards.brain_entries (
        id              text PRIMARY KEY DEFAULT gen_random_uuid()::text,
        category        text NOT NULL
                        CHECK (category IN
                            ('people','projects','ideas','actions',
                             'study','journal','inbox')),
        title           text NOT NULL,
        body            text NOT NULL DEFAULT '',
        props           jsonb NOT NULL DEFAULT '{}'::jsonb,

        -- Provenance + classification metadata
        source          text NOT NULL DEFAULT 'cli',
        confidence      real NOT NULL DEFAULT 0.0,
        needs_review    boolean NOT NULL DEFAULT false,
        quarantined     boolean NOT NULL DEFAULT false,
        original_body   text,

        -- Embedding (populated async by bgworker; see embed trigger
        -- below + step 6/7 for the actual provider call).
        embedding       vector(768),
        embedded_at     timestamptz,
        embedded_model  text,
        embedding_error text,

        -- Full-text search column maintained automatically.
        body_tsv        tsvector
                        GENERATED ALWAYS AS (
                            to_tsvector('english',
                                coalesce(title, '') || ' ' || coalesce(body, ''))
                        ) STORED,

        created_at      timestamptz NOT NULL DEFAULT now(),
        updated_at      timestamptz NOT NULL DEFAULT now()
    );

    CREATE INDEX brain_entries_category_idx
        ON stewards.brain_entries (category);
    CREATE INDEX brain_entries_created_idx
        ON stewards.brain_entries (created_at DESC);
    CREATE INDEX brain_entries_needs_review_idx
        ON stewards.brain_entries (needs_review)
        WHERE needs_review = true;
    CREATE INDEX brain_entries_fts_idx
        ON stewards.brain_entries USING gin (body_tsv);
    CREATE INDEX brain_entries_props_idx
        ON stewards.brain_entries USING gin (props);

    -- HNSW index for cosine similarity. NULL embeddings are skipped
    -- by the index naturally; we filter them in queries too.
    CREATE INDEX brain_entries_embedding_idx
        ON stewards.brain_entries
        USING hnsw (embedding vector_cosine_ops);

    -- Tags split out for query / index efficiency. Mirrors the
    -- existing brain SQLite layout.
    CREATE TABLE stewards.brain_entry_tags (
        entry_id text NOT NULL
                 REFERENCES stewards.brain_entries(id) ON DELETE CASCADE,
        tag      text NOT NULL,
        PRIMARY KEY (entry_id, tag)
    );
    CREATE INDEX brain_entry_tags_tag_idx
        ON stewards.brain_entry_tags (tag);

    CREATE TABLE stewards.brain_subtasks (
        id          bigserial PRIMARY KEY,
        entry_id    text NOT NULL
                    REFERENCES stewards.brain_entries(id) ON DELETE CASCADE,
        body        text NOT NULL,
        done        boolean NOT NULL DEFAULT false,
        sort_order  int NOT NULL DEFAULT 0,
        created_at  timestamptz NOT NULL DEFAULT now(),
        updated_at  timestamptz NOT NULL DEFAULT now()
    );
    CREATE INDEX brain_subtasks_entry_idx
        ON stewards.brain_subtasks (entry_id, sort_order);

    -- Snapshot history. Captures (title, category, body, props) at
    -- mutation time; the touch_updated_at trigger inserts here on UPDATE.
    CREATE TABLE stewards.brain_versions (
        id          bigserial PRIMARY KEY,
        entry_id    text NOT NULL
                    REFERENCES stewards.brain_entries(id) ON DELETE CASCADE,
        title       text NOT NULL,
        category    text NOT NULL,
        body        text NOT NULL,
        props       jsonb NOT NULL DEFAULT '{}'::jsonb,
        changed_by  text NOT NULL DEFAULT 'system',
        changed_at  timestamptz NOT NULL DEFAULT now()
    );
    CREATE INDEX brain_versions_entry_idx
        ON stewards.brain_versions (entry_id, changed_at DESC);

    -- ============================================================
    -- Sessions + messages (basic conversation log).
    -- Goal: have something to embed and query end-to-end so step 6
    -- can prove the round-trip on more than a single table.
    -- ============================================================

    CREATE TABLE stewards.sessions (
        id              text PRIMARY KEY DEFAULT gen_random_uuid()::text,
        label           text,
        kind            text NOT NULL DEFAULT 'chat'
                        CHECK (kind IN ('chat','agent','tool','study','dev')),
        created_at      timestamptz NOT NULL DEFAULT now(),
        last_active_at  timestamptz NOT NULL DEFAULT now()
    );

    CREATE TABLE stewards.messages (
        id              bigserial PRIMARY KEY,
        session_id      text NOT NULL
                        REFERENCES stewards.sessions(id) ON DELETE CASCADE,
        role            text NOT NULL
                        CHECK (role IN ('user','assistant','system','tool')),
        content         text NOT NULL DEFAULT '',
        model           text,
        tokens_in       int,
        tokens_out      int,
        -- Reasoning tokens are billed separately by some providers
        -- (kimi-k2.6 via OpenCode reports them under
        -- usage.completion_tokens_details.reasoning_tokens). They
        -- are NOT included in tokens_out, so cost computation must
        -- sum both. Captured here so we don't under-count.
        reasoning_tokens int,
        cost_usd        numeric(10, 6),

        -- Assistant messages may carry tool_calls instead of (or in
        -- addition to) content. Stored verbatim; Phase 1.6's loop
        -- will read this to dispatch tools. Step 7 just records.
        tool_calls      jsonb,
        finish_reason   text,
        tool_call_id    text,        -- set on role='tool' replies

        -- Reasoning fields. Required for echo-back when continuing a
        -- chat with thinking-enabled models (kimi-k2.6, o1-class).
        -- Without these, Moonshot returns 400:
        --   "thinking is enabled but reasoning_content is missing in
        --    assistant tool call message at index N"
        -- Capture both shapes — plain `reasoning` is what OpenRouter
        -- emits; `reasoning_details` is the structured array. We
        -- echo both back on the next request for cross-provider safety.
        reasoning_content text,
        reasoning_details jsonb,

        -- For role='tool' messages: which work_queue tool_dispatch
        -- row produced this. For 'assistant' messages: which 'chat'
        -- work_queue row produced this. NULL for 'user' / 'system'.
        -- Used for trace and to count loop iterations cleanly.
        parent_work_id  bigint REFERENCES stewards.work_queue(id) ON DELETE SET NULL,

        embedding       vector(768),
        embedded_at     timestamptz,
        embedded_model  text,
        embedding_error text,

        created_at      timestamptz NOT NULL DEFAULT now()
    );
    CREATE INDEX messages_session_idx
        ON stewards.messages (session_id, created_at);
    CREATE INDEX messages_embedding_idx
        ON stewards.messages
        USING hnsw (embedding vector_cosine_ops);

    -- ============================================================
    -- Triggers
    -- ============================================================

    -- Bump updated_at AND snapshot the previous version on UPDATE.
    -- Only snapshots when the *content* (title, category, body, props)
    -- actually changed. Embedding writes from the bgworker would
    -- otherwise create one junk brain_versions row per embed.
    CREATE FUNCTION stewards.touch_brain_entry() RETURNS trigger
    LANGUAGE plpgsql AS $func$
    BEGIN
        IF TG_OP = 'UPDATE' THEN
            IF NEW.title    IS DISTINCT FROM OLD.title
               OR NEW.category IS DISTINCT FROM OLD.category
               OR NEW.body     IS DISTINCT FROM OLD.body
               OR NEW.props    IS DISTINCT FROM OLD.props
            THEN
                INSERT INTO stewards.brain_versions
                    (entry_id, title, category, body, props, changed_by)
                VALUES
                    (OLD.id, OLD.title, OLD.category, OLD.body, OLD.props,
                     coalesce(current_setting('stewards.actor', true), 'system'));
                NEW.updated_at := now();
            END IF;
        END IF;
        RETURN NEW;
    END;
    $func$;

    CREATE TRIGGER brain_entries_touch
        BEFORE UPDATE ON stewards.brain_entries
        FOR EACH ROW EXECUTE FUNCTION stewards.touch_brain_entry();

    -- Enqueue an embedding job whenever title/body changes (or row
    -- is inserted). The bgworker (step 6) calls LM Studio's
    -- /v1/embeddings with model nomic-embed-text-v1.5 and writes
    -- the resulting 768-dim vector back to NEW.embedding.
    --
    -- Provider name 'lm_studio' resolves to the registry entry
    -- loaded from STEWARDS_PROVIDER_LM_STUDIO_*. Model name matches
    -- gospel-engine-v2 exactly so vectors are comparable across DBs.
    CREATE FUNCTION stewards.enqueue_brain_embed() RETURNS trigger
    LANGUAGE plpgsql AS $func$
    BEGIN
        IF TG_OP = 'INSERT'
           OR NEW.title IS DISTINCT FROM OLD.title
           OR NEW.body  IS DISTINCT FROM OLD.body
        THEN
            INSERT INTO stewards.work_queue (kind, provider, payload)
            VALUES (
                'embed',
                'lm_studio',
                jsonb_build_object(
                    'target_table', 'brain_entries',
                    'target_id',    NEW.id,
                    'text',         coalesce(NEW.title, '') || E'\n\n' || coalesce(NEW.body, ''),
                    'model',        'nomic-embed-text-v1.5',
                    'dimensions',   768
                )
            );
        END IF;
        RETURN NEW;
    END;
    $func$;

    CREATE TRIGGER brain_entries_enqueue_embed
        AFTER INSERT OR UPDATE OF title, body
        ON stewards.brain_entries
        FOR EACH ROW EXECUTE FUNCTION stewards.enqueue_brain_embed();

    CREATE FUNCTION stewards.touch_message() RETURNS trigger
    LANGUAGE plpgsql AS $func$
    BEGIN
        UPDATE stewards.sessions
        SET last_active_at = now()
        WHERE id = NEW.session_id;
        RETURN NEW;
    END;
    $func$;

    CREATE TRIGGER messages_touch_session
        AFTER INSERT ON stewards.messages
        FOR EACH ROW EXECUTE FUNCTION stewards.touch_message();

    -- ============================================================
    -- Helper SQL functions. Thin wrappers; the brain CLI driver
    -- (step 5) will call these instead of writing raw SQL.
    -- ============================================================

    -- Insert or update a brain entry. Returns the row's id.
    -- If `entry_id` is NULL a new id is generated and a row created;
    -- otherwise the matching row is updated. Tags are replaced wholesale
    -- (delete-then-insert under one transaction).
    CREATE FUNCTION stewards.brain_upsert(
        p_category text,
        p_title    text,
        p_body     text DEFAULT '',
        p_props    jsonb DEFAULT '{}'::jsonb,
        p_tags     text[] DEFAULT NULL,
        p_id       text DEFAULT NULL,
        p_source   text DEFAULT 'cli'
    ) RETURNS text
    LANGUAGE plpgsql AS $func$
    DECLARE
        v_id text;
    BEGIN
        IF p_id IS NULL THEN
            INSERT INTO stewards.brain_entries
                (category, title, body, props, source)
            VALUES
                (p_category, p_title, p_body, p_props, p_source)
            RETURNING id INTO v_id;
        ELSE
            INSERT INTO stewards.brain_entries
                (id, category, title, body, props, source)
            VALUES
                (p_id, p_category, p_title, p_body, p_props, p_source)
            ON CONFLICT (id) DO UPDATE SET
                category = EXCLUDED.category,
                title    = EXCLUDED.title,
                body     = EXCLUDED.body,
                props    = EXCLUDED.props,
                source   = EXCLUDED.source
            RETURNING id INTO v_id;
        END IF;

        IF p_tags IS NOT NULL THEN
            DELETE FROM stewards.brain_entry_tags WHERE entry_id = v_id;
            INSERT INTO stewards.brain_entry_tags (entry_id, tag)
            SELECT v_id, unnest(p_tags);
        END IF;

        RETURN v_id;
    END;
    $func$;

    -- Full-text search. Returns id, title, category, ts_rank score.
    CREATE FUNCTION stewards.brain_search_text(
        p_query    text,
        p_category text DEFAULT NULL,
        p_limit    int DEFAULT 20
    ) RETURNS TABLE (
        id       text,
        title    text,
        category text,
        rank     real
    )
    LANGUAGE sql STABLE AS $func$
        SELECT e.id, e.title, e.category,
               ts_rank(e.body_tsv, plainto_tsquery('english', p_query)) AS rank
        FROM stewards.brain_entries e
        WHERE e.body_tsv @@ plainto_tsquery('english', p_query)
          AND (p_category IS NULL OR e.category = p_category)
          AND NOT e.quarantined
        ORDER BY rank DESC
        LIMIT p_limit;
    $func$;

    -- Vector search. Caller passes a 768-dim embedding (computed
    -- elsewhere in step 3; in step 6 a sibling helper will accept
    -- raw text and route through Ollama via the work queue).
    CREATE FUNCTION stewards.brain_search_vec(
        p_embedding vector(768),
        p_category  text DEFAULT NULL,
        p_limit     int DEFAULT 20
    ) RETURNS TABLE (
        id       text,
        title    text,
        category text,
        distance real
    )
    LANGUAGE sql STABLE AS $func$
        SELECT e.id, e.title, e.category,
               (e.embedding <=> p_embedding)::real AS distance
        FROM stewards.brain_entries e
        WHERE e.embedding IS NOT NULL
          AND (p_category IS NULL OR e.category = p_category)
          AND NOT e.quarantined
        ORDER BY e.embedding <=> p_embedding
        LIMIT p_limit;
    $func$;
    "#,
    name = "create_brain_schema",
    requires = ["create_work_queue"],
);

// ---------------------------------------------------------------------------
// Phase 1.6: Tool wrappers (one-arg jsonb in, jsonb out).
// Convention: every sql_fn tool MUST have signature
//   fn(p_args jsonb) RETURNS jsonb
// so the Rust dispatcher is one line: SELECT <fn>($1). Underlying
// SQL fns can have arbitrary signatures; the wrapper unpacks args.
// ---------------------------------------------------------------------------

extension_sql!(
    r#"
    CREATE FUNCTION stewards.brain_search_text_tool(p_args jsonb)
    RETURNS jsonb
    LANGUAGE sql STABLE AS $func$
        SELECT coalesce(jsonb_agg(row_to_json(t)), '[]'::jsonb)
        FROM stewards.brain_search_text(
            p_args->>'query',
            p_args->>'category',
            coalesce((p_args->>'limit')::int, 20)
        ) t;
    $func$;

    -- load_skill_tool: returns the body of the named skill (variant-
    -- resolved against caller model is not done here; we just pick
    -- the longest matching pattern across active rows). The LLM sees
    -- the skill body as the tool reply and folds it into context.
    CREATE FUNCTION stewards.load_skill_tool(p_args jsonb)
    RETURNS jsonb
    LANGUAGE sql STABLE AS $func$
        SELECT coalesce(
            (SELECT to_jsonb(s.body)
               FROM stewards.skills s
              WHERE s.family = p_args->>'name' AND s.active
              ORDER BY length(model_match) DESC, model_match
              LIMIT 1),
            to_jsonb(format('skill not found: %s', p_args->>'name'))
        );
    $func$;
    "#,
    name = "create_tool_wrappers",
    requires = ["create_brain_schema"],
);

// ---------------------------------------------------------------------------
// Phase 1.5: Harness sketch — agents, skills, instructions, tool_defs.
//
// Goal: prove the prompt-assembly + tools[] round-trip BEFORE step 7
// makes a real chat call. `dry_run_chat(family, model, session, input)`
// returns the exact JSON body that would go to /v1/chat/completions
// so we can read it and judge the shape before sending bytes.
//
// Variant-by-glob design: agents/skills/instructions can have multiple
// rows for the same logical "family", differentiated by `model_match`
// (a glob like 'kimi-*'). The catch-all default uses '*', which
// glob-matches everything; resolution picks the LONGEST matching
// pattern, so '*' (length 1) is always the last-resort fallback and
// any specific glob wins over it. Using '*' instead of NULL keeps the
// PK clean and ON CONFLICT honest (PG treats NULL keys as distinct).
// This lets us tune prompts per-model without duplicating workflow
// rules. See `glob_match` and `resolve_*` below.
//
// Tools deliberately do NOT have variants in v1 — a tool's description
// is structural ("what does this do"), not stylistic ("how do I phrase
// this for Qwen"). Stylistic per-model guidance lives in instructions.
// ---------------------------------------------------------------------------

extension_sql!(
    r#"
    -- ============================================================
    -- glob matcher — used by all resolve_* and *_permission helpers.
    --
    -- Converts a shell-style glob ('kimi-*', 'brain_*') to a
    -- Postgres LIKE pattern. We escape `\`, `%`, `_` first so
    -- they match literally, then turn `*` into `%`. `?` (single-char)
    -- is intentionally NOT supported — model names don't need it
    -- and supporting it would require escaping `_` differently.
    -- ============================================================

    CREATE FUNCTION stewards.glob_match(p_pattern text, p_value text)
    RETURNS bool
    LANGUAGE sql IMMUTABLE AS $func$
        SELECT p_value LIKE
            replace(
                replace(
                    replace(
                        replace(p_pattern, '\', '\\'),
                        '%', '\%'),
                    '_', '\_'),
                '*', '%')
    $func$;

    -- ============================================================
    -- Agents — one row per (family, model_match). NULL model_match
    -- is the catch-all default; non-NULL globs win when they match.
    -- ============================================================

    CREATE TABLE stewards.agents (
        family       text NOT NULL,
        model_match  text NOT NULL DEFAULT '*',    -- glob; '*' = default
        description  text NOT NULL,
        mode         text NOT NULL DEFAULT 'primary'
                     CHECK (mode IN ('primary','subagent','all')),
        model_pin    text,                         -- override session model
        prompt       text NOT NULL,                -- agent persona/role
        temperature  real,
        top_p        real,
        steps        int NOT NULL DEFAULT 8,        -- max agentic iterations
        active       bool NOT NULL DEFAULT true,
        created_at   timestamptz NOT NULL DEFAULT now(),
        PRIMARY KEY (family, model_match)
    );

    -- ============================================================
    -- Skills — same variant pattern as agents.
    -- ============================================================

    CREATE TABLE stewards.skills (
        family       text NOT NULL
                     CHECK (family ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
        model_match  text NOT NULL DEFAULT '*',
        description  text NOT NULL
                     CHECK (length(description) BETWEEN 1 AND 1024),
        body         text NOT NULL,
        license      text,
        metadata     jsonb NOT NULL DEFAULT '{}',
        active       bool NOT NULL DEFAULT true,
        created_at   timestamptz NOT NULL DEFAULT now(),
        PRIMARY KEY (family, model_match)
    );

    -- ============================================================
    -- Instructions — flat-merged into the system prompt.
    -- scope = 'global' | 'agent:<family>' | 'session:<id>'
    -- ord = sort order within scope (lower first)
    -- ============================================================

    CREATE TABLE stewards.instructions (
        id            bigserial PRIMARY KEY,
        family        text NOT NULL,                  -- logical name for variant grouping
        model_match   text NOT NULL DEFAULT '*',
        scope         text NOT NULL,
        body          text NOT NULL,
        ord           int  NOT NULL DEFAULT 100,
        active        bool NOT NULL DEFAULT true,
        source_label  text,                            -- e.g. 'project:AGENTS.md'
        created_at    timestamptz NOT NULL DEFAULT now(),
        UNIQUE (family, model_match, scope)
    );
    CREATE INDEX instructions_scope_idx ON stewards.instructions (scope, ord);

    -- ============================================================
    -- Tool defs — what tools an agent can see. No variants in v1.
    -- name follows '<prefix>_<rest>' convention (brain_*, gospel_*).
    -- execute_target is jsonb describing dispatch. v1 supports:
    --   {"kind":"sql_fn","schema":"stewards","name":"brain_search_text"}
    -- Future kinds: 'http', 'subagent', 'mcp'.
    -- ============================================================

    CREATE TABLE stewards.tool_defs (
        name            text PRIMARY KEY
                        CHECK (name ~ '^[a-z][a-z0-9_]*$'),
        description     text NOT NULL,
        args_schema     jsonb NOT NULL,        -- JSON Schema for params
        execute_target  jsonb NOT NULL,
        active          bool NOT NULL DEFAULT true,
        created_at      timestamptz NOT NULL DEFAULT now()
    );

    -- ============================================================
    -- Per-agent permissions for tools and skills.
    -- Glob-matched against tool name / skill family.
    -- Last (longest) matching pattern wins. Default: 'allow' if
    -- no rule exists (mirrors opencode's default-allow behavior).
    -- ============================================================

    CREATE TABLE stewards.agent_tool_perms (
        agent_family  text NOT NULL,
        tool_pattern  text NOT NULL,
        action        text NOT NULL CHECK (action IN ('allow','ask','deny')),
        PRIMARY KEY (agent_family, tool_pattern)
    );

    CREATE TABLE stewards.agent_skill_perms (
        agent_family  text NOT NULL,
        skill_pattern text NOT NULL,
        action        text NOT NULL CHECK (action IN ('allow','ask','deny')),
        PRIMARY KEY (agent_family, skill_pattern)
    );

    -- ============================================================
    -- Tool calls — one row per tool invocation by an agent. Empty
    -- in v1 (no agent loop yet); the table exists so step 7+ can
    -- write to it without a migration.
    -- ============================================================

    CREATE TABLE stewards.tool_calls (
        id            bigserial PRIMARY KEY,
        message_id    bigint REFERENCES stewards.messages(id) ON DELETE CASCADE,
        tool          text NOT NULL,
        args          jsonb NOT NULL,
        result        jsonb,
        status        text NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending','running','done','error')),
        error         text,
        started_at    timestamptz,
        ended_at      timestamptz
    );
    CREATE INDEX tool_calls_message_idx ON stewards.tool_calls (message_id);

    -- ============================================================
    -- Resolution — pick the most-specific row matching this model.
    -- Longest non-NULL pattern wins; NULL is the catch-all fallback.
    -- ============================================================

    CREATE FUNCTION stewards.resolve_agent(p_family text, p_model text)
    RETURNS stewards.agents
    LANGUAGE sql STABLE AS $func$
        SELECT *
        FROM stewards.agents
        WHERE family = p_family
          AND active
          AND stewards.glob_match(model_match, p_model)
        ORDER BY length(model_match) DESC, model_match
        LIMIT 1
    $func$;

    CREATE FUNCTION stewards.resolve_skill(p_family text, p_model text)
    RETURNS stewards.skills
    LANGUAGE sql STABLE AS $func$
        SELECT *
        FROM stewards.skills
        WHERE family = p_family
          AND active
          AND stewards.glob_match(model_match, p_model)
        ORDER BY length(model_match) DESC, model_match
        LIMIT 1
    $func$;

    -- Permission lookup — returns 'allow'|'ask'|'deny'. Default 'allow'.
    CREATE FUNCTION stewards.tool_permission(p_agent text, p_tool text)
    RETURNS text
    LANGUAGE sql STABLE AS $func$
        SELECT coalesce(
            (SELECT action FROM stewards.agent_tool_perms
             WHERE agent_family = p_agent
               AND stewards.glob_match(tool_pattern, p_tool)
             ORDER BY length(tool_pattern) DESC LIMIT 1),
            'allow')
    $func$;

    CREATE FUNCTION stewards.skill_permission(p_agent text, p_skill text)
    RETURNS text
    LANGUAGE sql STABLE AS $func$
        SELECT coalesce(
            (SELECT action FROM stewards.agent_skill_perms
             WHERE agent_family = p_agent
               AND stewards.glob_match(skill_pattern, p_skill)
             ORDER BY length(skill_pattern) DESC LIMIT 1),
            'allow')
    $func$;

    -- ============================================================
    -- Composition — these are the functions step 7 will reuse.
    -- All STABLE / read-only. dry_run_chat is the verification target.
    -- ============================================================

    -- compose_system_prompt: agent.prompt + matching instructions
    -- + (if 'skill' tool permitted) <available_skills> XML block.
    CREATE FUNCTION stewards.compose_system_prompt(
        p_agent_family text, p_model text, p_session_id text
    ) RETURNS text
    LANGUAGE plpgsql STABLE AS $func$
    DECLARE
        v_agent stewards.agents;
        v_prompt text := '';
        v_instructions text;
        v_skills_block text;
    BEGIN
        v_agent := stewards.resolve_agent(p_agent_family, p_model);
        IF v_agent.family IS NULL THEN
            RAISE EXCEPTION
                'no agent variant resolved: family=% model=%',
                p_agent_family, p_model;
        END IF;
        v_prompt := v_agent.prompt;

        -- Append global + agent-scoped instructions (one row per
        -- family, picking the best model match per family).
        SELECT string_agg(body, E'\n\n' ORDER BY ord, family)
        INTO v_instructions
        FROM (
            SELECT DISTINCT ON (family)
                family, body, ord
            FROM stewards.instructions
            WHERE active
              AND scope IN ('global', 'agent:' || p_agent_family)
              AND stewards.glob_match(model_match, p_model)
            ORDER BY family, length(model_match) DESC, model_match
        ) t;
        IF v_instructions IS NOT NULL THEN
            v_prompt := v_prompt || E'\n\n' || v_instructions;
        END IF;

        -- Append <available_skills> if 'skill' tool isn't denied.
        -- Per opencode pattern: skills are advertised here, loaded
        -- on-demand by the agent calling skill({name: 'foo'}).
        IF stewards.tool_permission(p_agent_family, 'skill') <> 'deny' THEN
            SELECT E'\n\n<available_skills>\n' || string_agg(
                '  <skill>' || E'\n'
                || '    <name>' || family || '</name>' || E'\n'
                || '    <description>' || description || '</description>' || E'\n'
                || '  </skill>',
                E'\n'
                ORDER BY family
            ) || E'\n</available_skills>'
            INTO v_skills_block
            FROM (
                SELECT DISTINCT ON (family) family, description
                FROM stewards.skills
                WHERE active
                  AND stewards.glob_match(model_match, p_model)
                  AND stewards.skill_permission(p_agent_family, family) <> 'deny'
                ORDER BY family, length(model_match) DESC, model_match
            ) s;
            IF v_skills_block IS NOT NULL THEN
                v_prompt := v_prompt || v_skills_block;
            END IF;
        END IF;

        RETURN v_prompt;
    END;
    $func$;

    -- compose_messages: [system, ...history, ?user]
    --
    -- Each history row is emitted with the FULL OpenAI message shape
    -- so multi-turn tool flows are valid. Concretely:
    --   - role='user'/'system': {role, content}
    --   - role='assistant' WITHOUT tool_calls: {role, content}
    --   - role='assistant' WITH tool_calls: {role, content, tool_calls}
    --     (content may be empty string when only tool_calls were
    --     emitted; OpenAI requires the field to exist)
    --   - role='tool': {role, tool_call_id, content}
    --     (NO content field omission — must be present and string)
    --
    -- Stripping any of these would cause the provider to 400 with
    -- "messages with role 'tool' must follow an assistant message
    -- with tool_calls" or similar shape errors. Do not simplify.
    CREATE FUNCTION stewards.compose_messages(
        p_agent_family text,
        p_model text,
        p_session_id text,
        p_user_input text DEFAULT NULL
    ) RETURNS jsonb
    LANGUAGE plpgsql STABLE AS $func$
    DECLARE
        v_system  text;
        v_history jsonb;
        v_result  jsonb;
    BEGIN
        v_system := stewards.compose_system_prompt(p_agent_family, p_model, p_session_id);

        SELECT coalesce(jsonb_agg(
            CASE m.role
                WHEN 'tool' THEN jsonb_build_object(
                    'role', 'tool',
                    'tool_call_id', coalesce(m.tool_call_id, ''),
                    'content', m.content
                )
                WHEN 'assistant' THEN
                    -- Build the assistant message field-by-field. We
                    -- ALWAYS include role+content. tool_calls and the
                    -- reasoning fields are added only when present so
                    -- non-tool, non-thinking turns stay minimal.
                    --
                    -- Why both reasoning_content AND reasoning_details:
                    -- Moonshot's request-side validation reads
                    -- `reasoning_content` (string). OpenRouter's pass-
                    -- through reads `reasoning_details` (structured).
                    -- Sending both lets the next request work whether
                    -- the gateway normalizes or not.
                    jsonb_build_object('role', 'assistant', 'content', m.content)
                    || (CASE WHEN m.tool_calls IS NOT NULL
                             THEN jsonb_build_object('tool_calls', m.tool_calls)
                             ELSE '{}'::jsonb END)
                    || (CASE WHEN m.reasoning_content IS NOT NULL
                             THEN jsonb_build_object('reasoning_content', m.reasoning_content)
                             ELSE '{}'::jsonb END)
                    || (CASE WHEN m.reasoning_details IS NOT NULL
                             THEN jsonb_build_object('reasoning_details', m.reasoning_details)
                             ELSE '{}'::jsonb END)
                ELSE
                    jsonb_build_object('role', m.role, 'content', m.content)
            END
            ORDER BY m.created_at, m.id
        ), '[]'::jsonb)
        INTO v_history
        FROM stewards.messages m
        WHERE m.session_id = p_session_id;

        v_result := jsonb_build_array(
            jsonb_build_object('role', 'system', 'content', v_system)
        ) || v_history;

        IF p_user_input IS NOT NULL THEN
            v_result := v_result || jsonb_build_array(
                jsonb_build_object('role', 'user', 'content', p_user_input)
            );
        END IF;

        RETURN v_result;
    END;
    $func$;

    -- compose_tools: OpenAI-shape tools[] array, filtered by perms.
    -- 'ask' tools are included (the loop will handle prompting); only
    -- 'deny' is excluded.
    CREATE FUNCTION stewards.compose_tools(p_agent_family text)
    RETURNS jsonb
    LANGUAGE sql STABLE AS $func$
        SELECT coalesce(jsonb_agg(
            jsonb_build_object(
                'type', 'function',
                'function', jsonb_build_object(
                    'name', t.name,
                    'description', t.description,
                    'parameters', t.args_schema
                )
            )
            ORDER BY t.name
        ), '[]'::jsonb)
        FROM stewards.tool_defs t
        WHERE t.active
          AND stewards.tool_permission(p_agent_family, t.name) <> 'deny'
    $func$;

    -- dry_run_chat: returns the EXACT POST body /v1/chat/completions
    -- would receive — but does NOT send. The verification target.
    CREATE FUNCTION stewards.dry_run_chat(
        p_agent_family text,
        p_model text,
        p_session_id text,
        p_user_input text DEFAULT NULL
    ) RETURNS jsonb
    LANGUAGE plpgsql STABLE AS $func$
    DECLARE
        v_agent stewards.agents;
        v_body  jsonb;
    BEGIN
        v_agent := stewards.resolve_agent(p_agent_family, p_model);
        IF v_agent.family IS NULL THEN
            RAISE EXCEPTION
                'no agent variant resolved: family=% model=%',
                p_agent_family, p_model;
        END IF;

        v_body := jsonb_build_object(
            'model', coalesce(v_agent.model_pin, p_model),
            'messages', stewards.compose_messages(
                p_agent_family, p_model, p_session_id, p_user_input),
            'tools', stewards.compose_tools(p_agent_family)
        );
        IF v_agent.temperature IS NOT NULL THEN
            v_body := v_body || jsonb_build_object('temperature', v_agent.temperature);
        END IF;
        IF v_agent.top_p IS NOT NULL THEN
            v_body := v_body || jsonb_build_object('top_p', v_agent.top_p);
        END IF;

        RETURN v_body || jsonb_build_object(
            '_meta', jsonb_build_object(
                'agent_family', p_agent_family,
                'agent_variant_match', v_agent.model_match,
                'requested_model', p_model,
                'pinned_model', v_agent.model_pin,
                'session_id', p_session_id
            )
        );
    END;
    $func$;
    "#,
    name = "create_harness_schema",
    requires = ["create_brain_schema"],
);

// ---------------------------------------------------------------------------
// Phase 1.5 seed data — minimum to exercise dry_run_chat against
// real-shaped data. Idempotent; safe to re-run.
// ---------------------------------------------------------------------------

extension_sql!(
    r#"
    -- One agent family with a default + a kimi-specific variant
    -- so the resolver actually has to pick. Both share workflow
    -- rules (which live in instructions); only the persona differs.
    INSERT INTO stewards.agents
        (family, model_match, description, mode, prompt, temperature, top_p, steps)
    VALUES
        (
            'stewards-explore', '*',
            'Read-only researcher over the brain and gospel corpus',
            'primary',
            E'You are a careful researcher with access to a Postgres-backed brain of notes and a corpus of scripture.\n\nYour job: when asked a question, search before answering. Cite the brain entry IDs (or scripture references) you actually consulted. If the brain has no entry on a topic, say so plainly — do not invent IDs.',
            0.2, NULL, 8
        ),
        (
            'stewards-explore', 'kimi-*',
            'Read-only researcher (Kimi tuning)',
            'primary',
            E'You are a careful researcher with access to a Postgres-backed brain of notes and a corpus of scripture.\n\nYour job: when asked a question, search before answering. Cite the brain entry IDs (or scripture references) you actually consulted. If the brain has no entry on a topic, say so plainly — do not invent IDs.\n\nKimi-specific: be terse. Prefer 2-3 sentences over paragraphs. Skip throat-clearing.',
            0.2, NULL, 8
        )
    ON CONFLICT (family, model_match) DO NOTHING;

    -- Workflow rules shared across model variants.
    INSERT INTO stewards.instructions
        (family, model_match, scope, body, ord, source_label)
    VALUES
        (
            'honesty', '*', 'global',
            E'## Honesty\n- Read before quoting. Do not paraphrase from memory.\n- If a search returns no results, report that. Do not fabricate.',
            10, 'seed:phase-1.5'
        ),
        (
            'search-budget', '*', 'agent:stewards-explore',
            E'## Search budget\n- Run at most 3 searches before responding. If still uncertain after 3, say what you searched and ask the user to narrow the question.',
            20, 'seed:phase-1.5'
        )
    ON CONFLICT (family, model_match, scope) DO NOTHING;

    -- Two skills lifted in spirit from .github/skills/. Real bodies
    -- would be longer; these prove the shape, not the corpus.
    INSERT INTO stewards.skills
        (family, model_match, description, body, license, metadata)
    VALUES
        (
            'source-verification', '*',
            'Verify scripture and talk quotes against actual source files before quoting',
            E'# Source Verification\n\nBefore using quotation marks around any scripture or talk text, you must have read the actual source row in this session. Training-data memory confabulates.\n\nIf you have not verified, paraphrase using indirect speech ("Paul teaches that...") rather than direct quotation.',
            'MIT', '{"audience":"researcher"}'::jsonb
        ),
        (
            'scripture-linking', '*',
            'Format scripture and conference talk references as workspace-relative links',
            E'# Scripture Linking\n\nScripture references should be cited by their canonical short form (e.g., "Moroni 7:45-48") and accompanied by the brain entry ID if one exists.',
            'MIT', '{"audience":"researcher"}'::jsonb
        )
    ON CONFLICT (family, model_match) DO NOTHING;

    -- Tool defs the agent will actually see. Two for v1: a real
    -- search tool and the special skill-loader. brain_search_vec
    -- is intentionally omitted because the agent can't construct
    -- a vector input directly; a future brain_search_semantic
    -- (text-in, embed-via-worker, vec-search) will replace it.
    INSERT INTO stewards.tool_defs
        (name, description, args_schema, execute_target)
    VALUES
        (
            'brain_search_text',
            'Full-text search over brain entries (notes, ideas, study fragments). Returns ranked matches with id, title, category, and rank score.',
            $j${
                "type": "object",
                "properties": {
                    "query":    {"type": "string", "description": "Search terms (plain language)."},
                    "category": {"type": "string", "description": "Optional category filter.",
                                 "enum": ["inbox","study","journal","action","idea","person","project"]},
                    "limit":    {"type": "integer", "description": "Max results (default 20).", "minimum": 1, "maximum": 100}
                },
                "required": ["query"]
            }$j$::jsonb,
            $j${"kind":"sql_fn","schema":"stewards","name":"brain_search_text_tool"}$j$::jsonb
        ),
        (
            'skill',
            'Load the body of a named skill from the <available_skills> list and return its content into the conversation. Use when a skill''s description matches the task at hand.',
            $j${
                "type": "object",
                "properties": {
                    "name": {"type": "string", "description": "The skill family name (e.g., source-verification)."}
                },
                "required": ["name"]
            }$j$::jsonb,
            $j${"kind":"sql_fn","schema":"stewards","name":"load_skill_tool"}$j$::jsonb
        )
    ON CONFLICT (name) DO NOTHING;

    -- Permissions for stewards-explore: deny anything not brain_*
    -- or skill, allow those explicitly. Demonstrates the glob model.
    INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
    VALUES
        ('stewards-explore', '*',          'deny'),
        ('stewards-explore', 'brain_*',    'allow'),
        ('stewards-explore', 'skill',      'allow')
    ON CONFLICT (agent_family, tool_pattern) DO NOTHING;

    INSERT INTO stewards.agent_skill_perms (agent_family, skill_pattern, action)
    VALUES
        ('stewards-explore', '*', 'allow')
    ON CONFLICT (agent_family, skill_pattern) DO NOTHING;
    "#,
    name = "seed_harness",
    requires = ["create_harness_schema"],
);

// ---------------------------------------------------------------------------
// Step 7 / Phase 1.6: chat round-trip helpers + agent loop enqueuers.
//
// Architecture (Option B — work-item-per-iteration):
//   chat_enqueue      → chat_post_internal → enqueues kind='chat'
//   bgworker chat()   → POSTs, writes assistant message
//   if assistant.tool_calls present AND iteration<steps:
//     phase 3 enqueues kind='tool_dispatch' (carries parent_work_id)
//   bgworker tool_dispatch() → runs each tool, returns ToolsDispatched
//     phase 3 inserts N role='tool' messages, then enqueues kind='chat'
//     (no user input — the messages history already has the new tool
//     replies, compose_messages picks them up automatically)
//   loop terminates when finish_reason='stop'/'length'/'content_filter'
//   OR iteration count >= agent.steps.
//
// Stable-prefix discipline for prompt caching:
//   Every body produced by compose_messages within a session has the
//   same [system, ...prior_history] prefix. Only NEW messages append.
//   This is exactly what OpenAI/Moonshot automatic prompt caching
//   wants. Do not insert anything that varies between system and
//   history (e.g., timestamps, request IDs, freshly-rolled UUIDs).
// ---------------------------------------------------------------------------

extension_sql!(
    r#"
    -- chat_post_internal: compose body from CURRENT session state
    -- (no user input append) and enqueue a chat work item. Used by
    -- chat_enqueue for the first turn AND by tool_dispatch's phase 3
    -- to continue the loop after appending tool replies.
    CREATE FUNCTION stewards.chat_post_internal(
        p_agent_family text,
        p_model        text,
        p_session_id   text,
        p_provider     text
    ) RETURNS bigint
    LANGUAGE plpgsql AS $func$
    DECLARE
        v_body    jsonb;
        v_payload jsonb;
        v_work_id bigint;
    BEGIN
        -- compose with NULL user_input — history already contains
        -- everything we need (the user message was inserted by the
        -- caller of chat_enqueue, or the tool replies were inserted
        -- by tool_dispatch's phase 3).
        v_body := stewards.dry_run_chat(
            p_agent_family, p_model, p_session_id, NULL);

        v_payload := jsonb_build_object(
            'session_id',      p_session_id,
            'agent_family',    p_agent_family,
            'requested_model', p_model,
            'meta',            v_body->'_meta',
            -- Inject `user = <session_id>` so OpenCode (and other
            -- providers that surface per-session billing) can attribute
            -- cost AND so prompt caching keys on a stable user id.
            'body',            (v_body - '_meta')
                               || jsonb_build_object('user', p_session_id)
        );

        INSERT INTO stewards.work_queue (kind, provider, payload)
        VALUES ('chat', p_provider, v_payload)
        RETURNING id INTO v_work_id;

        RETURN v_work_id;
    END;
    $func$;

    -- chat_enqueue: persist user turn + delegate to chat_post_internal.
    -- Caller-facing entry point for starting or continuing a chat
    -- with a new user message. Returns the chat work_queue id.
    CREATE FUNCTION stewards.chat_enqueue(
        p_agent_family text,
        p_model        text,
        p_session_id   text,
        p_user_input   text,
        p_provider     text
    ) RETURNS bigint
    LANGUAGE plpgsql AS $func$
    BEGIN
        INSERT INTO stewards.messages (session_id, role, content, model)
        VALUES (p_session_id, 'user', p_user_input, p_model);

        RETURN stewards.chat_post_internal(
            p_agent_family, p_model, p_session_id, p_provider);
    END;
    $func$;

    -- tool_dispatch_enqueue: called from the bgworker (via SPI) when
    -- a chat response carried tool_calls AND iteration < agent.steps.
    -- Builds the tool_dispatch payload and inserts the work row.
    -- The actual tool execution happens in the bgworker dispatch arm.
    CREATE FUNCTION stewards.tool_dispatch_enqueue(
        p_parent_work_id bigint,
        p_agent_family   text,
        p_model          text,
        p_session_id     text,
        p_provider       text
    ) RETURNS bigint
    LANGUAGE sql AS $func$
        INSERT INTO stewards.work_queue (kind, provider, payload)
        VALUES (
            'tool_dispatch',
            p_provider,
            jsonb_build_object(
                'parent_work_id', p_parent_work_id,
                'agent_family',   p_agent_family,
                'model',          p_model,
                'session_id',     p_session_id
            )
        )
        RETURNING id;
    $func$;

    -- iteration_count: number of assistant messages in this session
    -- since the last user message. Used by the chat handler's phase 3
    -- to compare against agent.steps and decide whether to continue
    -- the loop or stop.
    CREATE FUNCTION stewards.iteration_count(p_session_id text)
    RETURNS int
    LANGUAGE sql STABLE AS $func$
        SELECT count(*)::int FROM stewards.messages
        WHERE session_id = p_session_id
          AND role = 'assistant'
          AND created_at > coalesce(
            (SELECT max(created_at) FROM stewards.messages
             WHERE session_id = p_session_id AND role = 'user'),
            'epoch'::timestamptz
          );
    $func$;

    -- synthesize_tool_failure: when a tool_dispatch row fails BEFORE
    -- the per-tool dispatcher could write its own role='tool' replies
    -- (mode 3 = dispatcher itself errors; mode 4 = bgworker crashed
    -- mid-dispatch and the reaper is cleaning up), this builds the
    -- missing tool replies AND enqueues the continuation chat so the
    -- loop never stalls.
    --
    -- For each tool_call in the parent assistant message that does
    -- NOT already have a matching role='tool' reply in the session
    -- history, insert a synthetic reply with the error message. Then
    -- call chat_post_internal to enqueue the continuation. The model
    -- sees the failure, decides whether to retry-with-different-args
    -- or give up gracefully.
    --
    -- Idempotent: if all tool_calls already have replies (e.g. half
    -- the dispatch succeeded before crash), only the missing ones get
    -- synthesized. If the parent has no tool_calls (caller invoked
    -- this for the wrong row), it's a no-op and returns NULL.
    CREATE FUNCTION stewards.synthesize_tool_failure(
        p_parent_work_id bigint,
        p_agent_family   text,
        p_model          text,
        p_session_id     text,
        p_provider       text,
        p_error          text
    ) RETURNS bigint
    LANGUAGE plpgsql AS $func$
    DECLARE
        v_parent_assistant_id bigint;
        v_tool_calls          jsonb;
        v_tc                  jsonb;
        v_tc_id               text;
        v_synthetic_count     int := 0;
        v_continuation_id     bigint;
    BEGIN
        -- Find the parent assistant message (the one that requested
        -- the tools).
        SELECT id, tool_calls
        INTO v_parent_assistant_id, v_tool_calls
        FROM stewards.messages
        WHERE parent_work_id = p_parent_work_id
          AND role = 'assistant'
        ORDER BY id DESC
        LIMIT 1;

        IF v_parent_assistant_id IS NULL OR v_tool_calls IS NULL
           OR jsonb_array_length(v_tool_calls) = 0 THEN
            RETURN NULL;
        END IF;

        -- For each tool_call, insert a synthetic reply UNLESS one
        -- already exists for that tool_call_id in this session.
        FOR v_tc IN SELECT * FROM jsonb_array_elements(v_tool_calls)
        LOOP
            v_tc_id := v_tc->>'id';
            IF v_tc_id IS NULL THEN CONTINUE; END IF;

            IF NOT EXISTS (
                SELECT 1 FROM stewards.messages
                WHERE session_id = p_session_id
                  AND role = 'tool'
                  AND tool_call_id = v_tc_id
            ) THEN
                INSERT INTO stewards.messages
                    (session_id, role, content,
                     tool_call_id, parent_work_id)
                VALUES (
                    p_session_id, 'tool',
                    jsonb_build_object(
                        'error', p_error,
                        '_synthetic', true,
                        '_reason', 'dispatcher failure; no tool execution occurred'
                    )::text,
                    v_tc_id,
                    p_parent_work_id
                );
                v_synthetic_count := v_synthetic_count + 1;
            END IF;
        END LOOP;

        -- Always enqueue continuation, even if all replies already
        -- existed (caller may be retrying after a previous reaper
        -- run wrote replies but didn't enqueue continuation).
        v_continuation_id := stewards.chat_post_internal(
            p_agent_family, p_model, p_session_id, p_provider);

        RAISE NOTICE 'synthesize_tool_failure: parent=% synthetic=% continuation=%',
            p_parent_work_id, v_synthetic_count, v_continuation_id;
        RETURN v_continuation_id;
    END;
    $func$;

    -- session_status: collapse a session's state into one row.
    -- Useful for any UI/API answering "did this loop finish or stall?".
    -- Joins the latest assistant message's finish_reason with the
    -- latest chat work_queue row's loop_stop_reason and any errored
    -- work_queue rows in the session's parent_work_id chain.
    CREATE VIEW stewards.session_status AS
    SELECT
        s.id AS session_id,
        s.kind,
        s.label,
        -- Latest assistant message in the session
        (SELECT m.finish_reason FROM stewards.messages m
         WHERE m.session_id = s.id AND m.role = 'assistant'
         ORDER BY m.id DESC LIMIT 1) AS last_finish_reason,
        (SELECT m.created_at FROM stewards.messages m
         WHERE m.session_id = s.id AND m.role = 'assistant'
         ORDER BY m.id DESC LIMIT 1) AS last_assistant_at,
        -- Latest chat work_queue row's loop_stop_reason (e.g.
        -- 'steps_exhausted' or 'truncated_tool_calls')
        (SELECT (w.result->>'loop_stop_reason') FROM stewards.work_queue w
         WHERE w.kind = 'chat'
           AND w.payload->>'session_id' = s.id
         ORDER BY w.id DESC LIMIT 1) AS last_loop_stop_reason,
        -- Anything pending or in_progress for this session?
        (SELECT count(*)::int FROM stewards.work_queue w
         WHERE w.payload->>'session_id' = s.id
           AND w.status IN ('pending', 'in_progress')) AS pending_work,
        -- Anything errored?
        (SELECT count(*)::int FROM stewards.work_queue w
         WHERE w.payload->>'session_id' = s.id
           AND w.status = 'error') AS errored_work,
        -- Token + cost rollup across all assistant turns
        (SELECT coalesce(sum(m.tokens_in), 0)::bigint
         FROM stewards.messages m
         WHERE m.session_id = s.id) AS total_tokens_in,
        (SELECT coalesce(sum(m.tokens_out + coalesce(m.reasoning_tokens, 0)), 0)::bigint
         FROM stewards.messages m
         WHERE m.session_id = s.id) AS total_billable_out,
        s.created_at
    FROM stewards.sessions s;

    -- NOTE: an earlier draft included a chat_round_trip() that
    -- enqueued + polled inside one SQL function. That's a footgun:
    -- the SQL function holds an open transaction for the whole loop,
    -- so the work_queue row it just inserted is invisible to the
    -- bgworker (MVCC), AND the still-open tx blocks other writers
    -- on row locks (e.g., the sessions row from the same call).
    -- Removed. Callers should `chat_enqueue()` then either LISTEN
    -- stewards_done or poll work_queue from a separate statement.
    "#,
    name = "create_chat_helpers",
    requires = ["seed_harness"],
);

// ---------------------------------------------------------------------------
// Diagnostic SQL functions
// ---------------------------------------------------------------------------

/// Build version of the extension. First sanity check from step 1.
#[pg_extern]
fn version() -> &'static str {
    env!("CARGO_PKG_VERSION")
}

/// pgrx version this extension was compiled against.
#[pg_extern]
fn pgrx_version() -> &'static str {
    "0.18.0"
}

/// Enqueue a work item. Returns the new row's id.
///
/// `kind` is a free-form string the worker uses to dispatch (e.g.
/// "echo", "embed", "chat"). `provider` is the friendly id of a
/// provider in the registry (e.g. "ollama", "lm_studio", "opencode_go",
/// or "echo" for the stub). `payload` is jsonb passed to the provider.
#[pg_extern]
fn enqueue(kind: &str, provider: &str, payload: pgrx::JsonB) -> i64 {
    Spi::get_one_with_args::<i64>(
        "INSERT INTO stewards.work_queue (kind, provider, payload) \
         VALUES ($1, $2, $3) RETURNING id",
        &[kind.into(), provider.into(), payload.into()],
    )
    .expect("INSERT returned a row")
    .expect("id is non-null")
}

/// List the providers the bgworker loaded from env at startup.
/// Returns one row per provider; **never returns the API key**.
#[pg_extern]
fn providers_loaded() -> TableIterator<
    'static,
    (
        name!(name, String),
        name!(base_url, String),
        name!(default_model, String),
        name!(kind, String),
        name!(has_api_key, bool),
    ),
> {
    let providers = PROVIDER_REGISTRY
        .get()
        .map(|r| r.summary())
        .unwrap_or_default();

    TableIterator::new(providers.into_iter().map(|p| {
        (p.name, p.base_url, p.default_model, p.kind, p.has_api_key)
    }))
}

// ---------------------------------------------------------------------------
// Provider registry (Phase 1: env-var bootstrap, in-process only)
// ---------------------------------------------------------------------------

/// Snapshot of one provider's metadata, minus the secret. Returned
/// from `stewards.providers_loaded()`.
#[derive(Clone, Debug)]
struct ProviderSummary {
    name: String,
    base_url: String,
    default_model: String,
    kind: String,
    has_api_key: bool,
}

#[derive(Clone, Debug)]
struct Provider {
    name: String,
    base_url: String,
    default_model: String,
    kind: String,
    api_key: Option<String>,
}

#[derive(Default, Debug)]
struct ProviderRegistry {
    providers: Vec<Provider>,
}

impl ProviderRegistry {
    /// Parse `STEWARDS_PROVIDER_<NAME>_<FIELD>` env vars into a
    /// registry. Lossy by design: malformed entries are skipped with
    /// a warning rather than aborting the worker.
    fn from_env() -> Self {
        use std::collections::BTreeMap;

        let mut by_name: BTreeMap<String, BTreeMap<String, String>> = BTreeMap::new();

        for (key, value) in std::env::vars() {
            let Some(rest) = key.strip_prefix("STEWARDS_PROVIDER_") else {
                continue;
            };
            // rest = "<NAME>_<FIELD>", where FIELD is one of
            // BASE_URL | API_KEY | DEFAULT_MODEL | KIND
            let Some((name, field)) = split_provider_key(rest) else {
                continue;
            };
            by_name.entry(name).or_default().insert(field, value);
        }

        let mut providers = Vec::with_capacity(by_name.len());
        for (name_upper, fields) in by_name {
            let Some(base_url) = fields.get("BASE_URL").cloned() else {
                pgrx::log!(
                    "stewards: provider '{}' missing BASE_URL, skipping",
                    name_upper
                );
                continue;
            };
            providers.push(Provider {
                name: name_upper.to_lowercase(),
                base_url,
                default_model: fields.get("DEFAULT_MODEL").cloned().unwrap_or_default(),
                kind: fields
                    .get("KIND")
                    .cloned()
                    .unwrap_or_else(|| "openai".to_string()),
                api_key: fields.get("API_KEY").cloned().filter(|s| !s.is_empty()),
            });
        }

        Self { providers }
    }

    fn summary(&self) -> Vec<ProviderSummary> {
        self.providers
            .iter()
            .map(|p| ProviderSummary {
                name: p.name.clone(),
                base_url: p.base_url.clone(),
                default_model: p.default_model.clone(),
                kind: p.kind.clone(),
                has_api_key: p.api_key.is_some(),
            })
            .collect()
    }
}

/// Parse `<NAME>_<FIELD>` where FIELD is one of the four known suffixes.
fn split_provider_key(rest: &str) -> Option<(String, String)> {
    const FIELDS: &[&str] = &["BASE_URL", "API_KEY", "DEFAULT_MODEL", "KIND"];
    for field in FIELDS {
        if let Some(stripped) = rest.strip_suffix(field) {
            if let Some(name) = stripped.strip_suffix('_') {
                if !name.is_empty() {
                    return Some((name.to_string(), field.to_string()));
                }
            }
        }
    }
    None
}

/// Lazily initialized once per bgworker process. Worker reads env on
/// startup and never reloads.
static PROVIDER_REGISTRY: OnceLock<ProviderRegistry> = OnceLock::new();

// ---------------------------------------------------------------------------
// Bgworker registration
// ---------------------------------------------------------------------------

#[pg_guard]
pub extern "C-unwind" fn _PG_init() {
    // Only register the bgworker when we are actually being preloaded
    // via shared_preload_libraries. Otherwise `CREATE EXTENSION` in a
    // database that doesn't preload us would fail.
    if unsafe { !pgrx::pg_sys::process_shared_preload_libraries_in_progress } {
        return;
    }

    // Parse provider registry once, in the postmaster. All backends
    // (and the bgworker) inherit it via fork() copy-on-write, so
    // `stewards.providers_loaded()` works from any psql session and
    // the worker doesn't need to re-parse.
    let registry = ProviderRegistry::from_env();
    pgrx::log!(
        "stewards: postmaster loaded {} provider(s) from env",
        registry.providers.len()
    );
    for p in &registry.providers {
        pgrx::log!(
            "stewards:   provider '{}' kind={} base_url={} model={} api_key={}",
            p.name,
            p.kind,
            p.base_url,
            p.default_model,
            if p.api_key.is_some() { "yes" } else { "no" }
        );
    }
    let _ = PROVIDER_REGISTRY.set(registry);

    BackgroundWorkerBuilder::new("pg_ai_stewards dispatcher")
        .set_function("stewards_dispatcher_main")
        .set_library("pg_ai_stewards")
        .enable_spi_access()
        .set_restart_time(Some(Duration::from_secs(5)))
        .load();
}

/// Worker entry point. Polls `stewards.work_queue` every 500ms,
/// claims one row, runs the stub provider, writes the result back.
#[pg_guard]
#[unsafe(no_mangle)]
pub extern "C-unwind" fn stewards_dispatcher_main(_arg: pg_sys::Datum) {
    BackgroundWorker::attach_signal_handlers(
        SignalWakeFlags::SIGHUP | SignalWakeFlags::SIGTERM,
    );

    let dbname = std::env::var("POSTGRES_DB").unwrap_or_else(|_| "stewards".to_string());
    BackgroundWorker::connect_worker_to_spi(Some(&dbname), None);

    let provider_count = PROVIDER_REGISTRY.get().map(|r| r.providers.len()).unwrap_or(0);
    pgrx::log!(
        "stewards: bgworker entering poll loop (500ms tick); {} provider(s) inherited from postmaster",
        provider_count
    );

    // Stale-claim reaper: any row left in 'in_progress' by a previous
    // bgworker crash is unreachable \u2014 we never reclaim our own
    // claims (that would risk double-side-effects). Mark them errored
    // at startup with a clear message so the caller knows what
    // happened and can decide whether to re-enqueue.
    //
    // For tool_dispatch rows specifically, also call
    // synthesize_tool_failure: write the missing role='tool' replies
    // and enqueue a continuation chat. Otherwise the parent chat's
    // loop stalls forever waiting for tool replies that will never
    // come (Phase 1.6.1).
    let _ = BackgroundWorker::transaction(|| {
        Spi::connect_mut(|client| {
            // Pull the rows we're about to reap so we can synthesize
            // continuations for tool_dispatch ones.
            let stale_rows: Vec<(i64, String, String, serde_json::Value)> = {
                let rows = client.select(
                    "SELECT id, kind, provider, payload \
                     FROM stewards.work_queue \
                     WHERE status = 'in_progress'",
                    None, &[],
                )?;
                rows.into_iter().filter_map(|r| {
                    let id: i64 = r.get(1).ok()??;
                    let kind: String = r.get(2).ok()??;
                    let provider: String = r.get(3).ok()??;
                    let payload: pgrx::JsonB = r.get(4).ok()??;
                    Some((id, kind, provider, payload.0))
                }).collect()
            };

            for (id, kind, provider, payload) in &stale_rows {
                if kind == "tool_dispatch" {
                    if let (Some(parent), Some(session), Some(family), Some(model)) = (
                        payload.get("parent_work_id").and_then(|v| v.as_i64()),
                        payload.get("session_id").and_then(|v| v.as_str()),
                        payload.get("agent_family").and_then(|v| v.as_str()),
                        payload.get("model").and_then(|v| v.as_str()),
                    ) {
                        let synth = client.select(
                            "SELECT stewards.synthesize_tool_failure($1, $2, $3, $4, $5, $6)",
                            Some(1),
                            &[
                                parent.into(),
                                family.to_string().into(),
                                model.to_string().into(),
                                session.to_string().into(),
                                provider.to_string().into(),
                                format!(
                                    "bgworker crashed mid-dispatch on work_item id={}; loop continued via reaper",
                                    id
                                ).into(),
                            ],
                        );
                        if let Err(e) = synth {
                            pgrx::log!(
                                "stewards: reaper synthesize failed for id={}: {}",
                                id, e
                            );
                        } else {
                            pgrx::log!(
                                "stewards: reaper synthesized tool failure for tool_dispatch id={} (parent={})",
                                id, parent
                            );
                        }
                    }
                }
            }

            client.update(
                "UPDATE stewards.work_queue \
                 SET status = 'error', \
                     error  = coalesce(error, '') \
                              || 'bgworker crashed before completion (stale in_progress reaped at startup)', \
                     done_at = now() \
                 WHERE status = 'in_progress'",
                None, &[]
            )?;
            Ok::<(), pgrx::spi::Error>(())
        })
    });

    while BackgroundWorker::wait_latch(Some(Duration::from_millis(500))) {
        if BackgroundWorker::sighup_received() {
            pgrx::log!("stewards: SIGHUP received");
        }

        // Drain whatever's pending. process_one_pending() returns
        // false when the queue is empty, so the loop bounds itself.
        let mut processed = 0u32;
        while process_one_pending() {
            processed += 1;
            // Cap a single tick to avoid starving signal handling.
            if processed >= 16 {
                break;
            }
        }
    }

    pgrx::log!("stewards: bgworker received SIGTERM, exiting");
}

/// Try to claim and process exactly one pending row. Returns true if
/// a row was processed (caller may want to immediately try again),
/// false if the queue was empty.
///
/// The work happens in three phases so we don't hold a row lock
/// across a slow HTTP call (LM Studio first-request model load can
/// be 30s+):
///
///   1. Tx A: claim oldest pending row, mark `in_progress`. Commit.
///   2. No tx: dispatch by kind, possibly making HTTP calls.
///   3. Tx B: write result or error, `NOTIFY stewards_done`. Commit.
fn process_one_pending() -> bool {
    // ----- Phase 1: claim -----
    let claim: Result<Option<(i64, String, String, serde_json::Value)>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect_mut(|client| {
                let claimed = client.update(
                    "WITH next AS ( \
                         SELECT id FROM stewards.work_queue \
                         WHERE status = 'pending' \
                         ORDER BY created_at \
                         FOR UPDATE SKIP LOCKED \
                         LIMIT 1 \
                     ) \
                     UPDATE stewards.work_queue q \
                     SET status = 'in_progress', claimed_at = now() \
                     FROM next \
                     WHERE q.id = next.id \
                     RETURNING q.id, q.kind, q.provider, q.payload",
                    Some(1),
                    &[],
                )?;

                let mut iter = claimed.into_iter();
                let Some(row) = iter.next() else {
                    return Ok(None);
                };

                let id: i64 = row.get(1)?.expect("id non-null");
                let kind: String = row.get(2)?.expect("kind non-null");
                let provider: String = row.get(3)?.expect("provider non-null");
                let payload: pgrx::JsonB = row.get(4)?.expect("payload non-null");
                Ok(Some((id, kind, provider, payload.0)))
            })
        });

    let Some((id, kind, provider, payload)) = (match claim {
        Ok(opt) => opt,
        Err(e) => {
            pgrx::log!("stewards: claim phase errored: {}", e);
            return false;
        }
    }) else {
        return false;
    };

    pgrx::log!(
        "stewards: claimed work_item id={} kind={} provider={}",
        id,
        kind,
        provider
    );

    // ----- Phase 2: dispatch (no tx; HTTP allowed) -----
    let outcome = dispatch(&kind, &provider, &payload);

    // ----- Phase 3: write result -----
    let write: Result<(), pgrx::spi::Error> = BackgroundWorker::transaction(|| {
        Spi::connect_mut(|client| {
            match &outcome {
                Ok(WorkOutcome::Embedded {
                    target_table,
                    target_id,
                    model,
                    embedding_text,
                    dimensions,
                }) => {
                    // Write vector back to the target row. We hard-code
                    // brain_entries for now; messages comes when chat
                    // step lands. The cast to vector(N) validates
                    // dimensions; mismatch raises a Postgres error
                    // that the outer match converts to row error.
                    let update_target = format!(
                        "UPDATE stewards.{} \
                         SET embedding = $2::vector({}), \
                             embedded_at = now(), \
                             embedded_model = $3, \
                             embedding_error = NULL \
                         WHERE id = $1",
                        target_table, dimensions
                    );
                    client.update(
                        &update_target,
                        None,
                        &[
                            target_id.clone().into(),
                            embedding_text.clone().into(),
                            model.clone().into(),
                        ],
                    )?;

                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "embed",
                        "provider": provider,
                        "model": model,
                        "dimensions": dimensions,
                        "target": format!("{}#{}", target_table, target_id),
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::Echo(value)) => {
                    let result_jsonb = pgrx::JsonB(value.clone());
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::Chatted {
                    response,
                    session_id,
                    model,
                    agent_family,
                    requested_model,
                    assistant_content,
                    assistant_tool_calls,
                    reasoning_content,
                    reasoning_details,
                    finish_reason,
                    tokens_in,
                    tokens_out,
                    reasoning_tokens,
                }) => {
                    // Insert the assistant turn. tool_calls and the
                    // reasoning fields are stored verbatim so the
                    // next compose_messages call can echo them back
                    // (required by Moonshot when thinking is enabled).
                    // parent_work_id ties this message back to THIS
                    // work item so tool_dispatch can find it.
                    let tool_calls_jsonb = assistant_tool_calls
                        .clone()
                        .map(pgrx::JsonB);
                    let reasoning_details_jsonb = reasoning_details
                        .clone()
                        .map(pgrx::JsonB);
                    client.update(
                        "INSERT INTO stewards.messages \
                            (session_id, role, content, model, \
                             tool_calls, finish_reason, \
                             tokens_in, tokens_out, reasoning_tokens, \
                             reasoning_content, reasoning_details, \
                             parent_work_id) \
                         VALUES ($1, 'assistant', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
                        None,
                        &[
                            session_id.clone().into(),
                            assistant_content.clone().into(),
                            model.clone().into(),
                            tool_calls_jsonb.into(),
                            finish_reason.clone().into(),
                            (*tokens_in).into(),
                            (*tokens_out).into(),
                            (*reasoning_tokens).into(),
                            reasoning_content.clone().into(),
                            reasoning_details_jsonb.into(),
                            id.into(),
                        ],
                    )?;

                    // Loop continuation: if assistant returned
                    // tool_calls AND we haven't exhausted agent.steps,
                    // enqueue a tool_dispatch row. The bgworker will
                    // pick it up on the next poll (~500ms).
                    let has_tool_calls = assistant_tool_calls
                        .as_ref()
                        .and_then(|v| v.as_array())
                        .map(|a| !a.is_empty())
                        .unwrap_or(false);
                    let mut continuation_enqueued: Option<i64> = None;
                    let mut stop_reason: Option<&'static str> = None;
                    if has_tool_calls && finish_reason.as_deref() == Some("tool_calls") {
                        // Pull iteration count and agent.steps in one
                        // round-trip. Default steps to 8 if the agent
                        // row's steps column is somehow NULL.
                        let iter_row = client.select(
                            "SELECT \
                                stewards.iteration_count($1) AS iter, \
                                coalesce((stewards.resolve_agent($2, $3)).steps, 8) AS max_steps",
                            Some(1),
                            &[
                                session_id.clone().into(),
                                agent_family.clone().into(),
                                requested_model.clone().into(),
                            ],
                        )?;
                        let mut iter_iter = iter_row.into_iter();
                        let iter_r = iter_iter.next().expect("iter row");
                        let iter_count: i32 = iter_r.get(1)?.unwrap_or(0);
                        let max_steps: i32 = iter_r.get(2)?.unwrap_or(8);

                        if iter_count < max_steps {
                            let enq_row = client.select(
                                "SELECT stewards.tool_dispatch_enqueue($1, $2, $3, $4, $5)",
                                Some(1),
                                &[
                                    id.into(),
                                    agent_family.clone().into(),
                                    requested_model.clone().into(),
                                    session_id.clone().into(),
                                    provider.to_string().into(),
                                ],
                            )?;
                            let mut e_iter = enq_row.into_iter();
                            let e_r = e_iter.next().expect("enqueue returns id");
                            continuation_enqueued = Some(e_r.get(1)?.unwrap_or(0));
                        } else {
                            pgrx::log!(
                                "stewards: agent step budget exhausted ({} >= {}); not continuing",
                                iter_count, max_steps
                            );
                            stop_reason = Some("steps_exhausted");
                        }
                    } else if has_tool_calls {
                        // Provider returned tool_calls but with a
                        // finish_reason other than 'tool_calls'
                        // (e.g., 'length' truncation mid-call). Don't
                        // try to continue — the call list may be
                        // incomplete and dispatching it would corrupt
                        // the conversation.
                        stop_reason = Some("truncated_tool_calls");
                    }

                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "chat",
                        "provider": provider,
                        "model": model,
                        "session_id": session_id,
                        "finish_reason": finish_reason,
                        "tokens_in": tokens_in,
                        "tokens_out": tokens_out,
                        "reasoning_tokens": reasoning_tokens,
                        "billable_output":
                            tokens_out.unwrap_or(0)
                            + reasoning_tokens.unwrap_or(0),
                        "tool_call_count":
                            assistant_tool_calls.as_ref()
                                .and_then(|v| v.as_array())
                                .map(|a| a.len())
                                .unwrap_or(0),
                        "continuation_enqueued": continuation_enqueued,
                        "loop_stop_reason": stop_reason,
                        "response": response,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::ToolsDispatched {
                    parent_work_id,
                    session_id,
                    agent_family,
                    model,
                    tool_messages,
                }) => {
                    // Insert one role='tool' message per dispatched
                    // call, with tool_call_id echoing the assistant's
                    // tool_call.id (provider requirement: each tool
                    // reply must reference its call). parent_work_id
                    // points at THIS tool_dispatch row for trace.
                    for (tc_id, _name, content) in tool_messages.iter() {
                        client.update(
                            "INSERT INTO stewards.messages \
                                (session_id, role, content, \
                                 tool_call_id, parent_work_id) \
                             VALUES ($1, 'tool', $2, $3, $4)",
                            None,
                            &[
                                session_id.clone().into(),
                                content.clone().into(),
                                tc_id.clone().into(),
                                id.into(),
                            ],
                        )?;
                    }

                    // Enqueue the next chat round. compose_messages
                    // will pick up the new tool messages automatically
                    // because they're now in the session history.
                    let next_row = client.select(
                        "SELECT stewards.chat_post_internal($1, $2, $3, $4)",
                        Some(1),
                        &[
                            agent_family.clone().into(),
                            model.clone().into(),
                            session_id.clone().into(),
                            provider.to_string().into(),
                        ],
                    )?;
                    let mut n_iter = next_row.into_iter();
                    let next_chat_work_id: i64 = n_iter
                        .next()
                        .and_then(|r| r.get(1).ok().flatten())
                        .unwrap_or(0);

                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "tool_dispatch",
                        "parent_work_id": parent_work_id,
                        "session_id": session_id,
                        "tool_count": tool_messages.len(),
                        "tools": tool_messages.iter()
                            .map(|(tc_id, name, _)| serde_json::json!({
                                "tool_call_id": tc_id,
                                "name": name,
                            }))
                            .collect::<Vec<_>>(),
                        "next_chat_work_id": next_chat_work_id,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Err(msg) => {
                    pgrx::log!("stewards: work_item id={} failed: {}", id, msg);
                    // Best-effort: also stamp the brain row's
                    // embedding_error if this was an embed job, so
                    // the failure surfaces in app queries.
                    if kind == "embed" {
                        if let (Some(table), Some(target_id)) = (
                            payload.get("target_table").and_then(|v| v.as_str()),
                            payload.get("target_id").and_then(|v| v.as_str()),
                        ) {
                            let stamp = format!(
                                "UPDATE stewards.{} SET embedding_error = $2 WHERE id = $1",
                                table
                            );
                            // Ignore secondary errors (e.g., table
                            // we don't know about) — primary error
                            // is already on its way to the queue.
                            let _ = client.update(
                                &stamp,
                                None,
                                &[target_id.to_string().into(), msg.clone().into()],
                            );
                        }
                    }
                    // tool_dispatch failures: write synthetic
                    // role='tool' replies + enqueue continuation so
                    // the loop never stalls. Phase 1.6 left this
                    // gap open. Phase 1.6.1 closes it.
                    let mut continuation: Option<i64> = None;
                    if kind == "tool_dispatch" {
                        if let (Some(parent), Some(session), Some(family), Some(model_str)) = (
                            payload.get("parent_work_id").and_then(|v| v.as_i64()),
                            payload.get("session_id").and_then(|v| v.as_str()),
                            payload.get("agent_family").and_then(|v| v.as_str()),
                            payload.get("model").and_then(|v| v.as_str()),
                        ) {
                            let synth = client.select(
                                "SELECT stewards.synthesize_tool_failure($1, $2, $3, $4, $5, $6)",
                                Some(1),
                                &[
                                    parent.into(),
                                    family.to_string().into(),
                                    model_str.to_string().into(),
                                    session.to_string().into(),
                                    provider.to_string().into(),
                                    msg.clone().into(),
                                ],
                            );
                            match synth {
                                Ok(rows) => {
                                    continuation = rows.into_iter().next()
                                        .and_then(|r| r.get(1).ok().flatten());
                                    pgrx::log!(
                                        "stewards: synthesized tool failure for parent={}; continuation={:?}",
                                        parent, continuation
                                    );
                                }
                                Err(e) => {
                                    pgrx::log!(
                                        "stewards: synthesize_tool_failure SPI failed: {} (loop will stall)",
                                        e
                                    );
                                }
                            }
                        }
                    }
                    let err_result = pgrx::JsonB(serde_json::json!({
                        "error": msg,
                        "continuation_after_failure": continuation,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'error', error = $2, result = $3, \
                             done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), msg.clone().into(), err_result.into()],
                    )?;
                }
            }

            // NOTIFY listeners with the row id as payload.
            let notify_sql = format!("NOTIFY stewards_done, '{}'", id);
            client.update(&notify_sql, None, &[])?;
            Ok(())
        })
    });

    if let Err(e) = write {
        pgrx::log!("stewards: write phase errored for id={}: {}", id, e);
    }
    true
}

/// Result of running a single work item, before it's written back.
enum WorkOutcome {
    Echo(serde_json::Value),
    Embedded {
        target_table: String,
        target_id: String,
        model: String,
        embedding_text: String,
        dimensions: i32,
    },
    Chatted {
        // Raw provider response (full JSON), for the work_queue audit trail.
        response: serde_json::Value,
        // Echo back so phase 3 can persist the assistant message.
        session_id: String,
        // Model the provider actually used (echo from response.model).
        model: String,
        // Continuation context: needed if assistant returned tool_calls
        // and we want to enqueue a tool_dispatch for the next loop step.
        // These mirror what was in the original payload.
        agent_family: String,
        requested_model: String,
        // Extracted bits we want to write into stewards.messages.
        assistant_content: String,
        assistant_tool_calls: Option<serde_json::Value>,
        // Captured reasoning fields. Required to echo back on the
        // next request when thinking is enabled (kimi-k2.6, o1).
        reasoning_content: Option<String>,
        reasoning_details: Option<serde_json::Value>,
        finish_reason: Option<String>,
        tokens_in: Option<i32>,
        tokens_out: Option<i32>,
        // OpenAI usage.completion_tokens_details.reasoning_tokens.
        // Billed separately from tokens_out by kimi/o1-style models;
        // store so cost computation can sum both. None when absent.
        reasoning_tokens: Option<i32>,
    },
    /// Result of executing one or more tool calls. Phase 3 inserts
    /// each (tool_call_id, content) as a `role='tool'` message and
    /// then enqueues the next chat round to continue the loop.
    ToolsDispatched {
        parent_work_id: i64,
        session_id: String,
        agent_family: String,
        model: String,
        // Per call: (tool_call_id, tool_name, content_jsonb_string).
        // content is what the model will see in the next turn. It's
        // either the JSON-stringified tool result or {"error": "..."}.
        tool_messages: Vec<(String, String, String)>,
    },
}

/// Dispatch a work item by `kind`. Returns `Ok(WorkOutcome)` on
/// success, `Err(message)` on failure (the message is stored in
/// `work_queue.error` and surfaces to callers).
fn dispatch(
    kind: &str,
    provider: &str,
    payload: &serde_json::Value,
) -> Result<WorkOutcome, String> {
    match kind {
        "echo" => Ok(WorkOutcome::Echo(serde_json::json!({
            "echo": payload,
            "kind": kind,
            "provider": provider,
            "stub": "pg_ai_stewards echo",
        }))),
        "embed" => embed(provider, payload),
        "chat"  => chat(provider, payload),
        "tool_dispatch" => tool_dispatch(payload),
        other => Err(format!("unknown work kind: {}", other)),
    }
}

/// Call an OpenAI-compatible /v1/embeddings endpoint and format the
/// response as a Postgres `vector` text literal (e.g. "[0.1,0.2,...]").
fn embed(provider_name: &str, payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let provider = PROVIDER_REGISTRY
        .get()
        .ok_or_else(|| "provider registry not initialized".to_string())?
        .providers
        .iter()
        .find(|p| p.name == provider_name)
        .ok_or_else(|| format!("unknown provider: {}", provider_name))?;

    let text = payload
        .get("text")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.text missing".to_string())?;
    let model = payload
        .get("model")
        .and_then(|v| v.as_str())
        .unwrap_or(&provider.default_model);
    let target_table = payload
        .get("target_table")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.target_table missing".to_string())?
        .to_string();
    let target_id = payload
        .get("target_id")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.target_id missing".to_string())?
        .to_string();
    let expected_dim = payload
        .get("dimensions")
        .and_then(|v| v.as_i64())
        .unwrap_or(768) as i32;

    let url = format!(
        "{}/embeddings",
        provider.base_url.trim_end_matches('/')
    );
    let body = serde_json::json!({
        "model": model,
        "input": text,
    });

    // 120s timeout: LM Studio's first request after a cold start
    // can take that long while it loads the model into memory.
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(120))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = client.post(&url).json(&body);
    if let Some(key) = &provider.api_key {
        req = req.bearer_auth(key);
    }

    let resp = req
        .send()
        .map_err(|e| format!("POST {}: {}", url, e))?;
    let status = resp.status();
    if !status.is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("embeddings HTTP {}: {}", status, body));
    }

    let parsed: serde_json::Value = resp
        .json()
        .map_err(|e| format!("decode embeddings response: {}", e))?;

    let arr = parsed
        .get("data")
        .and_then(|d| d.as_array())
        .and_then(|a| a.first())
        .and_then(|d| d.get("embedding"))
        .and_then(|e| e.as_array())
        .ok_or_else(|| {
            format!(
                "unexpected embeddings response shape: {}",
                parsed
            )
        })?;

    if arr.len() as i32 != expected_dim {
        return Err(format!(
            "embedding dimension mismatch: got {}, expected {}",
            arr.len(),
            expected_dim
        ));
    }

    // Build pgvector's text format: "[v1,v2,...]". No spaces; floats
    // formatted with full f32 precision.
    let mut s = String::with_capacity(arr.len() * 12);
    s.push('[');
    for (i, v) in arr.iter().enumerate() {
        if i > 0 {
            s.push(',');
        }
        let f = v
            .as_f64()
            .ok_or_else(|| format!("embedding[{}] not a number", i))?;
        // f32 max precision is ~9 digits; pgvector stores f32 anyway.
        s.push_str(&format!("{}", f));
    }
    s.push(']');

    Ok(WorkOutcome::Embedded {
        target_table,
        target_id,
        model: model.to_string(),
        embedding_text: s,
        dimensions: expected_dim,
    })
}

/// Call an OpenAI-compatible /v1/chat/completions endpoint.
///
/// Payload shape (built by stewards.chat_enqueue):
///   {
///     "session_id":      "<id>",
///     "agent_family":    "<family>",
///     "requested_model": "<model>",
///     "meta":            { ... audit only, not sent ... },
///     "body":            { "model":..., "messages":[...], "tools":[...], ... }
///   }
///
/// On success, returns Chatted with the parsed assistant message
/// extracted into top-level fields. Phase 3 inserts that message
/// into stewards.messages and stamps usage.
fn chat(provider_name: &str, payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let provider = PROVIDER_REGISTRY
        .get()
        .ok_or_else(|| "provider registry not initialized".to_string())?
        .providers
        .iter()
        .find(|p| p.name == provider_name)
        .ok_or_else(|| format!("unknown provider: {}", provider_name))?;

    let session_id = payload
        .get("session_id")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.session_id missing".to_string())?
        .to_string();
    let agent_family = payload
        .get("agent_family")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.agent_family missing".to_string())?
        .to_string();
    let requested_model = payload
        .get("requested_model")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.requested_model missing".to_string())?
        .to_string();
    let body = payload
        .get("body")
        .ok_or_else(|| "payload.body missing".to_string())?;

    let url = format!(
        "{}/chat/completions",
        provider.base_url.trim_end_matches('/')
    );

    // Same 120s timeout as embeddings — first kimi-k2.6 turn over
    // OpenCode Go can be slow if the gateway is cold.
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(120))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = client.post(&url).json(body);
    if let Some(key) = &provider.api_key {
        req = req.bearer_auth(key);
    }

    let resp = req
        .send()
        .map_err(|e| format!("POST {}: {}", url, e))?;
    let status = resp.status();
    if !status.is_success() {
        let resp_body = resp.text().unwrap_or_default();
        return Err(format!("chat HTTP {}: {}", status, resp_body));
    }

    let parsed: serde_json::Value = resp
        .json()
        .map_err(|e| format!("decode chat response: {}", e))?;

    // Standard OpenAI shape: { choices: [{ message: { role, content,
    // tool_calls? }, finish_reason }], usage: { prompt_tokens,
    // completion_tokens } }
    let choice = parsed
        .get("choices")
        .and_then(|c| c.as_array())
        .and_then(|a| a.first())
        .ok_or_else(|| format!("no choices[0] in response: {}", parsed))?;
    let message = choice
        .get("message")
        .ok_or_else(|| format!("no choices[0].message: {}", parsed))?;

    // OpenAI returns content as either a string OR null (when only
    // tool_calls are present). NOT NULL on messages.content with
    // default '' handles both — we coerce to "".
    let assistant_content = message
        .get("content")
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .to_string();
    let assistant_tool_calls = message.get("tool_calls").cloned();
    // Reasoning capture. Field names vary by gateway:
    //   OpenRouter / OpenCode Go: `reasoning` (string), `reasoning_details` (array)
    //   Moonshot direct:          `reasoning_content` (string)
    // Coalesce both string forms; keep details verbatim for fidelity.
    let reasoning_content = message
        .get("reasoning_content")
        .or_else(|| message.get("reasoning"))
        .and_then(|v| v.as_str())
        .map(String::from);
    let reasoning_details = message.get("reasoning_details").cloned();
    let finish_reason = choice
        .get("finish_reason")
        .and_then(|v| v.as_str())
        .map(String::from);

    let model = parsed
        .get("model")
        .and_then(|v| v.as_str())
        .unwrap_or_else(|| {
            body.get("model").and_then(|v| v.as_str()).unwrap_or("?")
        })
        .to_string();

    let usage = parsed.get("usage");
    let tokens_in = usage
        .and_then(|u| u.get("prompt_tokens"))
        .and_then(|v| v.as_i64())
        .map(|v| v as i32);
    let tokens_out = usage
        .and_then(|u| u.get("completion_tokens"))
        .and_then(|v| v.as_i64())
        .map(|v| v as i32);
    // OpenAI's newer usage shape:
    //   usage.completion_tokens_details.reasoning_tokens
    // Reasoning tokens are NOT a subset of completion_tokens for kimi/
    // o1-class models — they're billed separately. The OpenCode Go
    // dashboard's "OUTPUT" column sums both; we record them apart so
    // cost math stays honest.
    let reasoning_tokens = usage
        .and_then(|u| u.get("completion_tokens_details"))
        .and_then(|d| d.get("reasoning_tokens"))
        .and_then(|v| v.as_i64())
        .map(|v| v as i32);

    Ok(WorkOutcome::Chatted {
        response: parsed,
        session_id,
        model,
        agent_family,
        requested_model,
        assistant_content,
        assistant_tool_calls,
        reasoning_content,
        reasoning_details,
        finish_reason,
        tokens_in,
        tokens_out,
        reasoning_tokens,
    })
}

// ---------------------------------------------------------------------------
// Phase 1.6: tool_dispatch — execute the tool_calls from a parent
// assistant message and produce N role='tool' replies for phase 3.
//
// Two execute_target kinds are wired up:
//   sql_fn: SELECT <schema>.<name>($1::jsonb)::text
//   http:   POST args as JSON body, response.text() returned as-is
// Future kinds (subagent, mcp) are deferred.
//
// Tool errors are NOT returned as Err(). Each per-call failure is
// captured into the tool reply content as {"error": "..."}, so the
// model sees what went wrong and can recover. Only structural
// failures (no parent message, malformed payload) raise Err.
// ---------------------------------------------------------------------------

fn tool_dispatch(payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let parent_work_id = payload
        .get("parent_work_id")
        .and_then(|v| v.as_i64())
        .ok_or_else(|| "payload.parent_work_id missing".to_string())?;
    let session_id = payload
        .get("session_id")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.session_id missing".to_string())?
        .to_string();
    let agent_family = payload
        .get("agent_family")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.agent_family missing".to_string())?
        .to_string();
    let model = payload
        .get("model")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.model missing".to_string())?
        .to_string();

    // Tx A: read the parent assistant message's tool_calls and grab
    // the (already fetched) tool_def execute_target for each name.
    // We do this in one tx so the dispatch loop below sees a coherent
    // snapshot. The dispatch itself runs OUTSIDE this tx so HTTP
    // tools don't hold row locks.
    type Prep = Vec<(String, String, serde_json::Value, serde_json::Value)>;
    let prep: Result<Option<Prep>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect(|client| {
                // Find the assistant message produced by the parent
                // chat work item. We look it up by parent_work_id
                // (set in phase 3 of chat).
                let rows = client.select(
                    "SELECT tool_calls FROM stewards.messages \
                     WHERE parent_work_id = $1 AND role = 'assistant' \
                     ORDER BY id DESC LIMIT 1",
                    Some(1),
                    &[parent_work_id.into()],
                )?;
                let mut iter = rows.into_iter();
                let Some(row) = iter.next() else {
                    return Ok(None);
                };
                let tool_calls: pgrx::JsonB = row.get(1)?.expect("tool_calls non-null");
                let tcs = tool_calls.0.as_array().cloned().unwrap_or_default();

                // For each tool_call, look up the tool_def. Build
                // (tool_call_id, name, args_jsonb, target_jsonb).
                let mut prepped: Prep = Vec::with_capacity(tcs.len());
                for tc in tcs {
                    let tc_id: String = tc.get("id")
                        .and_then(|v| v.as_str())
                        .unwrap_or("unknown")
                        .to_string();
                    let name: String = tc.get("function")
                        .and_then(|f| f.get("name"))
                        .and_then(|v| v.as_str())
                        .unwrap_or("")
                        .to_string();
                    // OpenAI returns function.arguments as a STRING
                    // (JSON-encoded). Decode here so dispatch sees a
                    // jsonb. If the model emits malformed JSON, fall
                    // back to a sentinel so dispatch can still run
                    // and the tool can complain meaningfully.
                    let args_str = tc.get("function")
                        .and_then(|f| f.get("arguments"))
                        .and_then(|v| v.as_str())
                        .unwrap_or("{}");
                    let args = serde_json::from_str::<serde_json::Value>(args_str)
                        .unwrap_or_else(|_| serde_json::json!({
                            "_decode_error": "function.arguments was not valid JSON",
                            "_raw": args_str,
                        }));

                    // Look up tool_def. If absent, store a sentinel
                    // target so dispatch can return a structured
                    // error reply (the model needs to know).
                    let target_rows = client.select(
                        "SELECT execute_target FROM stewards.tool_defs \
                         WHERE name = $1 AND active",
                        Some(1),
                        &[name.clone().into()],
                    )?;
                    let target = match target_rows.into_iter().next() {
                        Some(r) => r.get::<pgrx::JsonB>(1)?.map(|j| j.0)
                            .unwrap_or(serde_json::json!({"kind":"missing"})),
                        None => serde_json::json!({"kind":"missing"}),
                    };
                    prepped.push((tc_id, name, args, target));
                }
                Ok(Some(prepped))
            })
        });

    let prepped = prep
        .map_err(|e| format!("tool_dispatch prep tx: {}", e))?
        .ok_or_else(|| format!(
            "no assistant message found for parent_work_id={}", parent_work_id
        ))?;

    // Phase 2 (no tx): execute each tool. Collect (tc_id, name, content).
    let mut tool_messages: Vec<(String, String, String)> =
        Vec::with_capacity(prepped.len());
    for (tc_id, name, args, target) in prepped.into_iter() {
        let content = match exec_one_tool(&name, &args, &target) {
            Ok(s) => s,
            Err(e) => {
                pgrx::log!("stewards: tool '{}' failed: {}", name, e);
                serde_json::json!({"error": e}).to_string()
            }
        };
        tool_messages.push((tc_id, name, content));
    }

    Ok(WorkOutcome::ToolsDispatched {
        parent_work_id,
        session_id,
        agent_family,
        model,
        tool_messages,
    })
}

/// Dispatch a single tool call. Returns the content string the
/// model will see as the tool reply. SHOULD be a JSON-parseable
/// string (the convention is that tools return JSON), but the
/// LLM is told this is a tool reply so freeform strings work too.
fn exec_one_tool(
    name: &str,
    args: &serde_json::Value,
    target: &serde_json::Value,
) -> Result<String, String> {
    let kind = target.get("kind")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "tool execute_target.kind missing".to_string())?;
    match kind {
        "sql_fn" => exec_sql_fn_tool(target, args),
        "http"   => exec_http_tool(target, args),
        "missing" => Err(format!("tool '{}' is not registered or inactive", name)),
        other    => Err(format!("unsupported tool kind: {}", other)),
    }
}

/// SQL function tool. Convention: the target SQL fn has signature
///   fn(p_args jsonb) RETURNS jsonb
/// Wrapped by stewards.<name>_tool functions for this convention.
fn exec_sql_fn_tool(
    target: &serde_json::Value,
    args: &serde_json::Value,
) -> Result<String, String> {
    let schema = target.get("schema")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "sql_fn target.schema missing".to_string())?;
    let fn_name = target.get("name")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "sql_fn target.name missing".to_string())?;

    // Identifier-safe-ish guard: schema and name must match a
    // simple identifier pattern. This is a defense-in-depth measure
    // because we're string-formatting these into SQL. The CHECK
    // constraint on tool_defs.name already enforces this at insert,
    // but the schema field is free-form.
    let safe = |s: &str| s.chars().all(|c| c.is_ascii_alphanumeric() || c == '_');
    if !safe(schema) || !safe(fn_name) {
        return Err(format!("unsafe identifier in sql_fn target: {}.{}", schema, fn_name));
    }

    let sql = format!("SELECT {}.{}($1)::text", schema, fn_name);
    // JsonB doesn't impl Clone in pgrx 0.18; build the value once and
    // wrap it in Rc so multiple PgTryBuilder retries (if any) could
    // share it. We currently only call it once, so a fresh build per
    // entry is fine — just don't try to .clone() it later.
    let args_value = args.clone();

    // Pre-flight: does the function exist with a jsonb signature?
    // PG ereports on missing function would otherwise reach the
    // bgworker via longjmp; PgTryBuilder is supposed to catch them
    // but in pgrx 0.18 + BackgroundWorker::transaction the longjmp
    // path empirically still kills the worker (verified in
    // testing — see verify-loop.sql). The cheapest defense is to
    // check pg_proc first and never trigger the ereport.
    let exists: Result<bool, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect(|client| {
                let row = client.select(
                    "SELECT EXISTS ( \
                        SELECT 1 FROM pg_proc p \
                        JOIN pg_namespace n ON p.pronamespace = n.oid \
                        WHERE n.nspname = $1 AND p.proname = $2 \
                          AND pg_get_function_arguments(p.oid) ILIKE '%jsonb%' \
                     )",
                    Some(1),
                    &[schema.into(), fn_name.into()],
                )?;
                Ok(row.into_iter().next()
                    .and_then(|r| r.get::<bool>(1).ok().flatten())
                    .unwrap_or(false))
            })
        });
    match exists {
        Ok(true) => { /* fall through */ }
        Ok(false) => return Err(format!(
            "sql_fn target {}.{}(jsonb) does not exist", schema, fn_name)),
        Err(e) => return Err(format!("sql_fn pre-flight: {}", e)),
    }

    use pgrx::PgTryBuilder;

    // PgTryBuilder wraps PG_TRY/PG_CATCH. It catches the ereport,
    // unwinds the implicit subtx pgrx opened around BackgroundWorker
    // ::transaction, and returns a recoverable Err we match on here.
    // The bgworker survives.
    //
    // We do NOT use SAVEPOINT here. SAVEPOINT requires an explicit
    // BEGIN, but BackgroundWorker::transaction opens an implicit one
    // \u2014 trying to issue SAVEPOINT inside it errors with
    // "SAVEPOINT can only be used in transaction blocks", which
    // ironically broke even the success path. PG_TRY handles the
    // unwind without our help.
    let result: Result<Option<String>, String> = PgTryBuilder::new(|| {
        let outer: Result<Option<String>, pgrx::spi::Error> =
            BackgroundWorker::transaction(|| {
                Spi::connect(|client| {
                    let row = client.select(
                        &sql, Some(1),
                        &[pgrx::JsonB(args_value.clone()).into()]
                    )?;
                    let mut iter = row.into_iter();
                    match iter.next() {
                        Some(r) => r.get::<String>(1),
                        None => Ok(None),
                    }
                })
            });
        outer.map_err(|e| format!("spi: {}", e))
    })
    .catch_others(|cause| {
        Err(format!("postgres error: {:?}", cause))
    })
    .execute();

    match result {
        Ok(Some(s)) => Ok(s),
        Ok(None) => Ok("null".to_string()),
        Err(e) => Err(format!("sql_fn {}.{}: {}", schema, fn_name, e)),
    }
}

/// HTTP tool. Target shape:
///   {"kind":"http", "method":"POST", "url":"...", "headers":{...}}
/// Method defaults to POST. Args are sent as JSON body. Response
/// body is returned as-is (assumed JSON; freeform strings also OK).
fn exec_http_tool(
    target: &serde_json::Value,
    args: &serde_json::Value,
) -> Result<String, String> {
    let url = target.get("url")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "http target.url missing".to_string())?;
    let method = target.get("method")
        .and_then(|v| v.as_str())
        .unwrap_or("POST")
        .to_uppercase();

    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(60))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = match method.as_str() {
        "POST" => client.post(url).json(args),
        "GET"  => client.get(url),
        other  => return Err(format!("unsupported http method: {}", other)),
    };

    if let Some(headers) = target.get("headers").and_then(|v| v.as_object()) {
        for (k, v) in headers {
            if let Some(vs) = v.as_str() {
                req = req.header(k.as_str(), vs);
            }
        }
    }

    let resp = req.send().map_err(|e| format!("POST {}: {}", url, e))?;
    let status = resp.status();
    let body = resp.text().unwrap_or_default();
    if !status.is_success() {
        return Err(format!("http {} {}: {}", method, status, body));
    }
    Ok(body)
}

// ---------------------------------------------------------------------------
// Tests (run with `cargo pgrx test`)
// ---------------------------------------------------------------------------

#[cfg(any(test, feature = "pg_test"))]
#[pg_schema]
mod tests {
    use pgrx::prelude::*;

    #[pg_test]
    fn version_returns_pkg_version() {
        let got = Spi::get_one::<&str>("SELECT stewards.version()")
            .expect("SPI succeeded")
            .expect("non-null result");
        assert_eq!(got, "0.1.0");
    }
}

#[cfg(test)]
pub mod pg_test {
    pub fn setup(_options: Vec<&str>) {}

    pub fn postgresql_conf_options() -> Vec<&'static str> {
        // For `cargo pgrx test` the bgworker needs to be preloaded.
        vec!["shared_preload_libraries='pg_ai_stewards'"]
    }
}
