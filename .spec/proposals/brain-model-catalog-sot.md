# Brain — Model Catalog Single Source of Truth

**Date:** 2026-04-23  
**Status:** Approved + executing  
**Stewardship:** Michael explicitly granted — *"I'm giving you stewardship over this. Make it usable."*

## Binding problem

The Commission dialog in brain offers `Claude Opus 4.6 (3.0×)` and `Claude Sonnet 4`. Both are wrong:

- Copilot migrated to **Opus 4.7** at **7.5×** premium-request cost weeks ago.
- Sonnet is **4.6** (not 4).
- Haiku 4.5 isn't listed at all even though it's the research-stage default.
- GPT-5 and GPT-5-mini aren't listed even though Copilot offers them.

The root cause isn't that the dialog is stale — it's that there are **three** parallel model tables in the codebase that drift independently:

1. `internal/config/config.go` → `AvailableModels` (Discord preset map, rate as string `"0.33x"`)
2. `internal/config/config.go` → `modelCosts` (map\[string\]float64)
3. `internal/steward/steward.go` → `DefaultConfig().EscalationChain` (ModelTier slice, cost as float64)
4. Hardcoded `<option>` elements in `CommissionDialog.vue` and `ProjectDetailView.vue` (label+cost as literal text)

Already-visible drift: `PipelineBigModel`'s comment says "3.0x" while `modelCosts` says 7.5.

## Success criteria

- [ ] One Go struct/file is the source of truth for every model the system knows about: id, display name, cost multiplier, family, whether it's in the auto-escalation chain, stage defaults.
- [ ] The three existing tables are consolidated or derived from this one.
- [ ] A `GET /api/models` endpoint returns the catalog so the UI doesn't have to hardcode it.
- [ ] `CommissionDialog.vue` and `ProjectDetailView.vue` inline commission dialog both render options from the catalog, not from hardcoded strings.
- [ ] Default selection is the correct plan-stage model (`claude-opus-4.7`), not `claude-opus-4.6`.
- [ ] GPT-5 and GPT-5-mini appear in the dropdown (user requested, even though they're not in the steward escalation chain).
- [ ] Screenshot after: dropdown shows `Claude Opus 4.7 (7.5×)`, `Claude Sonnet 4.6 (1.0×)`, `Claude Haiku 4.5 (0.33×)`, `GPT-5 (1.0×)`, `GPT-5 Mini (0×)`.

## Constraints

- Keep the existing steward escalation chain semantics: haiku → sonnet → opus. GPT models are selectable but not part of the auto-escalation chain.
- Don't break existing tests (there are ~20 test assertions hardcoded on model strings; those strings are correct, leave them).
- Don't break the Discord preset keys (`raptor`, `gpt-mini`, `haiku`, `sonnet`, `flash`, `gpt5`). Those are the user's mnemonic aliases.

## Implementation

**Single file:** `scripts/brain/internal/config/models.go` (new).

```go
package config

type Model struct {
    ID          string  `json:"id"`           // canonical id, e.g. "claude-opus-4.7"
    DisplayName string  `json:"display_name"` // "Claude Opus 4.7"
    Family      string  `json:"family"`       // "claude" | "gpt" | "gemini"
    Cost        float64 `json:"cost"`         // premium-request multiplier
    PresetKey   string  `json:"preset_key,omitempty"` // Discord alias
    InEscalation bool   `json:"in_escalation"` // part of auto-escalation chain
    EscalationRank int  `json:"escalation_rank,omitempty"` // 0=cheap, 1=mid, 2=top
}

var Catalog = []Model{
    {ID: "claude-haiku-4.5",  DisplayName: "Claude Haiku 4.5",  Family: "claude", Cost: 0.33, PresetKey: "haiku",    InEscalation: true, EscalationRank: 0},
    {ID: "claude-sonnet-4.6", DisplayName: "Claude Sonnet 4.6", Family: "claude", Cost: 1.0,  PresetKey: "sonnet",   InEscalation: true, EscalationRank: 1},
    {ID: "claude-opus-4.7",   DisplayName: "Claude Opus 4.7",   Family: "claude", Cost: 7.5,                          InEscalation: true, EscalationRank: 2},
    {ID: "gpt-5",             DisplayName: "GPT-5",             Family: "gpt",    Cost: 1.0,  PresetKey: "gpt5"},
    {ID: "gpt-5-mini",        DisplayName: "GPT-5 Mini",        Family: "gpt",    Cost: 0.0,  PresetKey: "gpt-mini"},
    {ID: "gemini-3-flash",    DisplayName: "Gemini 3 Flash",    Family: "gemini", Cost: 0.33, PresetKey: "flash"},
    {ID: "raptor-mini",       DisplayName: "Raptor Mini",       Family: "other",  Cost: 0.0,  PresetKey: "raptor"},
}

// StageDefaults are the default model per pipeline stage.
var StageDefaults = map[string]string{
    "research": "claude-haiku-4.5",
    "plan":     "claude-opus-4.7",
    "execute":  "claude-sonnet-4.6",
    "commission": "claude-opus-4.7", // user picks this from the dialog
}
```

**Refactor:**
- `modelCosts` → derive from `Catalog` (loop once at init).
- `AvailableModels` → derive from `Catalog` where `PresetKey != ""`.
- `PipelineCheapModel`/`SmartModel`/`BigModel` constants → keep them as thin references (still compile-safe) but point to catalog IDs.
- `steward.DefaultConfig().EscalationChain` → build from `Catalog` where `InEscalation` is true, sorted by `EscalationRank`.
- Fix the "3.0x" comment drift.

**API endpoint:** `GET /api/models` returns `{ models: Catalog, stage_defaults: StageDefaults }`.

**Frontend:**
- `src/composables/useModelCatalog.ts` — fetch once on app mount, cache in a module-level ref.
- `CommissionDialog.vue` — replace hardcoded `<option>`s with `v-for` over catalog; default `model.value` from `stage_defaults.commission`.
- `ProjectDetailView.vue` inline commission dropdown at ~line 1670 — same treatment.
- `ProjectDetailView.vue` hardcoded default at lines 59 and 313 → set from catalog.

## Verification

1. `go build -tags fts5` clean.
2. `go test ./...` for the brain — tests assert on specific strings; they should still pass because those strings are already correct.
3. Start brain, open Projects → an entry → Commission dialog. Confirm dropdown shows all six models with correct cost labels, defaulted to Opus 4.7.
4. Same check for the inline "+ New Entry" commission dialog in ProjectDetailView.
5. `curl http://localhost:8445/api/models | jq` returns the catalog JSON.

## Out of scope

- Routing to GPT or Gemini models (the Copilot SDK router may or may not handle them — if a commission with `gpt-5` is submitted and the router rejects it, the steward's existing model_limit escalation path takes over). If this becomes a real issue, handle in a follow-up.
- Moving `Discord` routing to use the catalog directly (the `AvailableModels` map is consumed by Discord handlers in a specific shape; keep the derived form).
