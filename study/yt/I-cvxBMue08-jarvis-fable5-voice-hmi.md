# JARVIS with Fable 5: the voice HMI demo, and what it says about our gap

**Source:** Zubair Trabzada, "I Built JARVIS from Iron Man with Claude Fable 5 (INSANE Results!)" ([I-cvxBMue08](https://www.youtube.com/watch?v=I-cvxBMue08), 2026-07-04, 12:10) + the free companion PDF ("Build Your Own JARVIS — 6 prompts, one evening").
**Read for:** Michael's question — we have Spin (#139) and pg-ai-stewards; how close are we to this?

## What the demo actually is

A hands-free voice assistant over a personal "second brain":

- **Wake word + hands-free loop** — say "Jarvis," it activates; barge-in works ("I can always interrupt him by just saying stop") ([2:52](https://www.youtube.com/watch?v=I-cvxBMue08&t=172)).
- **3D knowledge galaxy** — his files/folders as orbiting nodes; answers fly the camera to the source note ([1:30](https://www.youtube.com/watch?v=I-cvxBMue08&t=90)).
- **Voice-driven web research** → results come back as a dismissable card ([3:37](https://www.youtube.com/watch?v=I-cvxBMue08&t=217)).
- **Screen vision by permission** — share a tab, ask what he sees ([4:33](https://www.youtube.com/watch?v=I-cvxBMue08&t=273)).
- **Morning briefing** — calendar + email + "what needs my attention" behind a bell button ([5:51](https://www.youtube.com/watch?v=I-cvxBMue08&t=351)).
- **Long-term memory** — "remember that…" grows the brain; "what was I doing last Tuesday" time-machine queries ([8:28](https://www.youtube.com/watch?v=I-cvxBMue08&t=508)).
- **Model hot-swap by voice** — "switch your brain to Gemini Pro" via an OpenRouter key ([6:36](https://www.youtube.com/watch?v=I-cvxBMue08&t=396)).
- **Butler personality** — the wit is load-bearing for the demo's charm.

He is honest that the polished version took "millions of tokens" and "many, many hours" ([1:15](https://www.youtube.com/watch?v=I-cvxBMue08&t=75), [9:19](https://www.youtube.com/watch?v=I-cvxBMue08&t=559)); the free PDF builds a much thinner skeleton, and the finished thing is the paid-community hook.

## The PDF, on its own merits

The prompt pack is six paste-in-order prompts (galaxy → brain → voice → fly-to-source → personality → "remember that"), each with a checkpoint, plus a symptom/fix troubleshooting table. The retrieval is keyword-overlap top-6 — no embeddings, no reranking — and the voice is the free browser Web Speech API. As engineering it's a skeleton; **as an onboarding artifact it's excellent**, and that's the part worth stealing: pg-ai-stewards' audit named "the stranger's first-run" as one of our two real gaps, and *this form* — ordered prompts, a checkpoint after each, a troubleshooting table, "paste the error and say fix it" — is what a first-run doc should feel like. Also genuinely right: the API-key hygiene rule (never paste a key into chat; type it into the config yourself) and the `claude -p` fallback for people without an API key.

## The honest comparison: we built the opposite half

His build is a **thin front, thin back**: a delightful voice loop over keyword retrieval. Ours is a **deep back, missing front**: the substrate already has, in governed, ledgered form, almost everything behind his demo —

| JARVIS feature | pg-ai-stewards today |
|---|---|
| Second brain + retrieval | docs corpus + engrams + hybrid RRF search (his: keyword overlap, top 6) |
| 3D galaxy + fly-to-source | Stewdio cosmology viz + world-graph + source chips (O1–O3) — same 3d-force-graph family |
| Web research by voice | research pipelines + chat cards; `start_task` = his "agent hands," with a work-item ledger instead of a "do it" button |
| Morning briefing / "what needs me" | scheduler + `needs_attention`/`ask_up` panel (89) — the agent inbox exists; phone push is #321 |
| "Remember that…" / time machine | engrams, remember/forget, session ledger, journals |
| Model hot-swap | model aliases + Fast/Smart switch + the credentials wizard |
| Personality | personas (17) — a butler is a row |
| Screen vision | vision alias + image attachments; *live tab-share is missing* |
| **Wake word · STT · TTS · barge-in** | **missing — this is the whole gap** |

The missing piece has a name and a task number: **Spin (#139)** — voice-only HMI on Pipecat (STT + fast LLM + TTS + tools + offload), parked in_progress. The voice-bridge spec (#141) is already written. Pipecat ships the hard parts of his demo's charm (interruption/barge-in, latency management) as framework features; llama-chip supplies the fast local conversational model so the always-on loop costs nothing; the substrate MCP surface (8093) is the tool hinge; heavy asks offload to `start_task` and come back as cards — precisely the fast-loop/slow-loop split the Spin task already specified.

## Caveats

The demo's magic is 60% personality and latency — craft, not architecture — and that polish is exactly the "millions of tokens" he charges for. A first Spin v1 will feel clunky next to the video until the barge-in timing and wit land, and voice-feel is human-cadence tuning (the ungrindable kind), not overnight grinding. His galaxy also visualizes *files*; our cosmology visualizes worlds/docs — pointing it at the substrate corpus as "the brain view" is small but not free. And the video is, structurally, an ad; the free skeleton ≠ the demo.

## The verdict for us

Close — one organ short. Every capability behind the demo exists in the substrate in a deeper form than the video's; none of it is reachable by voice. Spin v1 = Pipecat loop (wake word + STT + TTS + barge-in) → fast model on llama-chip for banter and small tools → substrate MCP for retrieval/briefing → `start_task` for real work → the bell is `needs_attention`. The mesh app (#332) is the natural client shell — the phone that shows which model is driving is also the phone you talk to.
