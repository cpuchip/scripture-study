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

use pgrx::prelude::*;

mod bgworker;
mod providers;
mod schema;
mod tools;
mod types;
mod yaml;
use providers::{Provider, ProviderRegistry, ProviderSummary, PROVIDER_REGISTRY};

::pgrx::pg_module_magic!();


// =====================================================================
// Phase 2.6 / 2.7 / 3a — folded back from live-DB migration files.
//
// These five blocks were originally applied incrementally to the dev
// database via psql redirects (extension/2-6a-*.sql etc) so we could
// iterate on the substrate without rebuilding the extension binary
// every time. As of v0.2.0 they are part of the canonical install
// chain, in linear dependency order:
//
//   create_study_show
//      └─ create_workstreams           (2-6a: workstream vertices + DECLARED edges)
//          └─ create_todos             (2-6b: todos as persistent connector vertices)
//              └─ create_phases_context (2-6c: phase splitter + context_for() walk)
//                  └─ create_watchman_substrate (2-7a: dirty-bit + verdicts + findings + dirty_queue)
//                      └─ create_watchman_pass  (3a: watchman-consolidator agent + watchman_input())
//
// The .sql files remain in the repo as the canonical source of each
// block's text (extension_sql_file! reads them at compile time via
// include_str! semantics). Editing the SQL files is the right move;
// editing the macro signatures here is only for renames/dependency
// changes.
//
// Idempotency: every block uses CREATE OR REPLACE, ADD COLUMN IF NOT
// EXISTS, ON CONFLICT DO UPDATE, etc. so applying the same block twice
// is a no-op. This matters for `cargo pgrx schema` which may run blocks
// multiple times during development.
// =====================================================================

extension_sql_file!(
    "../2-6a-workstreams.sql",
    name = "create_workstreams",
    requires = ["create_study_show"],
);

extension_sql_file!(
    "../2-6b-todos.sql",
    name = "create_todos",
    requires = ["create_workstreams"],
);

extension_sql_file!(
    "../2-6c-phases-context.sql",
    name = "create_phases_context",
    requires = ["create_todos"],
);

extension_sql_file!(
    "../2-7a-watchman-substrate.sql",
    name = "create_watchman_substrate",
    requires = ["create_phases_context"],
);

extension_sql_file!(
    "../3a-watchman-pass.sql",
    name = "create_watchman_pass",
    requires = ["create_watchman_substrate"],
);

extension_sql_file!(
    "../2-7b1-watchman-automation.sql",
    name = "create_watchman_automation",
    requires = ["create_watchman_pass"],
);

extension_sql_file!(
    "../2-7b2-watchman-scheduler.sql",
    name = "create_watchman_scheduler",
    requires = ["create_watchman_automation"],
);

extension_sql_file!(
    "../2-7b3-watchman-budget.sql",
    name = "create_watchman_budget",
    requires = ["create_watchman_scheduler"],
);

extension_sql_file!(
    "../2-7b4-watchman-soak-prep.sql",
    name = "create_watchman_soak_prep",
    requires = ["create_watchman_budget"],
);

extension_sql_file!(
    "../3c1-pipelines-work-items.sql",
    name = "create_pipelines_work_items",
    requires = ["create_watchman_soak_prep"],
);

extension_sql_file!(
    "../3c2-work-item-advance-trigger.sql",
    name = "create_work_item_advance_trigger",
    requires = ["create_pipelines_work_items"],
);

extension_sql_file!(
    "../3c2-5-study-tools.sql",
    name = "create_study_tools",
    requires = ["create_work_item_advance_trigger"],
);

extension_sql_file!(
    "../3c3-stage-templating-and-study-write.sql",
    name = "create_stage_templating_and_study_write",
    requires = ["create_study_tools"],
);

extension_sql_file!(
    "../3c3-1-trigger-bugfixes.sql",
    name = "create_trigger_bugfixes_3c3_1",
    requires = ["create_stage_templating_and_study_write"],
);

extension_sql_file!(
    "../3c3-3-agent-tool-perms-provenance.sql",
    name = "create_agent_tool_perms_provenance",
    requires = ["create_trigger_bugfixes_3c3_1"],
);

