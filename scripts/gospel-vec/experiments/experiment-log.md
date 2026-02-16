# gospel-vec Experiment Log

This log tracks model experiments for gospel-vec's embedding and summary quality.
Results are auto-appended by `experiments/run-experiment.ps1`.

## How to Read This Log

- **Recall** = what fraction of expected relevant results were found in the top results
- **Perfect Recall** = queries where ALL expected results were found
- **Zero Recall** = queries where NONE of the expected results were found
- Higher recall is better, but also look at the per-query breakdown for patterns

## Comparing Experiments

When comparing two experiments, look for:
1. **Overall average recall** — which model finds more of what we expect?
2. **Per-query patterns** — does one model excel at narrative queries but fail at doctrinal ones?
3. **Score distribution** — are results clustered near the top (confident) or spread thin?
4. **Category performance** — doctrinal-concept vs narrative vs prophecy

---

