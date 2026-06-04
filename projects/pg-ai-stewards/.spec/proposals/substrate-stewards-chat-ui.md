# Substrate proposal (STUB — idea capture) — a chattable operator UI on stewards-ui

**Status:** IDEA, not ratified. Captured 2026-06-04 from Michael: *"I don't have a good way of kicking work off like this … we should close that gap and add a good chattable UI to stewards so I can push some of this work too."*

## The gap
Right now, kicking off substrate work (dispatching a `code-pr` work_item, escalating a model, resuming the soak, watching a run) requires the orchestrating agent to run SQL/MCP against the substrate. Michael can't easily do it himself — so he can only push this work *through* an agent session, not directly. That makes the agent a bottleneck on his own substrate.

## The idea
Add a **chat/operator surface to `stewards-ui`** (the existing Go + frontend at `scripts/stewards-ui/`) so Michael can direct the substrate conversationally:
- **Dispatch work** — "build X in repo Y with these acceptance criteria" → creates a `code-pr` work_item (the AX3 flow), kicks the first stage, and streams progress.
- **Watch** — live work_item stage/status, the plan/implement/review critiques, cost.
- **Steer** — resume/pause the soak, escalate a stage's model, approve/close a draft PR (the Hinge stays his).
- Possibly **converse** with the substrate / its personas directly (overlaps ai-chattermax — see below).

## Relation to ai-chattermax
Distinct but adjacent. **ai-chattermax** is a *social* multi-party room (humans + AI personas, D&D). **This** is an *operator console* — directing work, not socializing. They may share the chat transport and the persona concept, but the intents differ (do work vs. be together). Worth deciding whether this is a mode of ai-chattermax or its own stewards-ui surface.

## Open questions (for when this is picked up)
- Console-in-stewards-ui vs. a mode of ai-chattermax vs. both?
- How does a chat instruction become a well-formed `code-pr` work_item (the binding_question + acceptance_criteria are what make the coder succeed — does the UI help Michael write those, or does an agent draft them from his chat)?
- Auth (it's an admin surface — ibeco.me cookie + an allow-list).
- Does it expose the full work_item lifecycle or a curated set of safe actions?

*Deferred. Sits behind the ai-chattermax MVP + the context-tools batch. Captured so the gap is named.*
