# Context Engineering — Dev Stream Handoff

*Self-contained spec for parallel execution. No access to the planning conversation needed.*

---

## Binding Problem

The conference reindex test harness (`experiments/lm-studio/scripts/`) currently sends a minimal system prompt (~250 tokens) and a prompt template with content. The model (nemotron-3-nano at 131k context) scores surface-level because it lacks gospel vocabulary and TITSW principle definitions in its context window.

This dev stream updates the harness and prompt to support a richer context package — while a parallel study stream curates the actual context documents.

---

## Deliverables

### 1. Update `run-test.ps1` — Add `-Context` Parameter

**File:** `experiments/lm-studio/scripts/run-test.ps1`

**Current behavior:** The system message is loaded from `context.md` only (line ~83):
```powershell
$systemMessage = Get-Content $contextFile -Raw -Encoding UTF8
```

**Required change:** Add a `-Context` parameter that accepts a directory path. If provided, load all `.md` files from that directory (sorted alphabetically) and append them to the system message after `context.md`.

```powershell
# New parameter
[string]$Context = ""

# After loading context.md:
if ($Context) {
    $contextDir = Join-Path $scriptDir $Context
    if (-not (Test-Path $contextDir)) {
        Write-Error "Context directory not found: $contextDir"
        return
    }
    $contextFiles = Get-ChildItem $contextDir -Filter "*.md" | Sort-Object Name
    foreach ($cf in $contextFiles) {
        $systemMessage += "`n`n" + (Get-Content $cf.FullName -Raw -Encoding UTF8)
    }
    Write-Host "Context: $($contextFiles.Count) files from $Context" -ForegroundColor Cyan
}
```

**Usage:**
```powershell
.\run-test.ps1 -Prompt titsw-v3 -Content alma-32 -Model nemotron-3-nano -Context context
```

This loads `context.md` (base) + all files from `experiments/lm-studio/scripts/context/*.md` into the system message.

**Directory:** Create `experiments/lm-studio/scripts/context/` for the context package files. The study stream will deliver `gospel-vocab.md` and `titsw-framework.md` here. Dev can create placeholder files for testing the harness.

### 2. Add `cache_prompt: true` to Request Body

**File:** `experiments/lm-studio/scripts/run-test.ps1`

**Current request body** (around line ~145):
```powershell
$requestBody = @{
    model = $modelId
    messages = @(
        @{ role = "system"; content = $systemMessage }
        @{ role = "user"; content = $userMessage }
    )
    max_tokens = $MaxTokens
    temperature = $Temperature
    stream = $true
    stream_options = @{ include_usage = $true }
} | ConvertTo-Json -Depth 10
```

**Add `cache_prompt`:**
```powershell
$requestBody = @{
    model = $modelId
    messages = @(
        @{ role = "system"; content = $systemMessage }
        @{ role = "user"; content = $userMessage }
    )
    max_tokens = $MaxTokens
    temperature = $Temperature
    stream = $true
    stream_options = @{ include_usage = $true }
    cache_prompt = $true
} | ConvertTo-Json -Depth 10
```

