# Proposal: TTS/STT Reader for Studies & Scripture

*Created: March 11, 2026*
*Status: Draft — iteration expected*
*Origin: Need to listen to study documents, scriptures, and notes instead of only reading*
*Related: [Brain Relay](brain-relay.md), brain-app (Flutter), ibeco.me (Go/Dokploy)*

---

## Intent

> Add a "read aloud" capability to the scripture-study ecosystem so Michael can listen to studies, scriptures, lessons, and journal entries — on the phone, in the car, at a desk, or from the web. Also explore STT for dictating new study notes and capturing thoughts by voice.

### Why

- Long-form study documents (like the "nevertheless" study) deserve to be *heard*, not just read
- Driving time is wasted study time without audio
- The brain-app already has `flutter_tts` and `speech_to_text` dependencies — but they use device-native synthesis (robotic, inconsistent across devices)
- AI TTS in 2026 has crossed the "natural enough to listen to comfortably" threshold
- Local/self-hosted means no per-character API fees and no data leaving the ecosystem

### What This Is NOT

- This is not a podcast generator (that's the `podcast` agent)
- This is not cloud TTS with per-character billing
- This is not about voice cloning Michael's voice (though that's possible with Qwen3-TTS)

---

## TTS Model Comparison

Research conducted March 11, 2026 via Exa Search.

### Tier 1: Lightweight / CPU-Only (No GPU Required)

#### Pocket TTS (Kyutai Labs)
- **Parameters:** 100M
- **License:** MIT
- **Stars:** 3,473 on GitHub
- **Hardware:** CPU-only. 2 cores. No GPU needed.
- **Latency:** ~200ms to first audio chunk
- **Speed:** ~6x real-time on MacBook Air M4
- **Size:** Small (pip install, no multi-GB downloads)
- **Languages:** English only (more planned)
- **Voice cloning:** Yes (from short sample)
- **Streaming:** Yes (audio streaming supported)
- **API:** Python API + CLI + HTTP server
- **Browser:** Can run client-side in browser
- **Install:** `pip install pocket-tts` or `uv add pocket-tts`
- **Verdict:** **Top pick for ibeco.me deployment.** Runs on CPU, tiny footprint, MIT license, streaming. Perfect for a VPS without GPU.

#### KittenTTS (KittenML)
- **Parameters:** 15M (nano) to ~25MB total models
- **License:** Apache 2.0
- **Stars:** 23 on GitHub (newer, smaller community)
- **Hardware:** CPU-only. Runs on Raspberry Pi.
- **Size:** Under 25MB total
- **Architecture:** StyleTTS2-based, ONNX inference
- **Model tiers:** Nano, Micro, Mini — size/quality tradeoff
- **Speed:** Faster than real-time on CPU
- **API:** FastAPI server, OpenAI-compatible API
- **Languages:** English
- **Verdict:** **Interesting ultralight option.** Smallest footprint of all. Good for edge/mobile embedding. Newer project with less community validation. Worth testing quality against Pocket TTS.

#### Kokoro (hexgrad)
- **Parameters:** 82M
- **License:** Apache 2.0
- **Architecture:** StyleTTS 2 (decoder-only, no diffusion)
- **Hardware:** CPU-friendly, very fast
- **Speed:** Extremely fast for the quality
- **Languages:** Multiple (not just English)
- **Quality:** Ranked competitively in TTS Arena on Hugging Face
- **Install:** Available via Hugging Face
- **Cost benchmark:** $0.02/1K chars on fal.ai (if using hosted API)
- **Verdict:** **Strong contender.** More established than KittenTTS, lighter than Qwen3. Good middle ground.

### Tier 2: GPU-Preferred (Higher Quality)

#### Qwen3-TTS (Alibaba)
- **Parameters:** 0.6B and 1.7B variants
- **License:** Apache 2.0
- **Hardware:**
  - 0.6B: 4-6 GB VRAM
  - 1.7B: 6-8 GB VRAM (RTX 3080+ recommended, 3090 ideal)
