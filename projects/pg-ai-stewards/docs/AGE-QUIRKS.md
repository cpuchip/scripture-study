# AGE Quirks Catalog

Apache AGE on PostgreSQL 18. Quirks discovered while building
pg-ai-stewards Phase 2.6. Each entry has a workaround, a category
(`bug-candidate`, `spec-divergence`, `our-mistake`, `by-design`),
and where it was first encountered.

The categories matter for **Phase 6 — AGE upstream contributions**:
`bug-candidate` items are PR-worthy; `spec-divergence` items are
workarounds we live with (or document upstream as caveats);
`our-mistake` items belong here so we don't re-discover them.

| #   | Quirk                                                         | Category          | First seen |
| --- | ------------------------------------------------------------- | ----------------- | ---------- |
| 1   | No `ON CREATE SET` / `ON MATCH SET` after MERGE              | spec-divergence   | 2.6a       |
| 2   | Apostrophes break interpolated Cypher; param binding required | bug-candidate     | 2.6a       |
| 3   | `WITH` required between MERGE and a subsequent MATCH         | spec-divergence   | 2.6a       |
| 4   | Cypher labels cannot be parameter-bound                       | by-design         | 2.6b       |
| 5   | Variable-length path syntax `[r*1..N]` is awkward            | spec-divergence   | 2.6c       |
| 6   | `cypher()`'s 3rd argument MUST be a parameter placeholder     | bug-candidate     | 2.6c       |
| 7   | `#>>` jsonb-path operator fails on agtype scalars             | bug-candidate     | 2.6c       |
| 8   | PL/pgSQL OUT params shadow column names in `RETURNING`       | our-mistake       | 2.6c       |

---

## #1 — No `ON CREATE SET` / `ON MATCH SET` after MERGE

**Category:** spec-divergence. Cypher (Neo4j) supports it; AGE does not (yet).

**Symptom:**

```cypher
MERGE (n:Foo {slug: 'x'})
ON CREATE SET n.created_at = timestamp()
ON MATCH SET n.last_seen = timestamp()
-- ERROR: syntax error at or near "ON"
```

**Workaround:** unconditional `SET` after the MERGE — it runs on both
create and match. If you genuinely need different behavior, use two
queries (a MATCH-then-CREATE-if-null pattern via `WITH`).

```cypher
MERGE (n:Foo {slug: 'x'})
SET n.last_seen = timestamp()      -- runs always
SET n.created_at = coalesce(n.created_at, timestamp())  -- idiomatic substitute
```

**First seen:** 2.6a, while writing `link_declared_edges()` for workstreams.

---

## #2 — Apostrophes break interpolated Cypher; param binding required

**Category:** bug-candidate. The error message is unhelpful and the failure mode silent in some shapes.

**Symptom:** any Cypher string containing a literal apostrophe (e.g., a slug like `study/it's-easter` or a title with `'`) breaks when interpolated naively into the `cypher('graph', $$ ... $$)` body.

**Workaround:** always pass values via the `$param` mechanism + a
PostgreSQL-side `jsonb_build_object` cast to `agtype`. Never
string-concatenate user data into the Cypher body.

```sql
EXECUTE
    $cy$
    SELECT * FROM cypher('stewards_graph', $$
        MERGE (n:Study {slug: $slug})
        SET n.title = $title
    $$, $1) AS (v agtype)
    $cy$
USING (jsonb_build_object(
    'slug',  p_slug,
    'title', p_title
)::text)::ag_catalog.agtype;
```

**Upstream PR potential:** improve the parser error message OR auto-escape inside the cypher body.

**First seen:** 2.6a, when a workstream id with a hyphen got truncated and a study title with an apostrophe blew up.

---

## #3 — `WITH` required between MERGE and a subsequent MATCH

**Category:** spec-divergence. Neo4j's planner inserts an implicit
`WITH *`; AGE requires it explicit.

**Symptom:**

```cypher
MERGE (a:Foo {id: 'x'})
MATCH (b:Bar {id: 'y'})
MERGE (a)-[:REL]->(b)
-- ERROR: syntax error at or near "MATCH"
```

**Workaround:** explicit `WITH *` between the clauses.

```cypher
MERGE (a:Foo {id: 'x'})
WITH a
MATCH (b:Bar {id: 'y'})
MERGE (a)-[:REL]->(b)
```

**First seen:** 2.6a, while writing `link_declared_edges()` workstream→proposal MERGE.

---

## #4 — Cypher labels cannot be parameter-bound

**Category:** by-design (this matches the Cypher spec — labels are
schema, not data). Documenting because it's surprising for anyone
coming from a SQL `format()` mindset.

**Symptom:**

```cypher
MERGE (p:$kind {slug: $slug})
-- ERROR: variable `kind` not defined
```

**Workaround:** branch on the kind in PL/pgSQL and use `format()` with
the kind as a SQL literal. Safe ONLY when the kind value is constrained
(CHECK constraint, enum, or hard-coded list) — never use raw user input
here.

```sql
EXECUTE format(
    $cy$
    SELECT * FROM cypher('stewards_graph', $$
        MERGE (p:%s {slug: $slug}) ...
    $$, $1) AS (v agtype)
    $cy$,
    p_kind  -- safe: CHECK-constrained to {Workstream, Study, Phase, Todo}
)
USING (jsonb_build_object('slug', p_slug)::text)::ag_catalog.agtype;
```

**First seen:** 2.6b, while writing `create_todo()` which needs to
attach :HAS_TODO to a parent vertex of varying label.

---

## #5 — Variable-length path syntax `[r*1..N]` is awkward