**Why:** llama.cpp (LM Studio's backend) stores the KV cache per slot. When `cache_prompt: true` is set, matching token prefixes are skipped on subsequent requests — zero recomputation. The system message (~8,000 tokens) is identical across all talk evaluations, so requests 2+ skip all system token prefill.

**Technical details:**
- llama.cpp auto-assigns slots by prefix similarity (default `-sps 0.5`)
- The system prompt must be **byte-identical** across requests for cache hits
- Confirmed by ggerganov in llama.cpp discussion #8860
- Works with the OpenAI-compat `/v1/chat/completions` endpoint we already use

### 3. Write TITSW v3 Prompt

**File:** `experiments/lm-studio/scripts/prompts/titsw-v3.md`

**Changes from v2** (`prompts/titsw-v2.md`):

1. **Scale: 0-9** (was 0-3). Rubric anchors:
   - 0: Not present
   - 1-2: Incidental/minor
   - 3-4: Present but not a focus
   - 5-6: Intentional and significant
   - 7-8: Central to the teaching approach
   - 9: Defining — would be the textbook example

2. **Anti-inflation language** (add to rubric section):
   > "A score of 7+ means this content could be used as a teaching example for this principle. Most conference talks score 4-6 on most dimensions. Reserve 8-9 for content that is genuinely exceptional."

3. **Reference-aware instruction** (add after the principles section):
   > "If `<references>` are provided below the content, use them to inform your scoring. Cross-references that reveal deeper Christ connections should increase the `teach_about_christ` and `help_come_unto_christ` scores. Score based on the full available context, not just surface text."

4. **New JSON fields** — add to the output schema:
   ```json
   {
     "typological_depth": 0-9,
     "cross_reference_density": 0,
     "surface_vs_deep_delta": {
       "teach_about_christ": "explanation if informed reading changes score",
       "help_come_unto_christ": "explanation if informed reading changes score"
     }
   }
   ```
   - `typological_depth`: How much hidden Christ-typology exists beyond surface. 0 = what you see is what you get. 9 = the entire passage is a sustained type/shadow of Christ.
   - `cross_reference_density`: Count of explicit scripture/prophetic citations in the content.
   - `surface_vs_deep_delta`: For the two Christ-centered meta-principles, note whether the context package (gospel-vocab, references) changes the score vs surface reading only.

5. **All v2 content preserved** — the principle definitions, scoring rubric strictness language, specific examples requirement, "cite actual phrases with verse numbers" instruction. v3 extends v2, doesn't replace the core.

6. **Keep the `{{CONTENT}}` placeholder** at the end, same as v2.

### 4. Update `run-suite.ps1` (if needed)

**File:** `experiments/lm-studio/scripts/run-suite.ps1`

If this script calls `run-test.ps1`, it needs to pass through the `-Context` parameter. Check how it invokes run-test and add `-Context $Context` forwarding.

---

## Validation (requires study stream context files)

Once `gospel-vocab.md` and `titsw-framework.md` are delivered by the study stream:

### Test Cases

| # | Content | Prompt | Context | Expected Result |
|---|---------|--------|---------|-----------------|
| 1 | `alma-32` | `titsw-v3` | `context` | `teach_about_christ` ≥ 5 (was 1-2 without context) |
| 2 | `kearon-receive-his-gift` | `titsw-v3` | `context` | Scores stable vs v2 baseline, no inflation |
| 3 | `kearon-receive-his-gift` | `titsw-v3` | *(none)* | Similar to v2 scores (scale change only) |

### Ground Truth Reference

**File:** `experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md`

Contains hand-scored ground truth with reasoning:

**Alma 32 (informed):**
- teach_about_christ: 7-8
- help_come_unto_christ: 8
- love: 7-8
- spirit: 3-4
- doctrine: 8-9
- invite: 8-9

**Kearon (single score):**
- teach_about_christ: 8
- help_come_unto_christ: 8
- love: 4
- spirit: 3
- doctrine: 7
- invite: 7

### Success Criteria

1. Alma 32 `teach_about_christ` ≥ 5 with context (proves context engineering works)
2. Kearon scores within ±1 of ground truth on each dimension
3. No systematic inflation (average scores shouldn't jump 2+ points across the board)
4. `cache_prompt: true` doesn't break anything (same outputs with/without)

---

## File Inventory

| File | Status | Action |
|------|--------|--------|
| `experiments/lm-studio/scripts/run-test.ps1` | EXISTS | Edit: add `-Context` param + `cache_prompt: true` |
| `experiments/lm-studio/scripts/run-suite.ps1` | EXISTS | Check/edit: forward `-Context` param |
| `experiments/lm-studio/scripts/context.md` | EXISTS | No change (Layer 1 — base system context) |
| `experiments/lm-studio/scripts/context/` | NEW | Create directory for context package |
| `experiments/lm-studio/scripts/context/titsw-framework.md` | NEW | Study stream delivers this (Layer 2) |
| `experiments/lm-studio/scripts/context/gospel-vocab.md` | NEW | Study stream delivers this (Layer 3) |
| `experiments/lm-studio/scripts/prompts/titsw-v3.md` | NEW | Create: 0-9 scale, reference-aware |
| `experiments/lm-studio/scripts/prompts/titsw-v2.md` | EXISTS | Keep as baseline for comparison |
| `experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md` | EXISTS | Reference only — do not modify |

---

## Technical Context

- **Model:** nemotron-3-nano (31.6B total, ~3.6B active MoE). Loaded at 131k context. 24.52 GB VRAM on dual RTX 4090.
- **Performance:** 170-180 gen tok/s, 18.5s avg per talk evaluation.
- **API:** LM Studio at `localhost:1234`, OpenAI-compatible `/v1/chat/completions` endpoint.
- **Streaming:** SSE via `HttpWebRequest` (not `Invoke-RestMethod`) for token-by-token streaming.
- **Result storage:** JSON files in `results/`, summary rows in `results.tsv`.

---

## Sequencing

1. **Immediately:** Add `-Context` parameter to `run-test.ps1` + create `context/` directory with placeholder files
2. **Immediately:** Add `cache_prompt: true` to request body
3. **Immediately:** Write `titsw-v3.md` prompt
4. **After study stream delivers:** Run validation tests against ground truth
5. **If validation passes:** Update `run-suite.ps1` for batch context support
