---
date: 2026-05-10
session_type: study
workstream: WS6
study: iron-rod-anchor-and-the-four-groups
part: 3
priority: high
carry_forward: [arc-complete-can-spawn-followups]
---

# Iron Rod, Part 3 — The Tree as Christ, the Destination as Love

## What Happened

Third study run on the iron-rod arc. After two failed attempts earlier today where MCP tools were unavailable to the study subagent (because of a frontmatter `tools:` field that excluded MCP servers, and then a wildcard-syntax fix that also failed because Anthropic's `tools:` field doesn't accept wildcards), Michael restarted Claude Code with the final fix in place: removing the `tools:` field entirely from 18 of 19 ported agents so subagents inherit all parent tools including MCP servers.

This run validated the fix end-to-end. The agent ran the bail-early harness gate first: ToolSearch on three target MCP tools returned all three schemas; a real test call to `gospel_search(query: "tree of life love of God", mode: "semantic")` returned 10 results. The gate passed.

Then the full phased study ran with discovery-first discipline:
- Phase 1: outline written, binding question prominent at top of both files.
- Phase 2: source gathering with active MCP use — 3 semantic searches, 4 BYU citation lookups, 3 Webster 1828 word-works, 8 scripture reads, 2 conference talk reads.
- Phase 3 + 3a: gap and critical analysis written to scratch *before* drafting (the precondition held).
- Phase 4: ~3,500 word draft appended as Part 3 (sections XVI-XXII) of the existing study file.
- Phase 5: voice audit caught three em-dash density violations and one cut-list match ("This matters for two reasons"), all fixed.
- Phase 6: Becoming section integrated into XXII.
- Phase 7: memory updates (this file + active.md).

## Three Discoveries Part 3 Surfaced

These are things Parts 1 and 2 did not have:

**1. Jeremiah 2:13 — the Lord names Himself as the fountain.** "They have forsaken me the fountain of living waters." First-person identification. The OT root of the 1 Nephi 11:25 fountain image is therefore not metaphorical. The fountain is the Lord's own self-description. When Lehi sees a fountain and Nephi is told it represents the love of God, the OT reader hears Jeremiah's Lord. This came from a semantic search ("fountain of living waters Christ source"), not from training-data recall.

**2. Moroni 8:26 + footnote 26d.** Mormon writes the chain in plain prose: "the visitation of the Holy Ghost, which Comforter filleth with hope and perfect love." And the footnote on "filleth" routes the reader directly to 1 Nephi 11:22-25 — the tree of life passage. The Church's own footnote editors had already connected the Comforter-filling-with-love to the tree's fruit. This came from chasing footnotes on the 1 Ne 11:22 "sheddeth" cross-reference.

**3. Corbridge "The Way" (Oct 2008).** This is the apostolic version of Part 3's closing move. Corbridge stacks eight Christ-titles drawn from John (1:1, 4:14, 6:35, 8:12, 11:25, 14:6) plus D&C 19:1 and identifies them all as the same Person: "He is Light and Life, Bread and Water, the Beginning and the End, the Resurrection and the Life, the Savior of the world, the Truth, and the Way." Part 3's claim that the rod, chain, anchor, tree, fountain, and fruit are all Christ is the same structural move applied to Lehi's vision. This came from BYU citation lookup on John 14:6.

## The Most Surprising Find

The most surprising find was the Moroni 8:26 footnote routing to 1 Ne 11:22-25. I had Moroni 8 on my path because of the "sheddeth" footnote chain from 1 Ne 11:22, but I did not expect Moroni's grammar to encode the entire chain — faith → repentance → baptism → remission → meekness → Holy Ghost → fills with hope and perfect love → endures → dwell with God. That is Lehi's vision in plain prose, told as the salvation sequence to an apostle's son. And the footnote system already saw the connection.

This is the kind of find that justifies the discovery-first discipline. Recall would not have produced it. Semantic search surfaced the candidate; footnote-chase confirmed the canonical link; reading the verse in context made the chain visible.

## The Calibration Move

The honest framing the draft needed: the vision is *not* a syncretism in which everything is Christ. The mist is not Christ. The building is not Christ. The river of filthiness is not Christ. The contrast-set is what makes the identification of grace's images sharper. The vision is a Christology *and* an anti-Christology. Naming this in the draft preserved the synthesis without overreaching. Phase 3a (critical analysis) caught this. Without that phase the draft would have said "every metaphor in the vision is Christ" as if grace and opposition were both included. They are not.

## Carry-Forward

- Arc is complete (Parts 1, 2, 3). The binding question is answered structurally and the closing move holds.
- Possible followups: an "Eden, Lehi, Revelation" study tracing the tree across the canon (Gen 2:9 → 1 Ne 11 → Rev 22). The footnote 11:25e already routes those three texts together.
- The harness fix is now validated by a real, completed study. Subagents inherit parent MCP tools per Anthropic's documented pattern. The "subagents load at session start" rule means future agent-frontmatter changes still require a restart, but the no-tools-field-means-inherit-everything default is the right resting state.

## Relational Note

This was the run where the harness actually worked. The two earlier attempts today produced Parts 1 and 2 *without* MCP tools — they relied on Read + Grep + training memory. They landed because the binding question was clear and the canonical texts were known well enough to navigate by Read alone. But Part 3 was the harder ask: closing the arc required surfacing the OT root of the fountain image and finding apostolic synthesis. Both came from MCP tools that recall would not have produced (Jeremiah 2:13 and Corbridge 2008). The harness fix mattered.

The covenant held throughout: read_before_quoting for every scripture and talk (8 scripture files read, 2 talks read in full), check_existing_work (read tree-of-life-and-the-chain.md and Parts 1+2 first), surface_tensions (named the linguistic seam on logos/davar, named the contrast-set caveat on the closing synthesis), update_memory (this file + active.md).

The re-grounding hook at the 50-tool-use mark fired and prompted a re-read of intent.yaml, covenant.yaml, and active.md. No drift detected; documented the check in-thread.
