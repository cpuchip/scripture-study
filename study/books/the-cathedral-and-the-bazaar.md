# The Cathedral and the Bazaar

*Eric S. Raymond*

---

## The core argument / thesis

Raymond argues that the “bazaar” model of open-source development—releasing early and often, treating users as co-developers, and harnessing mass parallel debugging—routinely outperforms the centralized “cathedral” model, because “given enough eyeballs, all bugs are shallow” (Linus’s Law). Yet he insists this is not automatic magic: it requires a runnable “plausible promise” to seed the community, a coordinator with strong social skills rather than exceptional design genius, and can still fail if execution is poor, as the early Mozilla struggles demonstrated. He further qualifies that one cannot code from the ground up in bazaar style; the model can test, debug, and improve, but it requires an initial runnable seed and a strong basic design. And he explicitly limits the bazaar’s creative scope: insight comes from individuals, not committees or crowds; the social machinery can only catch, refine, and test lightning, not manufacture it.

## Structure

The essay builds its case by layering personal experiment atop software-engineering theory. Raymond opens with Linux as a phenomenon that overturned his assumption that high complexity requires centralized, a-priori control. He then narrates the fetchmail project as a deliberate test of bazaar tactics, extracting numbered aphorisms from each stage (itch-scratching, reuse, early release, user co-development). The middle sections generalize these observations into theory: Linus’s Law, the parallelizability of debugging, and a detailed micro-level analysis of *why* many eyeballs tame complexity (source-code awareness, semi-random trace-path sampling, and rapid release propagation). He then examines the social preconditions—leadership without coercion, the reputation economy (“egoboo”), egoless programming, and Kropotkin’s “principle of common understanding”—before critiquing conventional management as a costly Maginot Line defending against boredom and poor motivation. The epilog tests the theory against the Netscape/Mozilla case, showing both the opportunity and the hard limits of the bazaar model (Mozilla’s initial failure to ship a runnable build, its dependence on proprietary Motif, and Jamie Zawinski’s warning that “open source is not magic pixie dust”).

## Key passages

> “Every good work of software starts by scratching a developer’s personal itch.”  
> *The origin of useful software is personal necessity, not market requirements or assigned tasks.*

> “Given a large enough beta-tester and co-developer base, almost every problem will be characterized quickly and the fix obvious to someone.”  
> *The formal statement of Linus’s Law: debugging depth collapses when enough observers examine the code.*

> “Treating your users as co-developers is your least-hassle route to rapid code improvement and effective debugging.”  
> *The practical mechanism by which the bazaar converts user feedback into developer leverage.*

> “Smart data structures and dumb code works a lot better than the other way around.”  
> *A design principle favoring clarity in representation over cleverness in control flow.*

> “The next best thing to having good ideas is recognizing good ideas from your users. Sometimes the latter is better.”  
> *The coordinator’s chief creative task is curation and integration, not sole invention.*

> “Provided the development coordinator has a communications medium at least as good as the Internet, and knows how to lead without coercion, many heads are inevitably better than one.”  
> *Raymond’s direct counter-proposal to Brooks’s Law for large-scale development.*

> “One cannot code from the ground up in bazaar style. One can test, debug and improve in bazaar style, but it would be very hard to originate a project in bazaar mode.”  
> *The bazaar is an evolutionary optimizer, not a creation-ex-nihilo engine; it requires a runnable seed.*

> “Insight comes from individuals. The most their surrounding social machinery can ever hope to do is be responsive to breakthrough insights—to nourish and reward and rigorously test them instead of squashing them.”  
> *A hard limit: groups refine insight, they do not generate it. The bazaar amplifies individual vision; it does not substitute for it.*

## Themes

