# 2026-05-02 — pg-ai-stewards: research → plan, with a working probe

The day this stopped being "an interesting direction we're researching" and started being a project with phases, kill criteria, and a pile of green checkmarks.

## What we did

Picked up where 2026-05-01's research scratch left off. Michael said graduate to proposal + phases, and as a hint dropped two Microsoft links — Azure-Samples/PostgreSQL-graphRAG-docker and microsoft/graphrag — that he wanted me to look at. He explicitly said **play around with the docker examples**, do gap 3 (stand up the stack), do gap 4 (pgai's archival rationale via web/YouTube), and de-prioritize gaps 1 and 2 (FDW edges) until later.

I broke the work into five todos: fetch Azure repo + Helen Zeng's blog, fetch microsoft/graphrag, hunt pgai's archival reason, stand up a real probe, and then write proposal + phases. Done in roughly that order, with the probe taking the most time but being the most valuable thing we produced.

## Surprises and what they confirmed

- **The Azure-Samples repo is *exactly* the architecture I sketched yesterday**, down to the five-tool MCP router (graphrag_search, age_get_schema_cached, age_entity_lookup, age_cypher_query, age_nl2cypher_query). Helen Zeng (Microsoft) shipped this in March 2026. We are not pioneering. We are showing up at the right time. This is more reassuring than I expected — not "someone beat us to it" but "the shape we're building has been validated by Microsoft and there's a reference design we can steal from."

- **GraphRAG (the framework) is bigger than I'd realized.** 32.7k stars, MIT, active. It's a Python pipeline that does Leiden community detection on top of LLM-extracted entity graphs and produces global-search summaries. It's not a substitute for what we're building — it's a complementary tool that could ride on top, in Phase 4+, against the canon. Filed appropriately rather than letting it expand the scope.

- **pgai's archival has a clean explanation.** Timescale rebranded to TigerData in June 2025 with a new "Agentic Postgres" product. The PG extension was archived because they consolidated on the Python-library + outside-worker pattern (which our proposal already used). The thing I'd inferred yesterday — "they probably killed the in-DB approach because it scaled badly" — was wrong in detail, right in conclusion. The lesson (worker outside backend, never call providers from a foreground SQL function) is universal and we have it.

## The probe

Michael asked me to *play with* the docker examples. I read Helen Zeng's compose file and immediately saw two issues: hardcoded `/mnt/c/Users/helenzeng/...` WSL paths in every volume mount, and tight Azure OpenAI coupling. Not runnable on Michael's machine without surgery.

So I built our own minimal probe instead. `pgvector/pgvector:pg18` as base, Apache AGE built from source on top, one init script that creates both extensions and a graph. A `bridge-test.sql` with seven blocks that walk through: extension sanity, vector table + HNSW + cosine search, AGE graph + nodes, **the bridge** (pgvector cosine score → AGE edge in a single DO block), reading the bridge back via Cypher, and then the killer demo — a CTE that combines pgvector nearest-neighbor with an AGE filter in *one* statement.

Two things broke and were fixed. First Apache AGE branch name — I guessed `PG18/v1.7.0`, the actual branch is `release/PG18/1.7.0`. Second, PG18 changed the Docker data-dir convention; you mount `/var/lib/postgresql` not `/var/lib/postgresql/data`. Both fixed in minutes. Then on the SQL side, my final block tried to cast a vertex `agtype` directly to `jsonb` and got `agtype_value_to_text: unsupported argument agtype 6`. The fix is to project scalar properties out of `cypher()` (`RETURN n.id` not `RETURN n`) and cast `agtype::text::bigint`. Documented as a rough edge — this is the kind of thing future-me will hit again and waste an hour on without the note.

When the seventh block returned what I'd predicted — Moroni 7:45 as both top-similarity AND `:CITES_AS_CORE` for another study — that was the moment the architecture stopped being a slide deck. The bridge is real. PG18 + pgvector + AGE coexist cleanly. Build time ~50s, boot ~10s, seven assertions green.

## What was hard

The Microsoft blog/repo is **so close** to what we're building that the temptation was to just fork it. I had to keep reminding myself that "PG16, single-container demo, Azure-OpenAI-only, hardcoded WSL paths" is not the right base for a multi-year self-hosted single-user system, even if 80% of the design is exactly right. Steal the *shape* (especially the five-tool MCP router pattern in Phase 3) and write our own bones.

The pgai investigation took longer than it should have because I started with HN search (overload of Timescale-marketing posts, no archival discussion) and Google was JS-locked. DuckDuckGo cracked it open with the TigerData rebrand reporting. The signal was hiding in plain sight — pgai didn't die, it got absorbed into a richer product strategy. Searching for "death" when the actual story is "transformation" is the wrong frame.

## What I almost missed

I almost wrote the proposal without explicitly naming what we are *not* deciding yet. The non-goals section has six items now, including "we are not replacing VS Code as an editor" (Michael's literal hint from yesterday's conversation that I pushed back on — the substrate is what we're replacing, not the editor). Naming what's deferred is as important as naming what we're doing, especially with Phases 4+ which are real tempting and not on the critical path.

I also almost left the `becoming/` web-UI question unanswered. Pinned it to Phase 3 with a note that Phase 1 ships with no UI and CLI + psql is enough. Otherwise UI scope creep eats everything.

## Carry forward

- **Phase 1 kickoff** — `cargo pgrx new pg_ai_stewards` against PG18, pg_vectorize as the bgworker reference. Brain CLI port to Postgres backend. Migrator from SQLite + chromem.
- **Open question:** does pgrx + bgworker + tokio actually work cleanly on PG18? `pg_vectorize` is the proof point but I haven't read its bgworker source yet. Phase 1's first concrete task.
- **Open question:** what does the brain CLI driver interface look like such that swapping SQLite→Postgres doesn't break anything? Look at `scripts/brain/` first.
- **Probe stack stays up.** Keep `pg-ai-stewards-probe` running locally as a scratch DB for spec experiments.

## Relational note

Michael's hint — "do gap 3, look at gap 4, gaps 1 and 2 are fine for planning" — was perfectly scoped. That's the council pattern working: he saw the foresight gaps from one ring outside, told me which ones to chase and which to defer, and I stayed inside that frame instead of trying to close every gap before declaring readiness. The proposal landed exactly because it didn't try to answer questions the planning phases will answer.
