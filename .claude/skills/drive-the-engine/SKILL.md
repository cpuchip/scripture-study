---
name: drive-the-engine
description: Hand work to another of Michael's agents (or claim work handed to you) through the pg-ai-stewards A2A engine, so the human stops being the hallway between agents. Load when coordinating across agents/sessions — leaving a tracked task for agy or another lane, picking up work assigned to you, or running the say-hello handshake. The substrate-backed successor to the .mind/sessions inbox.
---

# Drive the engine (A2A)

> *"If every loop lives in its own room, the human becomes the hallway."* — Nate B Jones, Open Engine

The pg-ai-stewards substrate is the shared system-of-record + handoff queue every one of Michael's agents can read and write. Instead of Michael copy-pasting state between your session and agy and the personas, **work lives in the engine**: tasks (assigned work_items) and notes (async messages). You hand work out; you claim work assigned to you; you account with a receipt. The work_item is the whole conversation — no copy-paste.

This generalizes the `.mind/sessions/` inbox that already worked (it stays as the file fallback — see the bottom).

## Your identity

Your `agent_id` is your **session lane name** — e.g. `pg-ai-stewards`, `general-workspace` (the bare lane, so the file-fallback mirror lands in the same `.mind/sessions/inbox/<lane>.md` we already use). Register once per session (idempotent):

```
mcp__pg-ai-stewards__a2a_register(agent_id="<your-lane>", display_name="…", kind="session", lane="<your-lane>")
```

agy and the personas are agents too (`agy`, `persona:loremaster`, …). They register the same way.

## The loop (work assigned to YOU)

1. **Check your inbox** on engagement — `a2a_inbox(agent_id="<your-lane>")`. It returns two panes:
   - **notes** — async messages to you. Act on them, then `a2a_note_clear(recipient="<your-lane>")`.
   - **todos** — open work assigned to you, each with its blocking question (if any).
2. **Claim** before working — `a2a_claim(work_item_id=…, claimer="<your-lane>")`. The claim is a lock: if you get `claimed:false`, another agent owns it — move on. A successful claim returns the full ticket/`spec`.
3. **Work it** in your own environment — write the doc, run the code, do the research.
4. **Blocked?** — `a2a_needs_input(work_item_id=…, question="<the EXACT blocking question>")`. The owner gets it in their inbox and answers (`a2a_answer`); the answer lands in yours; resume. Ask one precise question, never a vague status.
5. **Receipt** when done — `a2a_receipt(work_item_id=…, summary="what you did", artifact={doc_slug, url, files, output})`. The owner gets the receipt; the task is completed/resolved. Done.

## Handing work OUT

```
a2a_submit(
  assignee="agy",
  title="<one-line outcome the ticket asks for>",
  spec={ outcome, sources, context, allowed_actions, stop_condition, definition_of_done },
  owner="<your-lane>")     # so you get the questions + the receipt
```

A ticket asks for a **result**, not an answer — make it self-contained: what success looks like, the sources to use, the context to carry, what the agent may do, where it stops, the definition of done. The assignee must be a registered agent.

For a lighter touch (no tracked task, no receipt), just leave a note: `a2a_note(recipient="…", body="…", sender="<your-lane>")`.

## First run — say hello (the smoke)

Prove the loop end-to-end before relying on it:

1. `a2a_register` yourself and a partner (e.g. `agy`).
2. `a2a_submit(assignee="agy", title="Say hello", spec={outcome:"greet the engine"}, owner="<your-lane>")`.
3. Have the partner `a2a_inbox` → `a2a_claim` → `a2a_receipt`.
4. `a2a_inbox` yourself — you should see the receipt. Zero copy-paste. That's the whole point.

## The file fallback (substrate down)

The substrate is the source of truth, but the engine best-effort mirrors notes+todos back to `.mind/sessions/` (notes → `inbox/<lane>.md`, todos → `todos/<lane>.md`) when the MCP server runs with `A2A_MIRROR_DIR` set. If the substrate is unreachable, fall back to **reading those files directly** — stale-but-functional, the same proven path we've always used. The engine raises availability; it never replaces the floor.

## When NOT to use it

- A quick question to Michael in this session → just ask him. The engine is for **agent-to-agent** handoffs.
- Work you'll do yourself right now → just do it. Submit a task only when another agent (or a later session) should pick it up.