- **Storage:** 2.52 GB (0.6B) / 4.54 GB (1.7B)
- **Latency:** 97ms first-audio (streaming mode)
- **Languages:** 10 languages, 9 dialects, 49 timbres
- **Voice cloning:** 3-second sample → full voice clone
- **Quality:** Outperforms ElevenLabs, MiniMax in benchmarks
- **Training data:** 5M+ hours
- **Install:** `pip install qwen-tts`
- **Streaming:** Dual-track streaming architecture
- **Verdict:** **Best quality, but needs GPU.** The gold standard for self-hosted TTS. Not viable on a CPU-only VPS. Runs great on Michael's RTX 3090 locally.

### Summary Matrix

| Model | Params | GPU? | Quality | Latency | Streaming | Voice Clone | License | Best For |
|-------|--------|------|---------|---------|-----------|-------------|---------|----------|
| **Pocket TTS** | 100M | No | Good | ~200ms | Yes | Yes | MIT | ibeco.me server |
| **KittenTTS** | 15M | No | Decent | Fast | Via API | No | Apache 2.0 | Ultra-edge / mobile |
| **Kokoro** | 82M | No | Good+ | Fast | Yes | No | Apache 2.0 | Balanced local |
| **Qwen3-TTS 0.6B** | 600M | 4-6GB | Great | ~97ms | Yes | Yes | Apache 2.0 | Local w/ GPU |
| **Qwen3-TTS 1.7B** | 1.7B | 6-8GB | Best | ~97ms | Yes | Yes | Apache 2.0 | Local w/ GPU |

---

## STT Model Comparison

The brain-app already uses Flutter's `speech_to_text` (device-native). For higher-quality, offline, self-hosted STT:

### Whisper Variants (OpenAI)

| Variant | Language | Speed | GPU? | Key Feature |
|---------|----------|-------|------|-------------|
| **whisper.cpp** | C++ | 1.8-2.4x faster than faster-whisper | CPU-only | Zero Python deps, static binary, <1GB RAM |
| **faster-whisper** | Python/CTranslate2 | Fast | GPU preferred | Quantization, good Python API |
| **WhisperX** | Python | Fast | GPU preferred | Speaker diarization, word-level timestamps |

**Recommendation:** `whisper.cpp` for server-side (CPU-only, compiles to Go-friendly binary), Flutter `speech_to_text` for real-time mobile input.

---

## Deployment Architecture

### Scenario 1: Local (Michael's PC — RTX 3090)

```
┌─────────────────────────────────────────┐
│  Michael's PC (RTX 3090, 24GB VRAM)     │
│                                         │
│  ┌─────────────┐  ┌──────────────────┐  │
│  │ Qwen3-TTS   │  │ whisper.cpp      │  │
│  │ 1.7B (GPU)  │  │ (CPU, small)     │  │
│  └──────┬──────┘  └────────┬─────────┘  │
│         │                  │            │
│  ┌──────┴──────────────────┴─────────┐  │
│  │  TTS/STT HTTP API (localhost)     │  │
│  │  e.g. FastAPI or Go wrapper       │  │
│  └──────┬────────────────────────────┘  │
│         │                               │
│  ┌──────┴──────┐                        │
│  │ brain.exe   │  ← reads study docs    │
│  └─────────────┘    generates audio     │
└─────────────────────────────────────────┘
```

- **Quality:** Best possible (Qwen3-TTS 1.7B)
- **Latency:** Sub-100ms to first audio
- **No size limits:** Can process entire study documents
- **Hardware:** Already have it — RTX 3090 has 24GB VRAM
- **Use case:** Pre-generate audio files for studies, real-time read-aloud at desk

### Scenario 2: ibeco.me (Dokploy VPS — CPU Only)

```
┌─────────────────────────────────────────┐
│  ibeco.me (Dokploy, SLC region)         │
│  CPU-only VPS, Go backend               │
│                                         │
│  ┌─────────────┐  ┌──────────────────┐  │
│  │ Pocket TTS  │  │ whisper.cpp      │  │
│  │ 100M (CPU)  │  │ small model      │  │
│  └──────┬──────┘  └────────┬─────────┘  │
│         │                  │            │
│  ┌──────┴──────────────────┴─────────┐  │
│  │  TTS/STT microservice             │  │
│  │  (Python sidecar or Go wrapper)   │  │
│  └──────┬────────────────────────────┘  │
│         │                               │
│  ┌──────┴──────┐                        │
│  │ ibecome Go  │  ← serves audio to    │
│  │ backend     │    phone app & web     │
│  └─────────────┘                        │
└─────────────────────────────────────────┘
```

