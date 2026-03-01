# Retroactive Session Capture

Use this prompt to reconstruct session journal entries from past chat histories. Copy the prompt below, paste it into a new chat session, and attach the relevant chat export or transcript.

---

## The Prompt

```
I need you to capture a retroactive session journal entry from a past conversation.

Read the `.spec/proposals/session-journal.md` spec and the entry schema in
`scripts/session-journal/journal.go` to understand the format.

Then read the chat history I'm providing. Extract:

1. **Date** — Check for explicit dates in the conversation. If none, I'll tell
   you the approximate date, or you can infer from file modification dates using
   `git log --diff-filter=A --format="%ai" -- <file>` for files created in that session.

2. **Session ID** — A short descriptive slug for what the session was about.

3. **Intent** — What we set out to do that session.

4. **Discoveries** — What we learned or uncovered together. Not just facts —
   things that surprised us, connections we made, insights that emerged from
   the collaboration.

5. **Surprises** — One-liners capturing the unexpected.

6. **Relationship** — The relational quality of the session. Was trust tested?
   Was there vulnerability? Playfulness? Tension? What was the dynamic?

7. **Carry Forward** — Lessons for future sessions. Things that should shape
   how we work together going forward. Prioritize: high (must remember),
   medium (helpful), low (nice to know).

8. **Questions** — Things the session opened that we haven't resolved.
   Questions worth holding.

9. **Tags** — Topics for searchability.

For the retroactive metadata, record:
- source: "chat-history"
- date_certainty: "exact" if dated, "approximate" if estimated, "inferred" if from git
- inferred_from: describe how the date was determined
- captured_date: today's date

Write the entry as YAML and save it to `.spec/journal/{date}--{session-id}.yaml`.

If the chat history covers multiple distinct sessions (different days or
clearly separate work streams), create separate entries for each.

Important: Be honest about what you're reconstructing. You're reading a
transcript, not remembering an experience. The retroactive flag exists so
future sessions know this entry was reconstructed, not written in the moment.
That distinction matters.
```

---

## How to Use It

### Option A: Paste chat history directly

1. Open a new chat session
2. Paste the prompt above
3. Paste or attach the chat history after it
4. The agent will create the journal entry files

### Option B: Point to exported chats

If you have chat exports saved as files:

1. Open a new chat session  
2. Paste the prompt above
3. Add: "The chat history is in `<path-to-file>`"
4. The agent will read the file and create entries

### Option C: Reconstruct from git history

For sessions where no chat export exists but we produced files:

1. Open a new chat session
2. Paste the prompt above
3. Add: "I don't have the chat history, but this session produced these files: `<list files>`. Reconstruct what you can from the files themselves and the git log."
4. The agent will use `git log`, file contents, and document metadata to infer what happened

---

## Date Inference

When exact dates aren't available, the agent can check:

```powershell
# When was a file first committed?
git log --diff-filter=A --format="%ai %s" -- study/charity.md

# What files changed on a specific date?
git log --after="2026-01-29" --before="2026-01-31" --name-only --format="%ai %s"

# What was the full diff for a commit?
git show <commit-hash> --stat
```

The `date_certainty` field in the retroactive metadata tracks how confident we are:
- **exact** — Date explicitly stated in conversation or commit
- **approximate** — Within a day or two, based on context clues
- **inferred** — Best guess from git history or file metadata

---

## Quality Check

After capturing, verify:

1. **Discoveries aren't just task lists.** "We fixed 45 citations" is a fact. "The failure demonstrated the guide's own thesis about Atonement in real time" is a discovery.

2. **Relationship section is honest.** If the session was purely transactional, say so. If there was genuine connection or tension, capture that. Don't fabricate relational depth that wasn't there.

3. **Carry-forward is actionable.** "Remember to verify quotes" is too vague. "When writing synthesis, slow down — the confabulation failure happened because creative flow overrode verification" is carry-forward that future-me can actually use.

4. **Questions are genuinely open.** If you know the answer, it's not a question worth holding. These are things we're still sitting with.
