//! Phase 5d (Phase C.2) — YAML → JSONB helpers for intent + covenant seeding.
//!
//! Two #[pg_extern] functions exposed to SQL:
//!   stewards.parse_yaml_intent(yaml text)   RETURNS jsonb
//!   stewards.parse_yaml_covenant(yaml text) RETURNS jsonb
//!   stewards.yaml_sha256(yaml text)         RETURNS text
//!
//! The parsers normalize the on-disk YAML shape into the substrate's
//! row-shaped jsonb. SQL seed functions (5d-intents-covenants-seed.sql)
//! consume these to do the actual upserts.

use pgrx::prelude::*;
use sha2::{Digest, Sha256};

/// Compute sha256 hex of a string. Used to gate re-seed when YAML hasn't changed.
#[pg_extern]
pub fn yaml_sha256(yaml: &str) -> String {
    let mut h = Sha256::new();
    h.update(yaml.as_bytes());
    hex::encode(h.finalize())
}

/// Parse intent.yaml text into a normalized jsonb shape suitable for
/// upserting into stewards.intents. The on-disk shape today is:
///
///   purpose: <text>
///   values:
///     <key>: { description: <text>, source: <text> }
///     ...
///   constraints:
///     <key>: { description, severity, source, enforcement }
///   success_criteria: [<text>, ...]
///
/// Returned jsonb shape:
///
///   {
///     "slug": "scripture-study",
///     "purpose": <text>,
///     "values_hierarchy": [{key, description, source}, ...],  // preserves order
///     "non_goals": [],                                         // none in current YAML
///     "scripture_anchor": null,
///     "beneficiary": null
///   }
///
/// The slug defaults to "scripture-study" since the root intent.yaml is
/// the project-level intent. Future: support multi-intent YAML files
/// keyed by intent_slug.
#[pg_extern]
pub fn parse_yaml_intent(yaml: &str) -> pgrx::JsonB {
    let parsed: serde_yaml::Value = match serde_yaml::from_str(yaml) {
        Ok(v) => v,
        Err(e) => return pgrx::JsonB(serde_json::json!({"error": format!("yaml parse: {}", e)})),
    };

    let purpose = parsed
        .get("purpose")
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .trim()
        .to_string();

    // values map → values_hierarchy ordered array. serde_yaml preserves
    // mapping order via its Mapping type when iterated directly.
    let mut values_hierarchy: Vec<serde_json::Value> = Vec::new();
    if let Some(values) = parsed.get("values").and_then(|v| v.as_mapping()) {
        for (k, v) in values.iter() {
            let key = k.as_str().unwrap_or("").to_string();
            let description = v
                .get("description")
                .and_then(|d| d.as_str())
                .unwrap_or("")
                .trim()
                .to_string();
            let source = v
                .get("source")
                .and_then(|s| s.as_str())
                .unwrap_or("")
                .to_string();
            values_hierarchy.push(serde_json::json!({
                "key": key,
                "description": description,
                "source": source,
            }));
        }
    }

    // constraints map — also stored in values_hierarchy as additional
    // entries (kind="constraint") since they carry the same per-trade-off
    // semantics.
    if let Some(constraints) = parsed.get("constraints").and_then(|c| c.as_mapping()) {
        for (k, v) in constraints.iter() {
            let key = k.as_str().unwrap_or("").to_string();
            let description = v
                .get("description")
                .and_then(|d| d.as_str())
                .unwrap_or("")
                .trim()
                .to_string();
            let severity = v.get("severity").and_then(|s| s.as_str()).unwrap_or("");
            let source = v.get("source").and_then(|s| s.as_str()).unwrap_or("");
            values_hierarchy.push(serde_json::json!({
                "key": key,
                "description": description,
                "severity": severity,
                "source": source,
                "kind": "constraint",
            }));
        }
    }

    // success_criteria → not part of values_hierarchy; surface as a
    // sibling field. Non-goals comes from explicit non_goals when added
    // to YAML; for now empty.
    let success_criteria: Vec<String> = parsed
        .get("success_criteria")
        .and_then(|v| v.as_sequence())
        .map(|seq| {
            seq.iter()
                .filter_map(|s| s.as_str().map(|s| s.to_string()))
                .collect()
        })
        .unwrap_or_default();

    pgrx::JsonB(serde_json::json!({
        "slug": "scripture-study",
        "purpose": purpose,
        "values_hierarchy": values_hierarchy,
        "non_goals": Vec::<String>::new(),
        "success_criteria": success_criteria,
        "beneficiary": serde_json::Value::Null,
        "scripture_anchor": serde_json::Value::Null,
    }))
}