**Category:** spec-divergence. AGE accepts the syntax but several
common operations on the resulting paths break:
- `length(path)` — sometimes returns null, sometimes errors
- direction extraction (`startNode(rel) = s`) — type confusion across hops
- `UNWIND` of relationship list — drops properties

**Symptom:** any non-trivial query over a variable-length path tends
to fail in confusing ways. Single-hop and fixed-length paths work
fine.

**Workaround:** for graph walks deeper than 1 hop, do **iterative
1-hop in PL/pgSQL** instead. Frontier table + visited set + 1-hop
helper function. See `stewards.context_for()` for the canonical
implementation.

```sql
-- Don't do this in AGE:
MATCH path = (s {slug: $slug})-[*1..3]-(n)
RETURN length(path), n

-- Do this instead (PL/pgSQL):
WHILE v_hop <= v_depth LOOP
    -- expand frontier 1 hop via stewards.context_for_hop()
    -- dedupe in SQL temp table
END LOOP;
```

**First seen:** 2.6c, building `context_for(slug, depth)`.

---

## #6 — `cypher()`'s 3rd argument MUST be a parameter placeholder

**Category:** bug-candidate. The restriction is real but the error
message is precise enough that it's at least diagnosable.

**Symptom:**

```sql
SELECT * FROM cypher('stewards_graph', $$
    MATCH (n {slug: $s}) RETURN n
$$, (jsonb_build_object('s', 'x')::text)::ag_catalog.agtype) AS (v agtype);
-- ERROR: third argument of cypher function must be a parameter
```

The 3rd argument cannot be an inline expression — even a fully-typed
`ag_catalog.agtype` literal — only a `$N` placeholder.

**Workaround:** wrap in `EXECUTE ... USING`. Bind the agtype value
through the `USING` clause so it arrives as `$1` inside the Cypher
function call.

```sql
DECLARE
    v_arg ag_catalog.agtype;
BEGIN
    v_arg := (jsonb_build_object('s', p_slug)::text)::ag_catalog.agtype;
    RETURN QUERY EXECUTE
        $sql$
        SELECT * FROM cypher('stewards_graph', $$
            MATCH (n {slug: $s}) RETURN n
        $$, $1) AS (v ag_catalog.agtype)
        $sql$
    USING v_arg;
END;
```

**Upstream PR potential:** the parser clearly knows the expected type
(`ag_catalog.agtype`); requiring a placeholder is an unnecessary
restriction. Worth a PR to AGE to relax this to "any expression of
type agtype."

**First seen:** 2.6c, building `context_for_hop()`.

---

## #7 — `#>>` jsonb-path operator fails on agtype scalars

**Category:** bug-candidate. The operator works on agtype objects
and arrays but errors on scalars (strings, numbers, nulls) returned
by Cypher `RETURN` clauses.

**Symptom:**

```sql
SELECT (etype #>> '{}')::text
  FROM cypher('g', $$ MATCH (s)-[r]->(n) RETURN type(r) $$, $1)
       AS h(etype agtype) ...
-- ERROR: right operand must be an array
```

**Workaround:** strip the JSON-ish quoting manually. The pattern
that handles all scalar types (string, number, null) cleanly:

```sql
nullif(trim(both '"' from value::text), 'null')   -> text or NULL
```

For numbers:

```sql
nullif(value::text, 'null')::float
```

**Upstream PR potential:** make `#>>` work on agtype scalars as a
pass-through (return the scalar's text form).

**First seen:** 2.6c, building `context_for_hop()`.

---

## #8 — PL/pgSQL OUT params shadow column names in `RETURNING`

**Category:** our-mistake (this is a PL/pgSQL quirk, not AGE — but it
bit us hard while building Watchman so it lives here for now).

**Symptom:** a function with `OUT neighbor text` (or any RETURNS
TABLE column) and an internal `INSERT ... RETURNING neighbor`
errors with `column reference "neighbor" is ambiguous`.

**Workaround:** qualify the column reference with the table name in
`RETURNING`.

```sql
CREATE FUNCTION foo() RETURNS TABLE (neighbor text) AS $$
BEGIN
    RETURN QUERY
    WITH ins AS (
        INSERT INTO _ctx_results(neighbor) VALUES ('x')
        RETURNING _ctx_results.neighbor AS new_neighbor   -- qualified!
    )
    SELECT i.new_neighbor FROM ins i;
END;
$$ LANGUAGE plpgsql;
```

**First seen:** 2.6c, building `context_for()`.

---

## Phase 6 candidates (PR-worthy upstream)

When pg-ai-stewards reaches a steady state, the following quirks are
worth contributing fixes for upstream:

- **#2** — improve apostrophe-in-interpolated-Cypher error message;
  consider auto-escape.
- **#6** — relax `cypher()` 3rd-arg restriction to accept any
  `ag_catalog.agtype` expression, not just placeholders.
- **#7** — extend `#>>` (and likely `->>`, `->`) to handle agtype
  scalars as pass-through.

Probably-by-design (document upstream as caveats, don't try to fix):
- #1, #3, #5 — Cypher-spec divergences. Either AGE catches up to
  Neo4j over time, or these stay workarounds forever.
- #4 — labels-as-schema is the Cypher spec.

Our problem (don't bother upstream):
- #8 — pure PL/pgSQL.

---

## Adding to this catalog

When you discover a new quirk:

1. Reproduce minimally (smallest Cypher that triggers it).
2. Confirm it's not just bad Cypher — check the
   [openCypher spec](https://opencypher.org/) and Neo4j docs.
3. Add an entry here with: symptom, workaround, category, where
   first seen.
4. If `bug-candidate`, also note "Upstream PR potential" with a
   one-line description of the desired fix.
5. Update the table at the top.
