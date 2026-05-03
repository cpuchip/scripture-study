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
    CREATE FUNCTION stewards.touch_brain_entry() RETURNS trigger
    LANGUAGE plpgsql AS $func$
    BEGIN
        IF TG_OP = 'UPDATE' THEN
            -- Snapshot the OLD row before changes land.
            INSERT INTO stewards.brain_versions
                (entry_id, title, category, body, props, changed_by)
            VALUES
                (OLD.id, OLD.title, OLD.category, OLD.body, OLD.props,
                 coalesce(current_setting('stewards.actor', true), 'system'));
            NEW.updated_at := now();
        END IF;
        RETURN NEW;
    END;
    $func$;

    CREATE TRIGGER brain_entries_touch
        BEFORE UPDATE ON stewards.brain_entries
        FOR EACH ROW EXECUTE FUNCTION stewards.touch_brain_entry();

    -- Enqueue an embedding job whenever title/body changes (or row
    -- is inserted). The bgworker will pick it up; in step 3 the
    -- echo stub still runs, so embedding stays NULL until step 6
    -- swaps the stub for a real Ollama HTTP call.
    --
    -- Provider name 'ollama' resolves to the registry entry loaded
    -- from STEWARDS_PROVIDER_OLLAMA_*. Match gospel-engine-v2's
    -- model: nomic-embed-text v1.5, 768 dimensions.
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
                'ollama',
                jsonb_build_object(
                    'target_table', 'brain_entries',
                    'target_id',    NEW.id,
                    'text',         coalesce(NEW.title, '') || E'\n\n' || coalesce(NEW.body, ''),
                    'model',        'nomic-embed-text:v1.5',
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
fn process_one_pending() -> bool {
    let outcome: Result<bool, pgrx::spi::Error> = BackgroundWorker::transaction(|| {
        Spi::connect_mut(|client| {
            // Claim oldest pending row. SKIP LOCKED makes this safe
            // if multiple worker processes ever exist.
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
                return Ok(false);
            };

            let id: i64 = row.get(1)?.expect("id non-null");
            let kind: String = row.get(2)?.expect("kind non-null");
            let provider: String = row.get(3)?.expect("provider non-null");
            let payload: pgrx::JsonB = row.get(4)?.expect("payload non-null");

            pgrx::log!(
                "stewards: claimed work_item id={} kind={} provider={}",
                id,
                kind,
                provider
            );

            // ---- Stub "echo" provider. -------------------------
            // Real provider dispatch (Ollama / LM Studio /
            // OpenCode Go) lives in step 6/7. For now every
            // provider value resolves to the echo stub so we
            // can prove the round-trip works end-to-end.
            let result_value = serde_json::json!({
                "echo": payload.0,
                "kind": kind,
                "provider": provider,
                "stub": "pg_ai_stewards step-2 echo",
            });
            let result_jsonb = pgrx::JsonB(result_value);
            // ----------------------------------------------------

            client.update(
                "UPDATE stewards.work_queue \
                 SET status = 'done', result = $2, done_at = now() \
                 WHERE id = $1",
                None,
                &[id.into(), result_jsonb.into()],
            )?;

            // NOTIFY listeners with the row id as payload.
            // Anyone running `LISTEN stewards_done` from a normal
            // client connection will get this on commit.
            let notify_sql = format!("NOTIFY stewards_done, '{}'", id);
            client.update(&notify_sql, None, &[])?;

            Ok(true)
        })
    });

    match outcome {
        Ok(processed) => processed,
        Err(e) => {
            pgrx::log!("stewards: bgworker tick errored: {}", e);
            false
        }
    }
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