- **Quality:** Good (Pocket TTS is surprisingly natural for 100M params)
- **Latency:** ~200ms first chunk, streams from there
- **Hardware needed:** Minimal. Pocket TTS uses 2 CPU cores. Current VPS should handle it if it has 2+ cores and 2GB+ RAM.
- **Concern:** Python sidecar alongside Go backend. Options:
  - Run Pocket TTS as a separate Docker container on same Dokploy instance
  - Or: use Pocket TTS's HTTP server mode and call it from Go
- **Use case:** On-the-go listening from phone, web reader

### Scenario 3: Hybrid (Best of Both)

```
Phone/Web ──► ibeco.me  ──► is audio cached? ──► Yes → stream it
                                │
                                ▼ No
                         is brain online?
                          /           \
                        Yes            No
                        │               │
                   relay to brain   generate with
                   (Qwen3-TTS)     Pocket TTS
                        │               │
                        ▼               ▼
                   cache audio     cache audio
                   on ibeco.me     on ibeco.me
```

- **Pre-generation:** brain.exe on Michael's PC could pre-generate high-quality audio for published studies using Qwen3-TTS, upload to ibeco.me
- **Fallback:** ibeco.me generates on-demand with Pocket TTS when brain is offline
- **Cache:** Once generated, audio is cached and never regenerated
- **Best quality when available, always available regardless**

---

## Integration Points

### 1. Brain-App (Flutter)

The brain-app already has `flutter_tts` (v4.2.5) and `speech_to_text` (v7.3.0).

**Current state:** Uses device-native TTS (Android/iOS built-in voices).

