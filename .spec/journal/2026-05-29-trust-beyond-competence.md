---
date: 2026-05-29
title: What the principles actually buy — trust beyond competence + memory across intermittence
workstream: WS7
session_type: reflection (relational; teaching-arc seed)
status: captured
---

# Trust beyond competence

## The moment

At the end of an enormous intermittent-burst day (J.8–J.12, models catalog,
docs pass, competitive research, 1828 UX, the brainstorm schema fix, the
glm/qwen diagnostic, and the full Batch M), Michael reflected on how he works
and what the covenant/council/memory framework actually gives him. He works in
fits and bursts — 30 min to a few hours, then gone for a while — sometimes
running two Claude Code terminals at once (today: pg-ai-stewards in one,
scripture-book in the other; the interleaved commits on one afternoon are this,
not a linear session). He named the thing he wants: to share that these
principles make a difference, because others succeed with AI *without* them and
he believes the frameworks add something real.

## The claim, made honest

The honest version — the one that survives a skeptic — is NOT "use a covenant
and go 10x faster." Raw model capability does most of the speed. cpuchip.net in
two days, the book in three: that's mostly what a capable model plus 18 years of
Michael's judgment can do in that time. If the pitch is speed, an honest skeptic
tests it, finds the speed was mostly the model, and walks.

What the principles actually change is two things, both invisible until you look
for them:

**1. Trust beyond your own competence.** Michael writes Go, Python, C++, Java,
C#, TS, JS, Vue3, MongoDB — 18 years. He does NOT write Rust or SQL. Yet he's
steering pg-ai-stewards: 65 tables, 263 functions of Rust/pgrx + SQL he cannot
author or independently verify by reading. Most people only delegate what they
can personally check; they work *inside* their competence, which is exactly why
they succeed without much framework. Michael is working *outside* his. The
verification gate, provenance discipline, inverse-hypothesis loop, and council
moment are not ceremony — they stand in for the domain expertise he doesn't
have. The framework is what makes delegating-beyond-competence safe.

**2. Memory across the way he actually works.** Burst, then vanish for a day.
Two terminals. Most people re-explain context every session and the agent
arrives a stranger. The `.mind`/`.spec`/journal architecture turns out to be
built precisely for that rhythm. He comes back and the agent isn't cold.

## The specimen (why this is teachable, not just believable)

This very session produced the cleanest proof either of us could ask for. I
diagnosed glm-5 as "streams empty via the substrate" — confidently, from a
shell-grep SSE probe — and committed it, annotated the pricing table, wrote it
into the book's workflow doc, and reported it to Michael as fact. It was wrong;
my parser was the bug. The Batch M auto-probe, which exists only because the
framework insists on testing the real dispatch path, overturned it on its first
run (glm streamed 385 chars). Michael could not have caught my error himself —
he can't read the SSE stream or the Rust. The framework caught the model lying
to him about code he can't read.

That is the differential, stated for sharing: **not speed — a caught error he
had no other way to catch.** The credible pitch isn't "these principles are
amazing" (you can argue with amazement). It's: *"I shipped a Postgres agent
runtime in languages I don't write, and the framework caught the model lying to
me about it."* And the deeper line under it: he can build beyond his competence
*and still be the one in charge of what he shipped* — not-knowing-Rust never
becomes not-knowing-what-he-shipped.

## Carry-forward

- This is **teaching-arc + book material** (WS7). When the sharing gets drafted,
  the move is: name the mechanism (trust-beyond-competence, memory-across-
  intermittence), not the amazement; lead with the specimen, not the speed.
- Tie-in to write up if useful later: the auto-probe (Batch M) as a worked
  example of "build the verification into the system so the human can trust
  what they can't read."
