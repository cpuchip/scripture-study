---
workstream: WS5
status: proposed
brain_project: 3
created: 2026-04-21
last_updated: 2026-04-21
---

# Tokenomics 2026

**Status:** Research / Open Questions
**Created:** 2026-04-21 (cleanup-2026-04 Phase 3)
**Owner:** Michael (review pending)

## Background

GitHub Copilot pricing shifted in April 2026. The brain pipeline defaults moved with it:

| Tier | Old | New | Multiplier |
|------|-----|-----|------------|
| Cheap | `raptor-mini` | `gpt-5-mini` | 0x → 0x |
| Smart | `claude-sonnet-4.6` | `claude-sonnet-4.6` | 1x |
| Big | `claude-opus-4.6` (3x) | `claude-opus-4.7` (7.5x) | **2.5× more expensive** |

The escalation chain in `scripts/brain/internal/steward/steward.go` `DefaultConfig()`:

```go
EscalationChain: []ModelTier{
    {Model: "claude-haiku-4.5",  Cost: 0.33},
    {Model: "claude-sonnet-4.6", Cost: 1.0},
    {Model: "claude-opus-4.7",   Cost: 7.5},
},
MaxCostPerEntry: 20.0,
```

## Open Questions

1. **Is `MaxCostPerEntry: 20.0` still right?**
   - Under old pricing: ≈6.6 opus calls before quarantine.
   - Under new pricing: ≈2.6 opus calls. That may be too tight for a stage that escalates twice.
   - Or it may be exactly right, forcing escalations to be more deliberate.

2. **Does the escalation chain still make sense at 7.5x?**
   - Sonnet → opus is now a 7.5× jump, not a 3× jump.
   - Should there be a `gpt-5` tier between sonnet and opus? GPT-5 is also 1x but may have different success rates per dollar.

3. **What's the actual new monthly burn?**
   - Need real usage data from a few days under new pricing before deciding anything.
   - Brain logs `PremiumRequestsUsed` per entry — aggregate across a week.

4. **When does GitHub publish official rate cards?**
   - Current numbers (7.5x for opus-4.7) came from observation, not documentation.
   - Verify against any official page before locking in.

## Out of Scope

- Actually changing `MaxCostPerEntry`, the escalation chain, or pricing logic.
- This proposal is research only. Defer all decisions to Michael.

## Reference

- Cleanup-2026-04 Phase 3: model identifier sweep that surfaced this.
- `scripts/brain/internal/steward/steward.go` (DefaultConfig, defaultModelForStage, modelCost).
- `scripts/brain/internal/steward/commission.go:63` (commission default model).
- `scripts/brain/internal/config/config.go` (`PipelineCheapModel` / `PipelineSmartModel` / `PipelineBigModel`).