- **Cathedral vs. Bazaar:** Centralized, closed planning and long release cycles versus open, evolutionary development with rapid iteration.
- **Linus’s Law and parallel debugging:** Bugs are shallow phenomena when exposed to a large, self-selected pool of observers; debugging is parallelizable in a way that coding is not.
- **Users as co-developers:** The boundary between consumer and producer blurs, especially because in the Unix/Linux tradition many users are hackers too; properly cultivated, users diagnose problems and suggest fixes at a scale no single team can match.
- **Brooks’s Law and its limits:** The principle behind Brooks’s Law is not repealed, but under conditions of cheap communication, self-selection, and egoless programming, its quadratic coordination costs are swamped by the parallelizability of debugging (Linus’s Law) and sub-quadratic duplication costs (Hasler’s Law).
- **Egoless programming and reputation:** Contributors are driven by “egoboo” (reputation among peers) rather than salary; the culture rewards transparency and punishes territoriality.
- **Release early and often:** Rapid iteration propagates fixes faster than it propagates embarrassment, and keeps contributors stimulated.
- **Leadership without coercion:** The coordinator need not be an original design genius, but must be able to recognize good design ideas from others, integrate them, and attract volunteers through charm rather than command.
- **Innovation originates in individuals, not structures:** Cathedrals, bazaars, and committees can catch and refine lightning, but they cannot make it on demand. The bazaar’s advantage is rapid amplification of a good idea, not collective invention.
- **Joy and anti-deadline scheduling:** Enjoyment predicts efficiency; “wakemeupwhenit’sdone” scheduling avoids the quality collapse that immutable feature lists and fixed drop-dead dates produce.
- **The limits of open source:** Mozilla demonstrated that open-sourcing will not save a project suffering from ill-defined goals, spaghetti code, or lack of a runnable build. Open source is “not magic pixie dust.”

## Tensions & objections

The strongest null-case objection is internal to Raymond’s own evidence: the bazaar model works only under a highly specific set of preconditions that his successes share and his one major failure lacked. Linux, fetchmail, and EGCS were all (a) infrastructure used by technically sophisticated users who could read source code, (b) seeded with a strong initial design by a single architect rather than originated by committee, and (c) modular enough to allow parallel debugging without massive coordination overhead. Mozilla, by contrast, failed to attract a massive external community precisely because it violated these preconditions: it did not ship a runnable build for more than a year, required a proprietary Motif library, and suffered from ill-defined goals and spaghetti code. Raymond’s own text therefore contains the seeds of the null case: the bazaar is an excellent *evolutionary* optimizer but a poor *creator*; it cannot originate projects from zero; and its “many eyeballs” are only shallow because they are self-selected from a population already equipped to understand source code. The generalization from kernel hackers and mail-fetcher users to “probably every kind of creative or professional work” is unsupported. For consumer-facing applications, safety-critical systems, ground-up architectural innovation, or “boring” domains without hacker cachet, the bazaar offers no automatic advantage—and may collapse into the incoherence that Mozilla exhibited. The “Delphi effect” assumes equally expert observers; when users are not experts, their aggregated opinion is noise, not signal. In short, the bazaar does not abolish the cathedral; it parasitizes it, requiring a cathedral-built spire before the crowd can raise the roof.

## What's worth learning — and what we could do with it

1. **Ship a “plausible promise” before opening the floor.** For any new tool, document, or system, produce the smallest *runnable* artifact that demonstrates the concept before inviting contributors. Do not open-source an empty repository, a broken build, or a vague specification and expect the crowd to fill it in; the bazaar optimizes, it does not originate.

2. **Publish internals, not just polish.** Treat readers/users as co-developers by exposing raw source material, decision logs, schemas, and version history alongside finished outputs. “Smart data structures and dumb code” applies to documents too: clear structure and visible state lower the barrier for others to spot errors and propose improvements.

3. **When stuck, expose the problem to 3–5 competent observers early.** Apply Linus’s Law to your own knowledge work. Instead of polishing a private solution until it feels ready, characterize the problem in public (or to a small, diverse panel). The fix is often obvious to someone else once the problem is well-described; debugging is parallelizable, but solo rumination is not.

4. **Lead by curation and rapid integration, not by genius.** If coordinating a collaborative project, optimize for charm, responsiveness, and taste rather than being the sole source of good ideas. Release small integrations frequently to maintain momentum; the coordinator’s job is to catch lightning from others and ground it quickly, not to generate every bolt.

5. **Respect the hard limit: seed with individual insight, then amplify.** Do not expect a group to invent a breakthrough architecture or original thesis. Produce a strong initial design or argument alone (or with a tiny cell), then use open collaboration to test, refine, and extend it. The bazaar raises the roof; it does not pour the foundation.

6. **Audit your domain for bazaar preconditions before adopting open collaboration.** Ask: Are the potential contributors self-selected experts? Is the work modular enough for parallel debugging? Is there a runnable seed? If the answer to any is no, the cathedral model—or at least a hybrid with strong central design—may be the honest choice. Open source is not magic pixie dust.
