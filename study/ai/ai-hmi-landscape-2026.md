# AI Human-Machine Interface Landscape — mid-2026

*Web research survey, 2026-07-09. Binding question: beyond voice-only turn-taking, what is the state of the art in AI human-machine interfaces, and what is practically adoptable for a self-hosted, privacy-conscious system like **spin** (voice front → loom-driven Claude seats on the pg-ai-stewards substrate)?*

Scope: realtime voice, vision-in, computer-use / screen-share, video-out / avatars, ambient / always-on patterns. Every named product and number below traces to a source in the Sources section that I fetched or saw in search results this session. Where a figure came from a secondary blog rather than a primary doc, it is marked. Where I could not verify a memory, it is dropped or flagged.

---

## For spin — recommendations (read this first)

**The constraint that decides everything: Claude has no speech-to-speech realtime API.** Claude Code voice mode and the Claude apps' voice mode are speech-to-*text* dictation with push-to-talk — audio goes up, text comes back, Claude does not stream speech and does not support barge-in. Claude Code's dictation is not even available when you authenticate with an API key, Bedrock, Vertex, or Foundry; it requires a Claude.ai account. Therefore a Claude brain can only enter a realtime voice loop the way spin already does it: a fast local layer owns the audio and hands text down to Claude. Nobody ships this off the shelf. It is spin's actual differentiation, not a gap.

The corollary: the two vendors who *do* have polished speech-to-speech (OpenAI Realtime, Gemini Live) can't run a Claude Max brain, are cloud-only, and cost roughly an order of magnitude more per minute than a local cascade. So spin's "rent the audio plumbing, own the brain" architecture is the right call for a Claude-brained, privacy-conscious system. The move is not to switch to an S2S API; it is to reproduce the *feel* of one on the local cascade.

**Highest-leverage additions, in order:**

