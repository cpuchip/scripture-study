//! Provider registry — env-var bootstrap, in-process only.
//!
//! Reads `STEWARDS_PROVIDER_<NAME>_<FIELD>` env vars at postmaster
//! startup and exposes the parsed registry to the bgworker via the
//! `PROVIDER_REGISTRY` `OnceLock`. Also holds the gospel-engine
//! resolver config in its own `OnceLock`.
//!
//! Extracted from lib.rs (Phase 3c.3.6 module split, 2026-05-08).
//! Items kept `pub(crate)` so the bgworker + dispatch helpers can
//! read them, but not exposed beyond the crate.

use std::sync::OnceLock;

/// Snapshot of one provider's metadata, minus the secret. Returned
/// from `stewards.providers_loaded()`.
#[derive(Clone, Debug)]
pub(crate) struct ProviderSummary {
    pub(crate) name: String,
    pub(crate) base_url: String,
    pub(crate) default_model: String,
    pub(crate) kind: String,
    pub(crate) has_api_key: bool,
}

#[derive(Clone, Debug)]
pub(crate) struct Provider {
    pub(crate) name: String,
    pub(crate) base_url: String,
    pub(crate) default_model: String,
    pub(crate) kind: String,
    pub(crate) api_key: Option<String>,
}

#[derive(Default, Debug)]
pub(crate) struct ProviderRegistry {
    pub(crate) providers: Vec<Provider>,
}

impl ProviderRegistry {
    /// Parse `STEWARDS_PROVIDER_<NAME>_<FIELD>` env vars into a
    /// registry. Lossy by design: malformed entries are skipped with
    /// a warning rather than aborting the worker.
    pub(crate) fn from_env() -> Self {
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

    pub(crate) fn summary(&self) -> Vec<ProviderSummary> {
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
pub(crate) static PROVIDER_REGISTRY: OnceLock<ProviderRegistry> = OnceLock::new();

/// Phase 2.2 — gospel-engine resolver config. Read once from env at
/// postmaster startup. URL has no trailing slash; token is bearer.
/// Both Optional so the resolver can fail gracefully if env is unset
/// (returns "GOSPEL_ENGINE_URL not set" which is stored in
/// resolved_refs.error and visible to callers).
#[derive(Debug, Clone, Default)]
pub(crate) struct GospelEngineConfig {
    pub(crate) url: Option<String>,
    pub(crate) token: Option<String>,
}
pub(crate) static GOSPEL_ENGINE_CONFIG: OnceLock<GospelEngineConfig> = OnceLock::new();
