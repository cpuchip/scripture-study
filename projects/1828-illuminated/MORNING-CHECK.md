# Morning check — 1828-illuminated

Quick walk-through for when you wake up. The project shipped 10 commits across three autonomous-loop iterations between ~midnight and ~3am. Each step below is verifiable in under a minute.

## Step 1 — Build + run (one command pair)

```sh
cd projects/1828-illuminated
docker build -t 1828-illuminated . && docker run -d --name 1828 -p 8080:80 1828-illuminated
# Then open http://localhost:8080
```

Image is ~30MB. Container starts in <2s. Last verified at 03:59:42 UTC by `docker build -q` + `Invoke-WebRequest /healthz`.

## Step 2 — Click through the eight surfaces

| Route | What to check | Expected |
|-------|--------------|----------|
| `/` | Home — three feature cards + tier-A chips | "intelligence", "obtain", "tempest", etc. as clickable chips |
| `/word` | Search "intelligence" → click | Card shows 1828 (matter-spectrum substance) + modern (cognitive faculty) + study link to `truth.md` |
| `/word/obtain` | Direct deep link | Should land on the obtain card with Webster 1828 + modern + priesthood-oath study reference |
| `/verse` | Pick D&C 130:18-19 from dropdown | "intelligence" highlighted; click pops the card with substrate study link |
| `/verse` (paste mode) | Paste any verse with KJV-style verbs | "suffereth" → highlights via stem to "suffer"; "endureth" → "endure"; "obtaining" → "obtain" |
| `/present?v=dc-130-18-19` | Tap a word | Fullscreen card with 1828 + modern; ← → arrow keys to navigate verses; Esc to close |
| `/dictionary` | Toggle reading level on "intelligence" | Elementary / 8th grade / college+ render at different depths; college+ shows key passages + GC reinforcement |
| `/about` | Scroll | Methodology + four honest cautions + provenance pointer to `research/gospel/1828/` |
| `/settings` | Click "LM Studio" preset | URL fills in `http://localhost:1234/v1`; status shows ✓ Configured |

## Step 3 — Test "Render in modern English" (optional, requires LM Studio running)

1. Open `/settings`, click "LM Studio" preset (or paste your opencode-go URL + key)
2. Open `/verse`, pick a verse, click **"Render in modern English"**
3. Wait 5-15s for the response
4. Verify the rendered output uses 1828 senses with `[bracket]` markers on substituted phrases

If LM Studio isn't running on `:1234`, the button shows an error inline — it does NOT crash the page. The site never sees your API key.

## Step 4 — Quick verify the substrate Thummim machinery (no LLM calls)

```sh
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -c \
  "SELECT family FROM stewards.pipelines WHERE family='thummim-define'; \
   SELECT count(*) FROM stewards.thummim_entries;"
```

Expected output:
```
     family
----------------
 thummim-define

 count
-------
     0
```

Zero entries means nothing's been generated yet. The pipeline is queued and ready.

## Step 5 — (Optional) dispatch your first Thummim entry

If you want to see the substrate generate one entry as a smoke test (~$0.30):

```sh
docker exec pg-ai-stewards-dev psql -U stewards -d stewards <<'SQL'
DO $$
DECLARE v_wi uuid; v_q bigint;
BEGIN
  v_wi := stewards.work_item_create(
    'thummim-define',
    jsonb_build_object('word', 'obtain',
                       'binding_question', 'How does the Restoration corpus use the word ''obtain'', and how does that compare to Webster 1828?'),
    'thummim-obtain', 'human', NULL,
    (SELECT id FROM stewards.intents WHERE slug='scripture-study'));
  v_q := stewards.work_item_dispatch_stage(v_wi);
  RAISE NOTICE 'Dispatched: % (queue id %)', v_wi, v_q;
END $$;
SQL
```

Watch progress with:
```sh
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -P pager=off -c \
  "SELECT slug, current_stage, status, (cost_micro_dollars::numeric/1000000)::numeric(10,4) AS usd FROM stewards.work_items WHERE pipeline_family='thummim-define';"
```

After it completes (3 stages, ~10-15 min total), the JSON output lands in `stage_results.review` and you can run `python3 scripts/export_thummim.py` to refresh the frontend bundle.

## Where to find things

| | Path |
|---|---|
| Project root | `projects/1828-illuminated/` |
| Intent + constraints | `projects/1828-illuminated/intent.yaml` |
| Per-project working protocol | `projects/1828-illuminated/CLAUDE.md` |
| Per-session journals | `projects/1828-illuminated/.spec/journal/` |
| 1828 proposal (parent vision) | `.spec/proposals/1828-illuminated-scriptures.md` |
| Thummim proposal | `.spec/proposals/thummim-restoration-dictionary.md` |
| Word-list research (P1-P5) | `research/gospel/1828/` |
| Substrate pipeline SQL | `projects/pg-ai-stewards/extension/thm1-thummim-pipeline-and-schema.sql` |

## What I deferred (and why)

- **Bundle optimization** (2.4MB unminified / 841KB gzipped) — the `useWordData` chunk eagerly loads all 858 words' definitions. Lazy-loading via dynamic `import()` would split `/home` from the data-heavy pages. Real win, real regression risk; deferred to a cleaner session with you awake.
- **Auto-promote synthesize JSON into `thummim_entries`** — currently the substrate writes the file but doesn't parse the JSON output into the structured table. A small `apply_thummim_result()` trigger would close that loop. Documented as D-THM-7 in the pipeline SQL header.
- **D-1828-1..5 and D-THM-1..6 ratifications** — neither captured the full set of forks. Sitting in the proposals waiting for your input. None blocked tonight's work, but next-session work either dispatches with current defaults or pauses for your call.
- **gospel-engine-v2 integration** — needs a system MCP key separate from the dev key. Frontend would call engine.ibeco.me for live scripture search; today it works without it.

## Commits

```
ce4c1d2 1828-illuminated: thummim export script + example-dispatch SQL
8e31203 mind: active.md — Thummim substrate pipeline shipped
153ff43 substrate(THM.1): thummim-define pipeline + thummim_entries schema
6a4663e 1828-illuminated: iteration-2 journal entry
0658e5b 1828-illuminated: stretch goal #3 — Thummim Dictionary scaffolding
0aef597 1828-illuminated: stretch goal #2 — presentation mode
ed45def 1828-illuminated: stretch goal #1 — LLM-rendering settings + verse render
2cce3dd 1828-illuminated: complete modern defs + stemming fallback + manual additions
b66444d 1828-illuminated: new project — Vue SPA + Docker for 1828.ibeco.me
2b7fd95 research(1828): overnight word-list groundwork for 1828-illuminated scriptures
```

If you like what you see and want this deployed, the Dockerfile is ready for Dokploy. Branch from main and push to a `1828-illuminated` remote, or pull the image to your registry. The static SPA expects to be served from any HTTPS origin — no env vars needed.

If something looks off, every output file in `research/gospel/1828/` is reproducible from `research/gospel/1828/.work/*.py`. Every substrate piece is in version control. The MVP can be wiped + rebuilt from source in 5 minutes.

Sleep well. The work survived the night.
