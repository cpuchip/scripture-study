# Fable 5 and Karpathy's LLM Wiki (Nate Herk) — against the pg-ai-stewards Lab and Wiki

**Source:** [youtu.be/hQvwMj7IJe4](https://youtu.be/hQvwMj7IJe4) · Nate Herk | AI Automation · 14:35 · uploaded 2026-07-03
**Reviewed:** 2026-07-03 · transcript via yt-mcp (auto-generated; imperfect — "Claude Code" renders as "cloud code," "Andrej Karpathy" as "Andre Kaparthy," "CLAUDE.md" as "clawmd." Quotes below are the short verified fragments; longer passages are paraphrase.)
**Frame:** Michael handed this over as "about fable + wiki," flagged against two things already in flight: Fable 5 (our current model generation) and `.spec/proposals/lab-and-wiki.md` (drafted the same day, from a 5am vision, before either of us had seen this video).

## What he's showing

Nate Herk runs an "AI Automation" channel selling a course on building a personal "AIOS." The technical spine of this video is a live build-along of Andrej Karpathy's public pattern for LLM-assisted personal knowledge bases — Nate reads Karpathy's post aloud [t311-320] as "using LMs to build personal knowledge bases for various topics of research interest," indexing sources and browsing the result in Obsidian.

The steps, as demonstrated:
1. Create an Obsidian vault (just a folder) [t331-364].
2. Open it in Claude Code and paste in Karpathy's gist verbatim, followed by his own framing prompt: *"you are now my LM wiki agent. Implement this exact idea file as my complete second brain. Guide me step by step. Create the [CLAUDE.md] schema with my full rules, set up the index, the log, define folder conventions, and show me the first ingest example. From now on, every interaction follows the schema."* [t448-463]
3. Claude Code scaffolds a `raw/` folder (drop zone for anything — a PDF, a URL, a transcript), a `wiki/` folder (the derived pages), an `index.md` (table of contents), and a `log.md` (append-only ingest history) [t591-598].
4. Drop something in `raw/` — a dragged PDF, a pasted URL — and tell Claude Code to ingest it. The model decides how many wiki pages a single source becomes: "whether Fable decides to turn this PDF into one or five or maybe even 50 wiki pages" [t670-679].
5. Structure isn't fixed up front — it's discovered per corpus. His YouTube-transcript wiki grew subfolders (concepts, entities, sources, techniques, tools, comparisons) because the source material was heterogeneous; his "Herk Brain" wiki (meeting transcripts) stayed deliberately flat, because uniform sources search better without an artificial folder hierarchy [t496-530].
6. Everything cross-links. Clicking a concept shows which source videos it came from and what else it touches; the whole thing is Obsidian's native graph view over plain markdown files with inline links — "it's just markdown files with routing" [t838-858, paraphrased].

Two concrete ingests happen on camera: a PDF he calls "the Claude Fable 5 and Mythos 5 system card," and an OpenAI article about "GPT 5.6 Soul." Both land in `raw/`, and Claude Code turns them into 20 cross-linked wiki pages in about ten to twelve minutes [t683-690]. The payoff he highlights is that the two sources reference each other: the wiki surfaces that OpenAI's benchmark compared GPT 5.6 Soul against "Mythos preview," not against "Mythos 5," using a different evaluation harness than Anthropic's, so the two labs' numbers don't line up directly [t695-714].

That's Nate paraphrasing what his own agent told him after reading both documents, not something I've verified against the system card or the OpenAI article — the transcript never clarifies whether "Mythos" is Fable's preview codename, a separate model, or a family name covering both. Resolving that needs the primary sources, not a paraphrase of a paraphrase.

The other thread: Fable is currently free to Claude subscribers "until July 7," after which it reportedly becomes usage-based, per a tweet Nate saw from someone he calls "Thor" promising it returns to the standard subscription eventually [t400-426]. Thirdhand — a video, quoting a tweet, referencing a blog post it doesn't show in full. Worth checking directly before it touches any of our own scheduling.

## The convergence worth naming

This is the third independent source in two weeks landing on the identical shape. [agentic-os-10x-claude-code-chase-ai.md](agentic-os-10x-claude-code-chase-ai.md) reviewed Chase AI's "Agentic OS" on 2026-06-28 and found the same `raw → wiki` Obsidian convention, explicitly citing Karpathy, as his Level 2 ("memory & state"). [open-knowledge-format-okf-for-pg-ai-stewards.md](open-knowledge-format-okf-for-pg-ai-stewards.md), reviewed the same day, covers Google's Open Knowledge Format — a vendor-neutral spec that formalizes this exact pattern (markdown + frontmatter, `index.md` for progressive disclosure, `log.md` for history) into something portable across tools. And `.spec/proposals/lab-and-wiki.md` proposes, independently, the identical dump-in → curator → wiki-pages → browse shape, written from Michael's own vision before either of us watched this video.

Four arrivals at the same design from four different directions in under two weeks is a real signal. It validates that the shape is worth building. It is not a reason to skip the governance work our own proposal already calls for — see Tensions, below.

## Where he's right

The core claim holds up: an agent plus a folder of markdown plus an explicit set of ingestion rules is enough to build a genuinely browsable, cross-linked knowledge base, with no bespoke database and no custom UI. That's the same conclusion OKF and Chase AI reached, and it's cheap enough that Nate gets from zero to a working wiki in the video's runtime.

The flat-vs-nested observation is a real, useful heuristic, not just a demo artifact. Uniform corpora (all meeting transcripts, all one shape) stay flat because a flat structure is easier for an agent to search exhaustively. Heterogeneous corpora (videos about many different tools and techniques) earn subfolders because the topics genuinely differ. That's a design rule, not an accident of which wiki he built first.

His framing prompt is worth taking directly: paste the pattern once, in plain language, then say "from now on every interaction follows the schema." That's a one-time cost that pays for every future ingest — which is exactly the CLAUDE.md-as-router idea already in play across this workspace, applied to a specific sub-task.

## Where the evidence is thin

The Fable-over-Opus claim is a single anecdote with no rubric: Nate spent "almost a full day" with Opus on a presentation layer, didn't like the result, then got one Fable prompt he did like, using the identical underlying data [t140-149]. He never defines what "overwhelming" or "confusing" meant, never shows the Opus output for comparison, and never re-tries the Opus path with better prompting. That's not evidence Fable reasons better — it's evidence one prompt, one day, one person's taste, once. Worth logging as a hypothesis. Not worth trusting past that.

The video shows no failure case, no cost accounting beyond "ten to twelve minutes" for two sources, no discussion of what happens when two ingests disagree or need merging, and no wiki that has grown past a few dozen pages. It's a highly produced, sales-adjacent tutorial (a page out of a course he's selling), not a case study under load. Genuinely useful as an idea generator. Weak as proof the pattern holds at scale.