1. **Turn-taking and latency tuning on the existing cascade (do this first, it's cheap).** The single thing that separates "walkie-talkie" from "Jarvis" is turn-taking, not model quality. Adopt Pipecat's `SmartTurnDetection` (semantic end-of-turn, an LLM-family classifier) on top of Silero VAD, and tune to the published conversational budget: barge-in under ~150 ms from end-of-speech to TTS flush, turn gap 200–450 ms, p95 end-to-end under 800 ms. This is exactly what a speech-to-speech API buys you, reproduced locally for the cost of a small model plus config. Since spin is already Pipecat-family with Kokoro, this is not a rewrite.

2. **Wire the pg-ai-stewards brain in over MCP with async "think while talking" (the real Jarvis step).** This is spin's own Tier 3, and the research reinforces it: because Claude has no S2S API, spin's split — fast reflex model up front, text handoff to Claude down — is the *only* way to put a Claude brain in a live voice loop. Curate a few substrate tools into spin's `mcp_servers` (consult/dispatch/remember), not the full ~50, so the front model doesn't drown. Use the async pattern (respond quick, dig while the brain works) for the slow brain calls where it finally belongs.

3. **Add "eyes" as a callable tool, not an always-on video pipe (later, optional).** For "look at my screen / look at this," feed frames at ≤1 FPS to a cheap vision model only when invoked, gated by region-of-change so you never pay per idle frame. Two stacks: quickest is Gemini Live (native ≤1 FPS video + screen share) as a side-channel; the fully-owned path is Vision Agents by Stream (open source, WebRTC, YOLO/Roboflow pre-filter before the LLM). Do the tool-call version before any continuous-vision build.

**Skip for now:** talking-head avatars (Q4). Real product category, wrong tool for a personal daily driver — they add latency and cloud dependency for near-zero utility. Voice + a good Vue UI (spin's Tier 2) beats a face. If a face is ever wanted, `TalkingHead` (3D, browser, MIT-ish) or Open-LLM-VTuber (Live2D) point their LLM backend at the same spin endpoint, no architectural change.

**Stack picks at a glance:** STT — keep faster-whisper `large-v3-turbo`, evaluate Parakeet-TDT (NVIDIA, Apache-2.0) or Moonshine v2 for streaming. TTS — keep Kokoro for speed, keep the `ScriptureNameFilter` plan for pronunciation (still the right engine-agnostic fix), evaluate Piper only if you need phoneme control without the filter. Turn detection — Pipecat SmartTurn + Silero VAD. Wake word (if going always-on) — openWakeWord / microWakeWord, "Hey Jarvis" ships out of the box. Brain — pg-ai-stewards over a curated MCP subset. Eyes — Gemini Live tool-call first, Vision Agents by Stream if you want it fully owned.

---

## 1. Voice: state of the art for realtime voice agents

### Two architectures, and the gap between them is closing

The field splits into **speech-to-speech** (one model ingests audio and emits audio) and **cascaded** (STT → LLM → TTS, three swappable stages). Through 2025 the cascade won on flexibility and lost on naturalness. In 2026 the gap is narrowing from both directions: S2S APIs got cheaper and more controllable, and cascades got semantic turn detection that removes most of the "talks over you" problem.

**Speech-to-speech APIs:**

- **OpenAI Realtime API** is generally available. The current model family is `gpt-realtime-2.1` (voice agent), `gpt-realtime-translate`, and `gpt-realtime-whisper` (transcription), per OpenAI's official developer guide. Transports are WebRTC (browser/mobile), WebSocket (server with an existing media pipeline), and SIP (telephony). A secondary guide reports median first-token latency under ~300 ms and all-in production cost around $0.25–$0.35/min with caching; treat those as reported, not official. The load-bearing point for spin: S2S is roughly an order of magnitude pricier per minute than a local cascade, and it can't host a Claude brain.

- **Gemini Live API** is the most vision-native realtime API. Per Google's official docs it is a stateful WebSocket (WSS) API taking audio in (16-bit PCM, 16 kHz), images at ≤1 FPS, and text; emitting audio out (16-bit PCM, 24 kHz). The Gemini Developer API docs still label it **Preview**; secondary sources report it reached **GA on Vertex AI** at I/O 2026 with native-audio models. The model LiveKit's own recipe wires is `gemini-2.5-flash-native-audio-preview-12-2025`. It holds an open connection, supports interruption, and can see a camera or shared screen while talking — the one realtime API that does audio + live video together.

- **Anthropic / Claude: no speech-to-speech.** Claude's voice is STT dictation only. Claude Code `/voice` (v2.1.69+, March 2026 rollout) streams recorded audio to Anthropic for transcription into the prompt; Claude does not speak back and there is no barge-in. The consumer app voice mode (iOS/Android, GA June 2025, 18 languages by June 2026) is likewise push-to-talk with no realtime interruption. This is the single most important fact for spin's architecture.

**Cascaded / open-source stacks** remain the right base for local-first, and 2026 made them competitive on feel:

- **Pipecat** (Daily, fully OSS, v1.0 April 2026) is the pipeline-first framework: a DAG of frame processors, 60+ service integrations, swap any STT/LLM/TTS without touching the rest. Its `SmartTurnDetection` feeds partial transcripts to a small classifier that predicts turn completion, cutting false interruptions ~30% versus pure-VAD, per the comparison writeups. This is spin's family.
- **LiveKit Agents** is infrastructure-first: your agent joins a WebRTC "room" as a participant, so multi-participant, video, and screen share are native rather than bolted on. Heavier infra; the room model is overkill for single-user local but is the cleaner base if spin ever wants multi-party or native video.
- **Vocode** and **TEN** exist in the same space; the two that matter for spin are Pipecat (control) and LiveKit (transport/rooms).

### Component models for a local cascade (the RTX 4090 tier)

**STT.** faster-whisper `large-v3-turbo` remains the quality baseline. For *streaming* specifically, two 2026 options beat Whisper's latency hard: **NVIDIA Parakeet-TDT** (0.6–1.1B, Apache-2.0, RTFx above ~2000, fastest raw throughput) and **Moonshine v2** (built for streaming — words appear as you speak with minimal token revision; Tiny ~50 ms, Small ~148 ms, Medium ~258 ms per the benchmarks, smallest model 27 MB). Whisper Large v3 by contrast measured in the seconds on the same streaming test. For spin, Parakeet is the strongest fit (GPU, permissive license, streaming); Moonshine is the edge/CPU option.

**TTS.** Quality has plateaued — the top local engines cluster MOS 4.5–4.7, so **latency and pronunciation control are the real differentiators**, not naturalness. **Kokoro-82M** (Apache-2.0, ~80 ms time-to-first-audio, MOS ~4.5, English-only, no in-engine pronunciation control) is the speed/quality-per-param leader and spin's current engine. **Piper** (Rhasspy, CPU, espeak-ng phonemes underneath) is the pronunciation-control option. **F5-TTS** is the leading open zero-shot cloner (3-second reference). **Orpheus-3B** and **Chatterbox** are the newer expressive, permissively-licensed options. XTTS v2 stays broadest for languages but is non-commercial. This validates spin's `ScriptureNameFilter` plan: fix pronunciation with an engine-agnostic phonetic-respelling text filter, treat any engine swap as a separate quality decision.

**VAD / turn / wake.** Silero VAD (sub-5 ms CPU) plus a small semantic turn detector (Pipecat SmartTurn class, ~100M+ params) is the standard 2026 stack. Wake words: openWakeWord and microWakeWord (the latter runs on ESP32-S3), with "Hey Jarvis" trained and shipping out of the box in Home Assistant.

### Latency budgets — what "conversational" means in 2026

The numbers are now specific and agreed across the practitioner writeups. Sub-800 ms at p95 end-to-end is the reliability floor below which conversation feels natural. The layer budget: STT 60–120 ms, LLM first-token 100–250 ms, TTS first-chunk 40–100 ms, network 20–60 ms. Barge-in should fire under ~150 ms from end-of-user-speech to TTS flush. The turn-taking gap from end-of-agent-speech to first-audio-of-next-turn should land 200–450 ms; humans converse at ~200–300 ms gaps, and up to ~500 ms still reads as natural. Pure-VAD agents lag at 800–1500 ms because they wait on silence; semantic turn detection closes that to ~300 ms without cutting people off. **This budget is the spec spin's turn-tuning work should hit.**

**Best local-first voice pipeline today:** Pipecat + Silero VAD + SmartTurnDetection + Parakeet/faster-whisper STT + Kokoro TTS, with a fast local LLM as the reflex tier and a text handoff to a heavier brain for real cognition. That is essentially spin's stack; the delta is turn-detection tuning to the budget above.

---

## 2. Video / vision in: assistants that can SEE

### The shipping products

**Gemini Live** is the clear leader for realtime "look at this": it takes a camera feed or shared screen alongside audio in one session, at ≤1 FPS image input. **ChatGPT** (advanced voice + vision) and **Microsoft Copilot Vision** offer camera/screen understanding in their apps. The pattern that generalizes: continuous vision is fed as *sampled frames* (≤1 FPS is typical and often enough), not a full video stream, because adjacent frames are nearly identical and the interesting event is rare.

### How continuous vision is fed to models affordably

This is the real engineering question, and 2026 has converged on a few techniques:

- **Cap the frame rate.** Gemini Live's own limit is 1 FPS. A surveillance-style example makes the point: 1 FPS over 24 h is 86,400 frames but the answer may hinge on three of them. Uniform sampling is cheap but drops the frames that matter, so adaptive selection is needed.
- **Temporal redundancy removal.** Sample at 1 FPS, compute lightweight features (e.g. DINOv2) over short windows, drop frames too similar to their neighbors — one reported pipeline keeps ~45.9% of frames after this stage.
- **Cascade: cheap detector before the expensive model.** Run a task-specific detector (YOLO26 / RT-DETR, ByteTrack for tracking) at the edge first; escalate only ambiguous or high-value frames to a VLM. Task-specific models beat VLMs by one to two orders of magnitude in cost and latency, and YOLO inference itself is ~5–15 ms. Most of the latency is capture/encode/transport, not inference.
- **Streaming-VLM memory.** Newer architectures (StreamingVLM-class) treat unbounded live video as continuous context via specialized KV/memory caches rather than re-encoding snapshots, so the agent stays situationally aware without re-paying for history.

**For spin:** don't build a continuous vision pipe. Make "eyes" a tool the brain calls — capture a frame (or screen region) on demand, gate on region-of-change so idle screens cost nothing, send ≤1 FPS to a cheap VLM (Gemini Flash cloud, or a local VLM on the spare GPU). Vision Agents by Stream (Q3/§6) is the pre-built open-source version of exactly this cascade if you want the continuous path later.

---

## 3. Desktop sharing / computer use

Two distinct axes here: **the agent operates the computer**, and **the user shares a screen so the agent can help**. Both matured in 2026.

### Agent-operates-the-computer

- **Anthropic computer use** exposes a generic tool: the model receives screenshots and returns input actions (click/type/scroll). It is portable — the same tool runs in a Docker container or on a Mac with full desktop control. **Claude for Chrome** launched as a Max-only preview in August 2025 and expanded to Pro/Team/Enterprise in December 2025.
- **OpenAI** folded standalone Operator into ChatGPT Agent in 2025; its Computer-Using Agent works primarily from screenshots (pixel approach) and does not read the accessibility tree. Codex added background computer use (April 2026) running agents in their own macOS desktop sessions.
- **Google Gemini Computer Use** descends from Project Mariner, runs as a Chrome extension, and privileges DOM/accessibility awareness over raw pixels — browser-scoped.
- **Microsoft** made computer-using agents GA in Copilot Studio on May 13, 2026. Its research system **UFO2** ("Desktop AgentOS") uses a hybrid control-detection pipeline: query Windows UI Automation (the accessibility tree) first for a semantically rich, high-precision control list, then fall back to vision parsing for what the tree can't see.

**The architectural fault line — accessibility tree vs. pixels vs. hybrid:**

- **Accessibility-tree** (Microsoft Playwright MCP, Gemini): structured snapshot of controls, no vision model needed, far cheaper in tokens, but blind to canvas-rendered and dense-visual UIs.
- **Pixel/screenshot** (OpenAI CUA, Anthropic computer use): works on anything a human can see, but each screenshot spends a large token budget the model must then interpret.
- **Hybrid** (UFO2): accessibility tree for the bulk of the screen, vision for the parts it can't capture. This is the direction the serious desktop-automation work is heading.

### User-shares-screen-for-help

The plumbing is the same realtime-media stack as voice: **WebRTC** carries the screen/camera track, and the framework hands sampled frames to the model. Pipecat ships a browser agent with screen-sharing built on its Voice UI Kit + Daily WebRTC transport. LiveKit enables it with `RoomOptions(video_input=True)` feeding frames to a Gemini realtime model. Gemini Live takes a shared screen natively at ≤1 FPS. Nobody streams raw RDP/VNC into a model; the pattern is WebRTC track → frame sampling → model.

**For spin:** the "user shares screen so Spin can help" axis is the natural, low-risk one and reuses the same WebRTC frame-sampling as §2. The "Spin operates my computer" axis is higher-risk and higher-value; if pursued, Anthropic computer use is the portable, Claude-native option (Docker-sandboxed), and the accessibility-tree-first hybrid is the cheaper, more reliable design. Gate any computer-operating capability behind the substrate's presiding/council rules — it is an outward, irreversible surface.

---

## 4. Video out / presence: avatars, and whether faces matter

The talking-head field is real and commercial, but it is **not** where personal-assistant leverage lives.

- **HeyGen LiveAvatar** is realtime two-way: a WebRTC-streamed avatar that listens, responds, and lip-syncs live, and lets you connect your own LLM. HeyGen also ships the Avatar IV API for generated video.
- **Tavus** builds a "digital twin" from ~2 minutes of footage and generates personalized video at scale; **D-ID** offers real-time streaming avatars for conversational agents; Synthesia is the enterprise library play.
- **Open-source:** `TalkingHead` (browser JS, 3D GLB avatar, real-time lip-sync + emoji-to-expression, supports full-body GLB + Mixamo FBX; open-source 3D lip-sync writeups appeared May 2026). **Open-LLM-VTuber** (Live2D, fully offline, pluggable LLM/ASR/TTS) and **Amica** (VRM 3D). **Wav2Lip** and variants run locally for lip-sync without a subscription.

**Is the field converging on voice+UI or faces?** For customer-facing kiosks, sales, and marketing video, faces are a growing product category. For a personal daily-driver assistant, the center of gravity is voice + a good UI (transcript, panels, ambient presence), with a face as optional polish. A face adds render latency, cloud dependency (for the good ones), and an uncanny-valley tax, for little added utility on a system you talk to all day.

**For spin:** skip the avatar. Spend the same effort on the Vue3 UI (Tier 2) — it is the thing that makes you want to *live* in Spin. If a face is ever wanted, the open-source options point their LLM backend at spin's existing endpoint, so it stays a cosmetic layer, not a rearchitecture.

---

## 5. Ambient / always-on patterns

### The open-source "Jarvis" reference architecture is Home Assistant

The most mature, genuinely self-hosted, privacy-first always-on stack in 2026 is **Home Assistant Assist** wired over the **Wyoming protocol** (a small socket standard from the Rhasspy project that lets you mix and match wake-word / STT / TTS services). The 2026 local stack is faster-whisper + Piper + a local LLM (Ollama), glued by the Assist pipeline. Hardware ranges from the $59 Voice Preview Edition puck to DIY ESP32-S3 satellites that stream mic audio over Wyoming while a server does the heavy lifting. Wake words run locally via microWakeWord/openWakeWord, with "Okay Nabu," "Hey Jarvis," and "Hey Mycroft" shipping out of the box; HA 2026.3 brought on-device wake-word to Android phones as satellites.

Other worth-studying open-source projects: **OpenVoiceOS** (privacy-first voice platform, Mycroft lineage), **OpenJarvis** (local-first personal agent over Ollama/vLLM/SGLang/llama.cpp), the **GLaDOS** assistants (openWakeWord + Whisper + HA), **EdgeVox** (offline VAD→Whisper→Gemma→Kokoro, sub-second, ships "Hey Jarvis"), and LiveKit's open wake-word training. Wyoming is the key architectural idea worth stealing: **a socket protocol so wake/STT/TTS/brain are independent, swappable network services.**

### Proactive / interrupt etiquette / memory

The 2026 shift is from reactive to **proactive** agents that act without being asked, and the discipline that comes with it:

- **Memory across sessions** is now table stakes for a "personal assistant with memory" — multi-type memory beyond chat history, user control over what's stored, cross-session persistence, and proactive recall (bringing up relevant memory unprompted). spin's substrate §7 faceted memory (`{persona:spin}`, `{room:office}`, global facts) is exactly this shape and is ahead of most.
- **Interrupt etiquette** has a named metric: **"Interruption Regret"** — how often a user dismisses or reacts negatively to an unsolicited intervention; the goal is a low regret rate. The design pattern is preference-aware timing (respect focus blocks, prefer async drafts over live interruptions unless urgent).
- **Idle-time compute** for proactive agents (anticipate and pre-compute during idle windows) is an active research direction (arXiv 2605.25971).

**For spin:** the always-on / multi-room layer is a real path (Wyoming-style satellites + openWakeWord), but it is polish relative to adding the brain. The higher-value borrow right now is the *proactive discipline*: since spin already has durable faceted memory, the missing pieces are an interruption-regret-aware notification policy and idle-time recall, both of which live in the middle layer spin owns.

---

## 6. Synthesis for spin

**What spin is:** a Pipecat-family local cascade (Silero VAD, faster-whisper/Kokoro), a middle layer spin owns (intent routing, persona, memory faceting), and a heavy brain (pg-ai-stewards + Claude seats over loom). Hardware is a dual-4090 box. Constraints: local-first, privacy-conscious, Claude Max + cheap API models.

**The three highest-leverage additions, in order:**

1. **Turn-taking + latency tuning (weeks, cheap, do first).** Add Pipecat `SmartTurnDetection` on Silero VAD; tune to the budget (barge-in <150 ms, turn gap 200–450 ms, p95 <800 ms). Stack: Pipecat SmartTurn, keep Kokoro, optionally swap STT to Parakeet-TDT for streaming latency. This reproduces the speech-to-speech "feel" that Gemini/OpenAI charge cloud dollars for, on hardware you already own. It is the difference between "small model with a mic" and "alive."

2. **Brain over MCP with async offload (the real Jarvis step).** Wire a curated pg-ai-stewards MCP subset into spin's tool list — `consult_subagent`, a `work_item` dispatch, `remember`/`forget` — not the full ~50 (a small front model drowns). Use the async pattern for slow brain calls: reply fast, dig while talking. Stack: `stewards-mcp` as a curated tool set now; the Pipecat multi-worker sidecar (job-over-bus to a substrate worker) when you want true "talk while the brain thinks." This is uniquely spin's because no vendor S2S API can host a Claude brain.

3. **Eyes as a callable tool (later, optional).** Add on-demand vision: capture a screen region or camera frame when invoked, gate on region-of-change, send ≤1 FPS to a cheap VLM. Stack: Gemini Live as a side-channel for fastest time-to-working, or **Vision Agents by Stream** (open source, WebRTC, YOLO pre-filter → LLM, MIT-family, self-hostable) for the fully-owned continuous path. Do the tool-call version before any always-on vision.

**Explicitly not now:** talking-head avatars (skip; do the Vue UI instead), a Go voice-engine rewrite (the roadmap's own low-value trap), and switching to a speech-to-speech API (kills the Claude brain and the privacy posture).

### The tension worth naming

The honest counter-case: for *pure conversational naturalness*, a speech-to-speech API (Gemini Live especially, since it also sees) is genuinely ahead of any cascade today, and it is tempting. spin should not take it — it forfeits the Claude brain, the local privacy posture, and per-minute cost — but spin *should* treat S2S as the bar its cascade is chasing, and steal the specific things that make S2S feel alive: semantic turn detection, the sub-800 ms budget, sub-150 ms barge-in, and streaming partials all the way through. The field's frontier for *feel* is speech-to-speech; spin's frontier for *substance* is the deep Claude brain that S2S can't reach. Build the second, borrow the first's ergonomics.

---

## Sources

Voice — realtime APIs and frameworks:
- OpenAI Realtime API (official developer guide): https://developers.openai.com/api/docs/guides/realtime
- OpenAI, "Introducing gpt-realtime and Realtime API updates" (official; page 403'd on fetch, title/claims from search): https://openai.com/index/introducing-gpt-realtime/
- OpenAI Realtime production/cost (secondary, pricing figures reported): https://tokenmix.ai/blog/openai-realtime-voice-api-2026-cost-latency and https://www.forasoft.com/blog/article/openai-realtime-api-voice-agent-production-guide-2026
- Best speech-to-speech APIs 2026 (Inworld): https://inworld.ai/resources/best-speech-to-speech-apis
- Gemini Live API overview (official): https://ai.google.dev/gemini-api/docs/live-api
- Gemini Live API on Vertex / GA (secondary): https://byteiota.com/gemini-live-api-production-vertex-ai/ and https://www.mindstudio.ai/blog/gemini-3-1-flash-live-screen-sharing-voice-ai
- Claude Code voice dictation (official, confirms STT-only / push-to-talk): https://code.claude.com/docs/en/voice-dictation
- Claude voice features status (secondary): https://www.datastudios.org/post/claude-voice-features-explained-current-status-and-upcoming-real-time-updates
- Claude Code voice mode launch (TechCrunch): https://techcrunch.com/2026/03/03/claude-code-rolls-out-a-voice-mode-capability/
- Pipecat vs LiveKit (Cekura): https://www.cekura.ai/blogs/pipecat-vs-livekit-the-real-difference
- Pipecat vs LiveKit vs TEN (Medium): https://medium.com/@ggarciabernardo/realtime-ai-agents-frameworks-bb466ccb2a09
- Vapi vs Pipecat vs LiveKit (Inworld): https://inworld.ai/resources/vapi-vs-pipecat-vs-livekit

Voice — component models and latency:
- Best local STT 2026 (onResonant): https://www.onresonant.com/resources/local-stt-models-2026
- Best open-source STT 2026 benchmarks (Northflank): https://northflank.com/blog/best-open-source-speech-to-text-stt-model-in-2026-benchmarks
- Moonshine v2 (arXiv): https://arxiv.org/html/2602.12241v1
- Best local TTS 2026 (Local AI Master): https://localaimaster.com/blog/best-local-tts-models
- Best TTS models 2026 (CodeSOTA): https://www.codesota.com/guides/tts-models
- Voice AI barge-in & turn-taking 2026 (FutureAGI): https://futureagi.com/blog/voice-ai-barge-in-turn-taking-2026/
- Voice agent barge-in & turn-taking tuning (CallSphere): https://callsphere.ai/blog/vw7d-voice-agent-barge-in-turn-taking-2026
- Production voice AI latency/architecture (Prodinit): https://prodinit.com/blog/production-voice-ai-agents-latency-architecture

Vision-in:
- Efficient video intelligence 2026 (Vikas Chandra / Meta): https://v-chandra.github.io/efficient-video-intelligence/
- Real-time vision AI architecture (GetStream): https://getstream.io/blog/realtime-vision-ai-architecture/
- Real-time video processing playbook 2026 (Forasoft): https://www.forasoft.com/blog/article/real-time-video-processing-with-ai-best-practices
- Visual AI in video 2026 landscape (Voxel51): https://voxel51.com/blog/visual-ai-in-video-2026-landscape

Computer use / desktop sharing:
- Computer use agents 2026 matrix (Digital Applied): https://www.digitalapplied.com/blog/computer-use-agents-2026-claude-openai-gemini-matrix
- Browser Use vs Operator vs Claude Computer Use (Particula): https://particula.tech/blog/browser-use-vs-operator-vs-claude-computer-use-web-agents
- Microsoft Copilot Studio computer use GA (Microsoft): https://techcommunity.microsoft.com/blog/copilot-studio-blog/computer-using-agents-in-microsoft-copilot-studio-are-now-generally-available/4519427
- UFO2: The Desktop AgentOS (arXiv): https://arxiv.org/pdf/2504.14603
- Computer use & GUI agents 2026 state of the art (Zylos): https://zylos.ai/research/2026-02-08-computer-use-gui-agents/
- Accessibility tree vs pixel architectures (DEV / Runtime Snapshots): https://dev.to/alexey_sokolov_10deecd763/runtime-snapshots-16-the-three-architectures-of-browser-agents-4gkc
- Pipecat + Gemini Live screensharing demo: https://docs.pipecat.ai/pipecat/features/gemini-live
- LiveKit Gemini realtime + live vision recipe: https://docs.livekit.io/recipes/gemini_live_vision/

Video out / avatars:
- HeyGen LiveAvatar: https://help.heygen.com/en/articles/12758516-introducing-liveavatar and https://www.liveavatar.com/
- Best avatar APIs 2026 (VEED): https://www.veed.io/learn/best-avatar-apis
- Open-source 3D talking avatars, real-time lip-sync (Adafruit): https://blog.adafruit.com/2026/05/13/open-source-3d-avatars-that-can-speak-and-lip-sync-in-real-time
- Open-LLM-VTuber (GitHub): https://github.com/Open-LLM-VTuber/Open-LLM-VTuber

Ambient / always-on:
- Home Assistant Voice Preview Edition: https://www.home-assistant.io/voice-pe/
- Wyoming protocol (Home Assistant): https://www.home-assistant.io/integrations/wyoming/
- Home Assistant approach to wake words: https://www.home-assistant.io/voice_control/about_wake_word/
- Self-hosted voice assistant 2026 guide: https://www.kunalganglani.com/blog/self-hosted-voice-assistant-home-assistant-2026-guide
- openWakeWord review: https://www.codeline.co/thoughts/repo-review/2024/openwakeword-open-source-wake-word-detection
- OpenVoiceOS: https://github.com/openvoiceos
- Proactive AI 2026 (AlphaSense): https://www.alpha-sense.com/resources/research-articles/proactive-ai/
- Idle-time compute for proactive agents (arXiv): https://arxiv.org/pdf/2605.25971
- Personal AI assistants with memory 2026 (Vellum): https://www.vellum.ai/blog/best-personal-ai-assistants-with-memory

Open-source vision + voice framework:
- Vision Agents by Stream (announcement, open source): https://getstream.io/blog/vision-agents-by-stream/
- Vision Agents by Stream (GitHub): https://github.com/GetStream/Vision-Agents

*Verification note: primary/official docs were fetched for Gemini Live, OpenAI Realtime (guide), Claude Code voice dictation, Vision Agents by Stream, LiveKit Gemini vision, and the GetStream vision architecture. OpenAI's gpt-realtime announcement page returned 403; its claims are carried from search-result summaries and marked. Gemini Live GA-vs-Preview status differs between the Gemini Developer API docs (Preview) and secondary Vertex reports (GA) — both are noted rather than resolved. Per-minute S2S pricing figures are blog-reported, not official.*