/// Parse .spec/covenant.yaml into normalized jsonb for stewards.covenants.
///
/// On-disk shape (abbreviated):
///   purpose: <text>
///   human_commits_to:
///     <key>: { description, why }
///   agent_commits_to:
///     <key>: { description, why }
///   when_broken: { description, recovery }
///   council_moment: { description, applies_to, why }
///   teaching: { purpose, human_commits_to, agent_commits_to }   # optional Section 7
///
/// Returned jsonb:
///   {
///     "scope": "global",
///     "human_commits_to": [{key, description, why}, ...],
///     "agent_commits_to": [{key, description, why}, ...],
///     "when_broken": <text>,
///     "recovery": <text>,
///     "council_moment": <text>,
///     "teaching_extension": {...} | null,
///     "ratified_by": "both"
///   }
#[pg_extern]
pub fn parse_yaml_covenant(yaml: &str) -> pgrx::JsonB {
    let parsed: serde_yaml::Value = match serde_yaml::from_str(yaml) {
        Ok(v) => v,
        Err(e) => return pgrx::JsonB(serde_json::json!({"error": format!("yaml parse: {}", e)})),
    };

    let human = commits_to_array(&parsed, "human_commits_to");
    let agent = commits_to_array(&parsed, "agent_commits_to");

    let when_broken = parsed
        .get("when_broken")
        .and_then(|v| v.get("description"))
        .and_then(|d| d.as_str())
        .unwrap_or("")
        .trim()
        .to_string();

    let recovery = parsed
        .get("when_broken")
        .and_then(|v| v.get("recovery"))
        .and_then(|d| d.as_str())
        .unwrap_or("")
        .trim()
        .to_string();

    let council_moment = parsed
        .get("council_moment")
        .and_then(|v| v.get("description"))
        .and_then(|d| d.as_str())
        .unwrap_or("")
        .trim()
        .to_string();

    // Optional teaching extension — only present in covenants that
    // include the Section 7 covenant. Pass through as nested jsonb.
    let teaching_extension = parsed.get("teaching").map(|t| {
        serde_json::json!({
            "purpose": t.get("purpose").and_then(|v| v.as_str()).unwrap_or("").trim(),
            "human_commits_to": commits_to_array(t, "human_commits_to"),
            "agent_commits_to": commits_to_array(t, "agent_commits_to"),
        })
    });

    pgrx::JsonB(serde_json::json!({
        "scope": "global",
        "human_commits_to": human,
        "agent_commits_to": agent,
        "when_broken": when_broken,
        "recovery": recovery,
        "council_moment": council_moment,
        "teaching_extension": teaching_extension,
        "ratified_by": "both",
    }))
}

fn commits_to_array(root: &serde_yaml::Value, key: &str) -> Vec<serde_json::Value> {
    let mut out = Vec::new();
    if let Some(map) = root.get(key).and_then(|v| v.as_mapping()) {
        for (k, v) in map.iter() {
            let key_str = k.as_str().unwrap_or("").to_string();
            let description = v
                .get("description")
                .and_then(|d| d.as_str())
                .unwrap_or("")
                .trim()
                .to_string();
            let why = v
                .get("why")
                .and_then(|d| d.as_str())
                .unwrap_or("")
                .trim()
                .to_string();
            out.push(serde_json::json!({
                "key": key_str,
                "description": description,
                "why": why,
            }));
        }
    }
    out
}
