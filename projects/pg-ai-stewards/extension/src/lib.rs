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
        content         text NOT NULL,
        model           text,
        tokens_in       int,
        tokens_out      int,
        cost_usd        numeric(10, 6),

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
        steps        int,                          -- max agentic iterations
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
            jsonb_build_object('role', m.role, 'content', m.content)
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
            $j${"kind":"sql_fn","schema":"stewards","name":"brain_search_text"}$j$::jsonb
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
            $j${"kind":"builtin","name":"load_skill"}$j$::jsonb
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
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'error', error = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), msg.clone().into()],
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