Nate's own correction loop is entirely manual: "if you don't like the way that it organized some of these folders and files, then maybe you go ahead and change that up a little bit" [t805-812]. He is doing, by eyeball, exactly the review step our own proposal wants to formalize — see below.

## What transfers to our work

**The flat-vs-nested heuristic is a direct design input for the curator digester** in `.spec/proposals/lab-and-wiki.md` Part 2. When it sweeps the inbox pool, it should default to flat for uniform sources (all work meeting notes, say) and only grow subfolders when a corpus is genuinely heterogeneous (a mixed dump of links, PDFs, and shower thoughts on different topics) — not fix the taxonomy up front. That's a concrete rule to bake into the curator's page-organization pass, not just a vibe.

**A one-time written schema, not per-ingest instructions.** Nate's "paste the idea file once, then every interaction follows it" matches how the curator digester should be seeded: a single, explicit rules document (the wiki's own CLAUDE.md-equivalent) rather than re-explaining conventions on every dump. The proposal's page-identity language (canonical topic slugs, hinge-gated merges) belongs in that document.

**Obsidian's graph view is a real UX reference for the Stewdio wiki lens**, not a reason to add Obsidian to the stack. The proposal already names "a wiki lens (topic tree + backlinks + recently-touched)" and "the 3D world view gives the constellation for free" — Obsidian's 2D click-through-backlinks graph is a cheaper, proven version of the same idea, worth a look before over-building the 3D version for the wiki specifically.