extension_sql_file!(
    "../3c3-5-work-items-to-studies.sql",
    name = "create_work_items_to_studies_promotion",
    requires = ["create_agent_tool_perms_provenance"],
);

extension_sql_file!(
    "../3e2-1-mcp-bridge-schemas.sql",
    name = "create_mcp_bridge_schemas",
    requires = ["create_work_items_to_studies_promotion"],
);

extension_sql_file!(
    "../3e2-2-mcp-proxy-dispatch.sql",
    name = "create_mcp_proxy_dispatch",
    requires = ["create_mcp_bridge_schemas"],
);

extension_sql_file!(
    "../3e2-3-mcp-tool-cache-promote.sql",
    name = "create_mcp_tool_cache_promote",
    requires = ["create_mcp_proxy_dispatch"],
);

extension_sql_file!(
    "../3e2-4-fetch-md-seed.sql",
    name = "create_fetch_md_seed",
    requires = ["create_mcp_tool_cache_promote"],
);

extension_sql_file!(
    "../3e2-5-broaden-mcp-grants.sql",
    name = "create_broaden_mcp_grants",
    requires = ["create_fetch_md_seed"],
);

extension_sql_file!(
    "../3e2-6-mcp-servers-linux-paths.sql",
    name = "create_mcp_servers_linux_paths",
    requires = ["create_broaden_mcp_grants"],
);

extension_sql_file!(
    "../3e2-7-git-mcp-seed.sql",
    name = "create_git_mcp_seed",
    requires = ["create_mcp_servers_linux_paths"],
);

// ---------------------------------------------------------------------------
// Phase 4a — Substrate-Phase-A schema (D-A4 cost tracking + D-B1 escalation
// chain + D-EC3 human-mediated escalation queue).
// Spec: projects/pg-ai-stewards/.spec/proposals/{cost-tracking,escalation-chain}.md
// ---------------------------------------------------------------------------

extension_sql_file!(
    "../4a-cost-tracking.sql",
    name = "create_phase_4a_cost_tracking",
    requires = ["create_git_mcp_seed"],
);

extension_sql_file!(
    "../4a-escalation-chain.sql",
    name = "create_phase_4a_escalation_chain",
    requires = ["create_phase_4a_cost_tracking"],
);

extension_sql_file!(
    "../4a-steward.sql",
    name = "create_phase_4a_steward",
    requires = ["create_phase_4a_escalation_chain"],
);

extension_sql_file!(
    "../4b-dispatch-override.sql",
    name = "create_phase_4b_dispatch_override",
    requires = [
        "create_phase_4a_steward",
        "create_stage_templating_and_study_write"
    ],
);

extension_sql_file!(
    "../4c-steward-dispatch.sql",
    name = "create_phase_4c_steward_dispatch",
    requires = ["create_phase_4b_dispatch_override"],
);

extension_sql_file!(
    "../4d-steward-realign.sql",
    name = "create_phase_4d_steward_realign",
    requires = ["create_phase_4c_steward_dispatch"],
);

extension_sql_file!(
    "../4g-ad-hoc-cost.sql",
    name = "create_phase_4g_ad_hoc_cost",
    requires = ["create_phase_4d_steward_realign"],
);

extension_sql_file!(
    "../5a-maturity-gate.sql",
    name = "create_phase_5a_maturity_gate",
    requires = [
        "create_phase_4g_ad_hoc_cost",
        "create_stage_templating_and_study_write"
    ],
);

extension_sql_file!(
    "../5b-scenarios-verify.sql",
    name = "create_phase_5b_scenarios_verify",
    requires = ["create_phase_5a_maturity_gate"],
);

extension_sql_file!(
    "../5c-sessions-gate-kind.sql",
    name = "create_phase_5c_sessions_gate_kind",
    requires = ["create_phase_5b_scenarios_verify"],
);

extension_sql_file!(
    "../5d-intents-covenants.sql",
    name = "create_phase_5d_intents_covenants",
    requires = ["create_phase_5c_sessions_gate_kind"],
);

extension_sql_file!(
    "../5d2-seed-fns.sql",
    name = "create_phase_5d2_seed_fns",
    requires = ["create_phase_5d_intents_covenants"],
);

extension_sql_file!(
    "../5d3-compose-with-intent.sql",
    name = "create_phase_5d3_compose_with_intent",
    requires = ["create_phase_5d2_seed_fns"],
);

