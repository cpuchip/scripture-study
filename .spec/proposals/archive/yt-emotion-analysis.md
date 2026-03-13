# Proposal: YouTube Audio Emotion Analysis

*Created: March 11, 2026*
*Status: Idea — needs exploration*
*Related: [YT Evaluation Template](../../docs/yt_evaluation_template.md), yt-mcp, chip-voice*

---

## Intent

> When evaluating YouTube content (conference talks, commentary, podcasts), analyze not just *what* is said but *how* it's delivered. Detect emotional tone, sincerity, aggression, warmth, sarcasm, urgency — the delivery layer that shapes how a message lands. Was it said with malice? With genuine care? With manufactured outrage?

### Why This Matters

- The same words said with contempt vs. compassion produce opposite effects
- Our yt eval process currently transcribes and analyzes text — but the *tone* of delivery is invisible
- Conference talks convey truth partly through the Spirit, but also through measured, warm, sincere delivery
- YouTube commentary and "gospel" content often uses emotionally manipulative delivery — urgency, fear, anger — even when the words themselves sound doctrinally correct
- Detecting this programmatically would add a powerful dimension to evaluations

### Use Cases

1. **YouTube evaluations:** Flag segments where delivery shifts to anger, contempt, or fear-mongering
2. **Conference talk analysis:** Measure the emotional arc of a talk — where does the speaker shift from teaching to testifying?
3. **Comparison:** Put a manipulative YouTube "prophet" next to Elder Holland and show the emotional signature difference
4. **Self-awareness:** Analyze your own recorded talks/lessons for emotional tone

---

## Technical Approach

### Step 1: Get Audio from YouTube

We already have `yt-mcp` for downloading transcripts. Extend or complement it to download audio.

| Tool | What It Does | Notes |
|------|-------------|-------|
| **yt-dlp** | Download YouTube audio as mp3/opus/wav | Already the standard. `yt-dlp -x --audio-format mp3 URL` |
| **yt-mcp** | Our existing MCP server | Could add an audio download action alongside transcript |

### Step 2: Speech Emotion Recognition (SER)

Pre-trained models that classify emotion from audio — no training needed.

| Model | Source | Emotions | Architecture | Notes |
|-------|--------|----------|-------------|-------|
| **wav2vec2-IEMOCAP** (SpeechBrain) | [HuggingFace](https://huggingface.co/speechbrain/emotion-recognition-wav2vec2-IEMOCAP) | Angry, Happy, Sad, Neutral | wav2vec2 fine-tuned | Most established. Good accuracy on conversational speech. |
| **emotion2vec** | HuggingFace | 8 emotions + valence/arousal | Self-supervised | Newer, strong benchmarks. Multiple model sizes. |
| **Emonity** | [GitHub](https://github.com/sv6095/Emonity) | Multiple emotions | CNN-BiLSTM, MFCCs + spectrograms | PyTorch-based, multi-dataset training |
| **Librosa + custom** | DIY | Configurable | MFCC/Chroma feature extraction → classifier | More work, more control |

**Recommendation:** Start with **SpeechBrain wav2vec2-IEMOCAP** — it's pip-installable, well-documented, and runs on CPU. Upgrade to emotion2vec if we need finer granularity.

### Step 3: Segment-Level Analysis

Don't just classify whole videos. Chunk audio into segments (10-30 seconds each) and classify each:

```
Video: "Why the Church is Wrong About X" (32 min)

Timestamp    | Emotion    | Confidence | Text (from transcript)
-------------|------------|------------|------------------------
0:00-0:30    | Neutral    | 0.82       | "Hey everyone, today we're going to..."
0:30-1:15    | Angry      | 0.71       | "I can't believe they would say..."
1:15-2:00    | Contempt   | 0.65       | "These people don't even understand..."
...
28:00-28:30  | Fear       | 0.78       | "If you don't wake up now..."
```

### Step 4: Emotional Signature / Summary

Aggregate into a profile:

```
Overall Tone Distribution:
  Neutral:  35%
  Angry:    25%
  Contempt: 15%
  Fear:     12%
  Sad:       8%
  Happy:     5%

Emotional Shifts: 14 (high volatility)
Sustained Anger Segments: 3 (>2 min each)
Peak Intensity: 1:15-2:00, 28:00-28:30

Assessment: Delivery relies heavily on anger and fear.
            Message may have valid points but delivery is manipulative.
```

---

## Integration with YT Eval Workflow

Currently the `eval` agent produces evaluations like:

> "This video claims X. Compared to [scripture], this is partially correct but missing Y."

With emotion analysis added:

> "This video claims X. Compared to [scripture], this is partially correct but missing Y. **The delivery shifts to sustained anger at 1:15 and uses fear language at 28:00. The emotional signature (25% anger, 15% contempt) suggests persuasion through agitation rather than invitation through the Spirit.**"

---

## Resource Requirements

| Component | CPU-only? | Size | Notes |
|-----------|-----------|------|-------|
| yt-dlp | Yes | Small | Already used by yt-mcp |
| wav2vec2-IEMOCAP | Yes (slow) / GPU preferred | ~1.2GB | Fine for batch processing |
| librosa | Yes | Small | Audio feature extraction |
| ffmpeg | Yes | Small | Audio format conversion |

Could run on ibeco.me (CPU, batch mode) or locally (GPU, faster).

---

## Open Questions

1. **Accuracy on religious speech:** SER models are trained on conversational/acted datasets (IEMOCAP, RAVDESS). How well do they transfer to conference talks and YouTube commentary? Needs testing.
2. **Sarcasm / manipulation:** These are harder than basic emotion. May need a second-pass analysis combining text sentiment + audio emotion to detect mismatch (saying nice words angrily).
3. **Cultural calibration:** A passionate testimony might read as "high arousal" — that's not the same as anger. Need to calibrate for religious speech patterns.
4. **Privacy / ethics:** Only analyze publicly-available content. Don't profile individuals — profile delivery in specific videos.
5. **Integration point:** Does this become part of yt-mcp? A separate tool? A chip-voice sibling?

---

## Phases

### Phase 0: Proof of Concept
- [ ] Download audio from a YouTube video via yt-dlp
- [ ] Run SpeechBrain wav2vec2 emotion detection on segments
- [ ] Compare results on: a General Conference talk vs. a sensationalist YouTube video
- [ ] Evaluate whether the output is useful or noise

### Phase 1: Integration
- [ ] Add audio download to yt-mcp (or create yt-audio tool)
- [ ] Build segment-level emotion pipeline
- [ ] Add "emotional signature" section to yt eval template

### Phase 2: Refinement
- [ ] Test emotion2vec for finer granularity
- [ ] Build text+audio mismatch detection (says kind words angrily)
- [ ] Calibrate for religious speech patterns