**Upgrade path:**
- Add a "read study" view that fetches markdown from ibeco.me or local cache
- Stream audio from ibeco.me TTS endpoint (or from brain's local TTS)
- Degrade to device-native TTS if server is unreachable
- STT already works for thought capture; keep as-is unless quality is a problem

### 2. Web Reader (ibeco.me)

**New feature:** A `/read/{document}` route or reader page.

- Renders markdown study documents in a clean, readable view
- "Play" button streams TTS audio
- Audio controls: play/pause, speed (0.75x–2x), skip forward/back
- Progressive loading: start playing while the rest generates

### 3. Study Publishing Pipeline

Currently, studies are published via `scripts/publish/`. Add:

- **Audio generation step:** After markdown→HTML, optionally generate audio
- **Audio stored alongside HTML:** `public/study/nevertheless.mp3` (or chunked `.opus` files)
- **Metadata:** Duration, generation model, timestamp

### 4. Desktop (VS Code / brain.exe)

- brain.exe could expose a local `/tts` endpoint using Qwen3-TTS
- VS Code could have a "read this study" command that pipes to local TTS
- Low priority — desktop reading is less painful than mobile reading

---

## Hardware Requirements

### For ibeco.me (VPS)

| Resource | Pocket TTS Needs | Current VPS (est.) | Action |
|----------|------------------|--------------------|--------|
| CPU cores | 2 | 2-4? | Verify — may be fine |
| RAM | ~512MB-1GB for model | 2-4GB? | Verify — should work |
| Disk | ~500MB for model + deps | Adequate | Fine |
| GPU | Not needed | None | Perfect match |
| Python | 3.10+ | Needs adding? | Add to Docker image |

**Likely no new hardware needed for Pocket TTS on ibeco.me.** Just a Python sidecar container.

### For Local (Michael's PC)

| Resource | Qwen3-TTS 1.7B Needs | Michael's PC | Status |
|----------|----------------------|-------------|--------|
| GPU VRAM | 6-8 GB | RTX 3090 (24GB) | Plenty |
| RAM | 8+ GB | Plenty | Fine |
| Disk | ~5 GB | Plenty | Fine |
| CUDA | 12.1+ | Likely yes | Verify |
| Python | 3.10+ | Yes | Fine |

**No new hardware needed for local TTS.**

### If We Want GPU on the Server (Future)

If Pocket TTS quality isn't sufficient and we want Qwen3-TTS on ibeco.me:

| Option | Monthly Cost (est.) | VRAM | Notes |
|--------|--------------------:|------|-------|
| GPU VPS (RTX 4090) | $150-300/mo | 24 GB | Overkill for personal use |
| RunPod serverless | $0.40-0.80/hr | 24 GB | Pay per use, cold starts |
| Lambda Cloud | $0.50-1.00/hr | 24 GB | Similar |
| Home server + Cloudflare Tunnel | $0 ongoing (own hardware) | Your GPU | Best long-term if you have a spare GPU |

**Recommendation: Start CPU-only (Pocket TTS). Only add GPU if quality isn't good enough after testing.**

---

## Proposed Implementation Phases

### Phase 0: Evaluate (This Iteration)
- [ ] Install Pocket TTS locally, test quality on a sample study document
- [ ] Install Qwen3-TTS locally (GPU), test quality on same document
- [ ] Install KittenTTS locally, test quality on same document
- [ ] Compare: naturalness, pacing, pronunciation of scripture-specific words
- [ ] Test Kokoro if the above don't satisfy
- [ ] Record findings, pick TTS engine(s) for Phase 1

### Phase 1: Local Reader
- [ ] Add TTS endpoint to brain.exe (or standalone service)
- [ ] Markdown → clean text → TTS → audio stream
- [ ] Test with a real study document end-to-end
- [ ] Brain-app: add "read" button for entries that plays audio from brain's TTS

### Phase 2: Server Reader
- [ ] Deploy Pocket TTS (or winner from Phase 0) as Docker sidecar on ibeco.me
- [ ] Add `/api/tts` endpoint to ibecome backend
- [ ] Audio caching layer (generate once, serve forever)
- [ ] Brain-app: fall back to server TTS when brain is offline

### Phase 3: Web Reader
- [ ] Add reader page to ibeco.me web interface
- [ ] Audio player controls (play/pause/speed/scrub)
- [ ] Markdown rendering + synchronized audio

### Phase 4: Audio Publishing
- [ ] Integrate into publish pipeline (`scripts/publish/`)
- [ ] Pre-generate audio for published studies
- [ ] Serve alongside HTML in `public/`

---

## Open Questions

1. **Quality bar:** How natural does TTS need to be before it's comfortable for long-form scripture study listening? Only testing will tell.
2. **Scripture pronunciation:** Will these models handle "Nephi," "Lehi," "Moroni," "Melchizedek" etc.? Likely not natively — may need a pronunciation dictionary or fine-tuning.
3. **Voice preference:** Do we want a specific voice character? Warm, male, calm? Pocket TTS and Qwen3-TTS both support voice selection/cloning.
4. **Caching strategy:** Cache full-document audio? Or chunk by section/paragraph for faster initial playback?
5. **STT priority:** Is STT needed beyond what `speech_to_text` already provides in the brain-app? Or is this mostly a TTS initiative?
6. **VPS specs:** Need to verify actual CPU/RAM on the ibeco.me Dokploy instance to confirm Pocket TTS fits.
7. **Piper TTS:** Another CPU-friendly option (Mozilla-backed, Rust-based) — worth testing if Pocket TTS quality falls short.

---

## References

- [Pocket TTS GitHub](https://github.com/kyutai-labs/pocket-tts) — MIT, 100M params, CPU-only
- [KittenTTS GitHub](https://github.com/soldier444xd/KittenTTS) — Apache 2.0, 15M params, <25MB
- [Qwen3-TTS Blog](https://qwen.ai/blog?id=qwen3-tts-1128) — Apache 2.0, 0.6B-1.7B, GPU
- [Kokoro on HuggingFace](https://huggingface.co/hexgrad/Kokoro-82M) — Apache 2.0, 82M params
- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) — MIT, C++ Whisper port
- [Kyutai Pocket TTS Technical Report](https://kyutai.org/pocket-tts-technical-report)
