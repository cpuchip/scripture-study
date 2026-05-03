//! pg_ai_stewards — Postgres extension scaffold (Phase 1, step 1).
//!
//! Goal of this file: prove the pgrx toolchain compiles, the extension
//! shared object loads into PG18, and one SQL function runs end-to-end.
//! Nothing else lives here yet. The bgworker, the schema, and the
//! provider dispatch arrive in subsequent commits per `phases.md`.

use pgrx::prelude::*;

::pgrx::pg_module_magic!();

/// Returns the extension build version. First sanity check that the
/// shared object loaded and pgrx wired the schema correctly.
#[pg_extern]
fn version() -> &'static str {
    env!("CARGO_PKG_VERSION")
}

/// Returns the pgrx version we were built against. Useful when
/// debugging "did my install pick up the rebuilt .so?" questions.
#[pg_extern]
fn pgrx_version() -> &'static str {
    "0.18.0"
}

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
        vec![]
    }
}