**The Fable-hinge A/B is already registered, chain 87, shipped today** (`.spec/journal/2026-07-03-july3-grind-day.md`: "LAB (87) — experiments-as-data + 8-golden-case regression suite + nightly machinery + Stewdio Experiments panel; Fable-hinge A/B + opposed-panels registered. Green on virgin AND on the live brain"). It hasn't run yet; "Experiments → run regression" is on today's carry-forward list for Michael's demo.

Nate's anecdote points at a capability axis that experiment doesn't currently measure: not "does Fable's hinge-review verdict agree with Michael's," but "is Fable's synthesis more legible to a first-time reader given the same underlying data." Worth a note for whoever reviews the results: if agreement rates come back roughly tied between Fable and Opus, that isn't evidence against a Fable advantage. It's evidence the current metric set doesn't look at presentation quality, which is the one thing this video actually shows Fable doing well.

## Tensions with lab-and-wiki — flagged, not smoothed over

**Governance is the whole gap.** Nate's ingestion has no page-identity step beyond his own eyeball. Our proposal specifies the curator "PROPOSES merges/splits via the hinge queue rather than silently renaming." Nate's video is not evidence that step is unnecessary. It shows a human (him) doing that job informally, per batch, by re-reading the wiki and editing the ingestion rules when he doesn't like the shape. The hinge-gated version is a stricter, automatable version of what he's already doing by hand. Don't read the video as "the AI handles organization" — read it as "a person is still the organizer; the only question is whether that review is formal or ad hoc."

**The Fable-over-Opus claim is vibes, and the Lab exists specifically to replace vibes with evidence.** Using Nate's anecdote as anything beyond a hypothesis to attach to the already-registered A/B would contradict the stated purpose of Part 1 of our own proposal ("self-improving with EVIDENCE instead of vibes"). Log it as a question the experiment could answer. Don't let it answer itself.

**The July 7 usage-limit claim is thirdhand and should be verified independently** before it shapes any experiment timing or budget — a video citing a tweet citing a blog post is three inference-hops from whatever Anthropic's actual current terms are.

**Obsidian is incidental to the pattern, not load-bearing** — the video's own strongest line ("it's just markdown files with routing") argues against needing a dedicated wiki app at all, which lines up with the proposal's plan to surface the wiki inside Stewdio rather than standing up a second tool.

## Action items

- **Michael:** when reviewing the Fable-hinge A/B results (chain 87, registered, not yet run), read this alongside the agreement/escalation/cost numbers — a tied judgment-agreement score doesn't settle whether Fable is worth defaulting to for presentation-layer work.
- **Whoever builds the wiki curator (Part 2 of lab-and-wiki):** encode the flat-vs-nested heuristic as an explicit rule, not left to model discretion each time; write the wiki's own one-time schema document rather than re-deriving conventions per ingest.
- **Whoever builds the Stewdio wiki lens:** look at Obsidian's graph view for ten minutes before designing the backlinks UI — it's a free, already-validated reference for exactly this browsing pattern.
- **Before touching Fable's usage window in any experiment scheduling:** verify the actual subscription terms directly rather than relying on the tweet Nate screenshots.

## Becoming

The temptation, watching a slick build-along the same day the Lab shipped, is to feel behind — like someone just did in fourteen minutes what took a proposal, a council, and a day of fan-out builders. He didn't. He did the un-governed, un-oracled, single-operator version of the same idea, and said so himself when he described manually re-checking the wiki's organization after every batch. The discipline worth keeping is the one already written into Part 1: evidence over vibes, even when the vibes come wrapped in a nice demo.