extension_sql_file!(
    "../5d4-backfill-intent.sql",
    name = "create_phase_5d4_backfill_intent",
    requires = ["create_phase_5d3_compose_with_intent"],
);

extension_sql_file!(
    "../5d5-tools-off-and-templates.sql",
    name = "create_phase_5d5_tools_off_and_templates",
    requires = ["create_phase_5d4_backfill_intent"],
);

extension_sql_file!(
    "../5e-lessons-and-pipeline-flags.sql",
    name = "create_phase_5e_lessons_and_pipeline_flags",
    requires = ["create_phase_5d5_tools_off_and_templates"],
);

extension_sql_file!(
    "../5e2-sabbath.sql",
    name = "create_phase_5e2_sabbath",
    requires = ["create_phase_5e_lessons_and_pipeline_flags"],
);

extension_sql_file!(
    "../5e3-atonement.sql",
    name = "create_phase_5e3_atonement",
    requires = ["create_phase_5e2_sabbath"],
);

extension_sql_file!(
    "../5e4-promotion-gate-and-triggers.sql",
    name = "create_phase_5e4_promotion_gate_and_triggers",
    requires = ["create_phase_5e3_atonement"],
);

extension_sql_file!(
    "../5f-trust.sql",
    name = "create_phase_5f_trust",
    requires = ["create_phase_5e4_promotion_gate_and_triggers"],
);

extension_sql_file!(
    "../5f2-evaluate-trust.sql",
    name = "create_phase_5f2_evaluate_trust",
    requires = ["create_phase_5f_trust"],
);

extension_sql_file!(
    "../5f3-gate-trust-check.sql",
    name = "create_phase_5f3_gate_trust_check",
    requires = ["create_phase_5f2_evaluate_trust"],
);

extension_sql_file!(
    "../5f4-retry-with-lessons.sql",
    name = "create_phase_5f4_retry_with_lessons",
    requires = ["create_phase_5f3_gate_trust_check"],
);

extension_sql_file!(
    "../5f5-apply-gate-override.sql",
    name = "create_phase_5f5_apply_gate_override",
    requires = ["create_phase_5f4_retry_with_lessons"],
);

extension_sql_file!(
    "../5g-council.sql",
    name = "create_phase_5g_council",
    requires = ["create_phase_5f5_apply_gate_override"],
);

extension_sql_file!(
    "../5g2-convene-council.sql",
    name = "create_phase_5g2_convene_council",
    requires = ["create_phase_5g_council"],
);

extension_sql_file!(
    "../5g3-synthesize-and-resolve.sql",
    name = "create_phase_5g3_synthesize_and_resolve",
    requires = ["create_phase_5g2_convene_council"],
);

extension_sql_file!(
    "../5g4-bishop-and-suggest.sql",
    name = "create_phase_5g4_bishop_and_suggest",
    requires = ["create_phase_5g3_synthesize_and_resolve"],
);

// Batch G — make the substrate land in real files.
extension_sql_file!(
    "../6a-studies-file-path-nullable.sql",
    name = "create_batch_g1_studies_file_path_nullable",
    requires = ["create_phase_5g4_bishop_and_suggest"],
);

extension_sql_file!(
    "../6b-steward-retry-lessons.sql",
    name = "create_batch_g2_steward_retry_lessons",
    requires = ["create_batch_g1_studies_file_path_nullable"],
);

extension_sql_file!(
    "../6c-quarantine-fires-atonement.sql",
    name = "create_batch_g3_quarantine_fires_atonement",
    requires = ["create_batch_g2_steward_retry_lessons"],
);

extension_sql_file!(
    "../6d-pending-file-writes.sql",
    name = "create_batch_g4_1_pending_file_writes",
    requires = ["create_batch_g3_quarantine_fires_atonement"],
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
// Module-split breadcrumbs (Phase 3c.3.6, 2026-05-08):
//   - Provider registry types + GospelEngineConfig → providers.rs
//   - WorkOutcome enum → types.rs
//   - _PG_init + bgworker tick loop + dispatch/embed/chat → bgworker.rs
//   - resolve_ref + tool_dispatch + exec_* helpers → tools.rs
// ---------------------------------------------------------------------------


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
