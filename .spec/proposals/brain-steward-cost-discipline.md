# Brain — Steward Cost Discipline

**Date:** 2026-04-23
**Status:** Approved + executing
**Stewardship:** Michael — *"This is a hard fix, we cannot use brain pipeline until we fix this, it's too expensive."*
**Budget context:** ~40 opus requests left for the next 7 days. Whatever we do here, do it cheap.

## Binding problem

A single commission with a 100-credit budget burned 105 credits looping verify→revise unbounded, all on opus-4.7. Diagnosis breakdown is in the prior turn / journal. Three concrete bugs:

1. **Commission `model` field overrides every stage.** It was meant as the steward's judgment model. It's being passed verbatim to research, spec, execute, verify, revise — every call. The `StageDefaults` map (haiku/sonnet/opus per stage) is never consulted during a commission run.
2. **Verify→revise loop is unbounded.** Spins until budget exhausts.
3. **Verify is pinned to whatever the commission picked.** Verify is a yes/no judgment — should be cheap.

## Success criteria

- [ ] When commission picks Opus, only **steward decisions** (gate evaluation, judgment) use Opus. Pipeline stages (research/spec/execute/revise) use their `StageDefaults` model.
- [ ] Verify is hard-pinned to `claude-haiku-4.5` regardless of commission setting.
- [ ] Revise loop caps at **2 per commission**. On the 3rd verifier rejection, the steward surfaces with `loop_limit_exceeded` and pulls the human in.
- [ ] User-facing visibility: the entry's commission detail (and the entry detail header) shows revise-loop count badge ("Revised 2/2") so user sees the wall coming.
- [ ] All existing tests pass (most assert on specific model strings — those need to be updated to match the new stage-defaults behavior, NOT the test logic).
- [ ] Build: `go build -tags fts5` clean, `npm run build` clean, all `go test -tags fts5 ./...` pass.

## Constraints

- **Do not redesign the pipeline.** This is patching three discrete defects in the existing flow.
- **Per-stage user overrides are out of scope** for this proposal. The architecture must allow adding them later (so route every "what model does stage X use?" through a single helper), but no UI for it.
- **Keep the commission `model` field semantically meaningful** — it now means "steward judgment model" (used by EvaluateGate). UI label updates accordingly.
- **Cost discipline applies to this work too.** Don't burn opus on the implementation. Frame as a focused refactor.

## Implementation

### 1. Stage model resolution helper

In `internal/steward/commission.go`, add:

```go
// modelForStage returns the model to use for a given pipeline stage.
// Verify is hard-pinned to haiku. Other stages use config.StageDefaults.
// The commission's Model field is reserved for steward judgment (gate evaluation).
func (s *Steward) modelForStage(c *store.Commission, stage string) string {
    if stage == "verify" {
        return "claude-haiku-4.5"
    }
    if m, ok := config.StageDefaults[stage]; ok {
        return m
    }
    return c.Model // fallback for unknown stages
}
```

Replace every `c.Model` argument passed into `runner.RetryAdvance`, `runner.RetryExecute`, `runner.GenerateScenarios`, and the verify path with `s.modelForStage(c, stage)`. Same for the `recordDecision` model field.

**Keep `c.Model` as the argument to `runner.EvaluateGate(...)`** — that IS the steward's judgment call and should use the commission's selected model.

Cost tracking: `s.modelCost(s.modelForStage(c, stage))` instead of `s.modelCost(c.Model)` everywhere except for the eval cost.

### 2. Add `verify` and `research` to StageDefaults

In `internal/config/models.go`, expand `StageDefaults`:

```go
var StageDefaults = map[string]string{
    "research":   "claude-haiku-4.5",
    "plan":       "claude-opus-4.7",
    "spec":       "claude-sonnet-4.6",
    "execute":    "claude-sonnet-4.6",
    "verify":     "claude-haiku-4.5",
    "revise":     "claude-sonnet-4.6",
    "commission": "claude-opus-4.7", // default for the steward judgment model
}
```

(Verify is also hard-pinned in `modelForStage` as a belt-and-suspenders guard. If someone "fixes" StageDefaults later, the pin survives.)

### 3. Revise loop cap

Add `RevisionCount int` to `store.Commission` (DB migration: ALTER TABLE commissions ADD COLUMN revision_count INTEGER NOT NULL DEFAULT 0).

In the verify→revise branch of `commissionAdvanceStage`, increment `c.RevisionCount` and persist. If `c.RevisionCount >= 2` (i.e. about to start the third revise), instead of running the revision:

```go
s.commissionSurface(c, "loop_limit_exceeded",
    fmt.Sprintf("Verifier rejected the work %d times. Surfacing for human review.\n\nLast feedback: %s",
        c.RevisionCount, feedback))
return false, nil
```

Record a `decision` with action=`surface`, reasoning includes the loop count and last feedback.

### 4. UI visibility

**Frontend (`CommissionPanel.vue` or equivalent — find where commission status renders on the entry detail page):**
- Show "Revised X/2" badge when `revision_count > 0`. Yellow at 1, red at 2.
- Show the active stage and last decision in real time (it may already; verify and improve).

**Backend:** ensure `revision_count` is included in `Commission` JSON response.

If the entry detail doesn't currently show *which commission is running and at what stage*, add a single line: "Steward running: stage={current_stage}, model={c.Model judgment / actual stage model}, revisions={n}/2, cost={used}/{max}." Don't redesign the UI — one informational line is enough.

### 5. Test updates

Tests in `commission_test.go` and `steward_test.go` assert on specific model strings:
- Lines that assert `"claude-opus-4.7"` for execute/spec stages need to be updated to `"claude-sonnet-4.6"`.
- Lines that assert verify uses the commission model need to be updated to `"claude-haiku-4.5"`.
- Add **new** test: revise loop hits cap → surface action, no third revision attempted.
- Add **new** test: commission model is opus, but execute/spec/verify stages use their stage defaults.

Do NOT weaken any test by removing assertions. Update the expected values to match the new (correct) behavior.

## Verification

1. `go build -tags fts5 ./cmd/brain/` clean.
2. `go test -tags fts5 ./internal/steward/... ./internal/config/...` all pass.
3. `npm run build` clean.
4. **Trace test (mental, not actual run — don't burn opus on a live commission):** Trace through what would happen for a new commission with `model=claude-opus-4.7`, intent="build a thing":
   - Research stage → haiku call → ~0.33 credits
   - Gate eval → opus call → 7.5 credits
   - (advance) Plan stage → opus call → 7.5 credits
   - Gate eval → opus call → 7.5 credits
   - (advance) Spec stage → sonnet call → 1.0 credits
   - Execute stage → sonnet call → 1.0 credits
   - Verify stage → haiku call → 0.33 credits
   - **Best case clean run: ~25 credits** (was ~52.5 before).
   - **Worst case 2 revise loops then surface: ~30-35 credits** (was unbounded before).

## Out of scope

- Per-stage user overrides UI (architecture allows adding later).
- Cost forecasting before commission start.
- Auto-retry with cheaper model on opus rate-limit (steward already has model_limit escalation; this is the inverse and not needed yet).
- Restructuring the verify→revise control flow beyond the loop cap.
