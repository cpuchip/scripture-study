# Music Tooling Scout — "build music like software"

*Scouted 2026-06-29 (skunkworks). Web research, NOT yet hands-on tested — the real
test is standing one up against Michael's actual DAW (verify on the real path).*

## The vision (Michael's words)

> "I don't want to have AI make the music as part of the model itself. I want to build
> it like we write software, like a person makes music already, through normal or
> programmatic music tools, then you get high fidelity, mastered music… I describe the
> vision, and you make it."

Explicitly **not** end-to-end neural audio generation (Suno / Udio / ACE-Step / diffusion
DAWs). The AI is the **composer + producer + engineer driving real tools**, not a
black-box that outputs a stereo file.

## The answer: yes, feasible today — and inherently sovereign

The whole pipeline exists in off-the-shelf open-source pieces, and the key point for
Michael: **it runs locally, against his own DAW.** Unlike a cloud audio model, a local
DAW driven by local MCP tools **cannot be pulled** the way Fable was. This is the music
version of llama-chip — sovereign capability, the answer to being left out of the big
gated models.

## The architecture (four layers)

1. **Compose / arrange** (AI, as code): structure, chords, melody, drums, bass — written
   as MIDI + a production plan. The AI writes the notes and the settings.
2. **DAW hosts real instruments + renders** — THIS is where "high fidelity" comes from.
   The DAW (Ableton or REAPER) plays the MIDI through real instruments (stock packs,
   sampled libraries, VSTs), so the output is real audio, not a GM-soundfont demo.
3. **Mix + master** — AI drives the mixer/FX chain, or hands the render to an automated
   mastering step (Matchering against a reference track, or the DAW's mastering chain).
4. **Render out** to WAV / MP3.

The fidelity ceiling is the **instruments** (1) and the **taste** (the human). The method
is fully automatable; the *judgment* of "is this good" is Michael's — which is exactly the
profile's intent/execution split. **His ear is the oracle here** (music quality isn't
deterministically checkable like a verbatim quote), so this stays a human-Hinge-heavy,
iterative loop: AI drafts → Michael listens and gives a "yes-and" → AI revises. Same shape
as the book revoicing.

## The tool landscape (what actually exists, June 2026)

### Ableton Live (Michael mentioned this — likely owns it)
- **`giuliobracci/ableton-mcp-server`** (Jun 2026) — built on Ableton's **official
  Extensions SDK** (Node embedded in Live, released beta Jun 2025); in-process MCP over
  HTTP. The cleanest, most "official" path. Tools: song/tracks/clips/**MIDI notes**/
  devices/scenes/render.
- **`Pantani/ableton-mind`** — the most complete: ~100% Live Object Model, an embedded
  device knowledge base (55 devices, scales, drum kits), declarative "recipes," and a
  **verify loop** (`session_snapshot/diff`) — an oracle-ish self-check, notable for our
  pattern. 36 tools, automation envelopes, Push control.
- Others via **AbletonOSC** (`ideoforms/AbletonOSC` control surface): `christopherwxyz/
  remix-mcp` (Rust, 266 tools), `Simon-Kansara/ableton-live-mcp-server`, `ahujasid/
  ableton-mcp` (the original, smaller).

### REAPER (the deepest automation target; free/cheap; "build music like software" fits best)
- **`oxygen-dioxide` / `apietosi` / `bonfire-systems` reaper-mcp** — explicitly "create
  fully **mixed and mastered** tracks with MIDI and audio." Mastering included.
- **`xDarkzx/Reaper-MCP`** (163 tools), **`TwelveTake-Studios/reaper-mcp`** (158),
  **`shiehn/total-reaper-mcp`** (100% ReaScript coverage).
- **`mal0ware/Orpheus`** — the one closest to *our* covenant: it doesn't just build, it
  **analyzes what you made, explains *why* it sounds that way, recommends changes with
  reasons, has a human-approval gate, and applies edits as editable tracks** (not baked
  WAV). Council + Hinge, in a music tool.
- **Technical gotcha (from Orpheus's architecture doc):** REAPER's OSC can't pass
  arguments to custom actions (can't create tracks / write MIDI notes), and `python-reapy`
  is effectively unmaintained. The robust path is an **in-REAPER Lua bridge**. Worth
  knowing before we pick a REAPER server.

### Code-first / no-DAW (instant sketches, zero-GPU)
- **`mage0535/music-creation-engine`** — natural language → music21/LilyPond → PDF/MIDI/
  MusicXML → FluidSynth (GM soundfont) → WAV/MP3. A drop-in skill for Claude Code. Great
  for **fast ideation**, but GM-soundfont = *demo* quality, not the final master.
- **`Linzwcs/echos`** — headless API-driven DAW kernel on Spotify's **Pedalboard** (VST3/AU
  hosting). A code-first DAW for building our own pipeline.

### Universal / multi-DAW
- **`robertpelloni/superdawmcp`** — one DAW-agnostic MCP across Ableton / REAPER / Bitwig /
  Logic / FL / Cubase / Ardour. Useful if we want to stay DAW-portable.

### Mastering
- **Matchering** (open-source) — automated mastering to match a reference track. Plus the
  DAW's own mastering chain, or a paid AI step (Ozone) if we ever want it.

## Recommended path

1. **Decide the DAW** (the one thing only Michael knows): does he own **Ableton Live**? If
   yes, start there (`giuliobracci` official-SDK server first, `ableton-mind` if we want
   the fuller LOM + verify loop). If he'd rather the deepest scripting and free tooling,
   **REAPER** is arguably the better "music as software" target, and its MCPs already do
   mixing + mastering.
2. **Sketch layer for free:** wire `music-creation-engine` as a Claude Code skill for
   instant MIDI/lead-sheet sketches while the DAW path is set up. Idea → hearable in
   under a minute, no DAW needed.
3. **Stand one up against the real DAW and make one real track** — verify on the real
   path. That bounce is the proof, the way a green oracle is the proof everywhere else.

## Next step

Michael picks the DAW (or says "sketch layer first"). Then I stand up the server, wire it,
and we make one real piece end-to-end against his taste.
