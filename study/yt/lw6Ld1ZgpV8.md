# I Tried /teach and 10x'd My Ability To Learn | Kacper Rutkiewicz | AI Made Simple


## Thesis

Kacper Rutkiewicz argues that the `/teach` skill (created by Matt Pocock) transforms Claude Code from a stateless Q&A tool into a stateful personal tutor that remembers your progress, calibrates difficulty to your level, and ultimately aims to make you independent of AI. The core mechanism is persistence: lesson records, glossaries, and a mission statement accumulate across sessions so the agent knows where you are and picks the next lesson accordingly.

## How it builds

Rutkiewicz structures the video around three moves:

1. **The problem with stateless AI learning** — He opens by contrasting typical AI interactions (one-off answers, no memory, context-window decay) with what a real teacher does. The substitute-teacher analogy frames the entire argument.
2. **The `/teach` skill as a stateful alternative** — He demonstrates the skill's architecture: persistent lesson records, a glossary, a mission statement, and a three-level teaching philosophy (knowledge → skills → wisdom). He shows how it vets sources, builds practice exercises, and points toward communities.
3. **The zone of proximal development** — The conceptual core: every lesson is pitched at a difficulty calibrated to the learner's current level — not too easy, not too hard — based on the tracking data the skill accumulates.
4. **Installation walkthrough** — A practical demo showing how to install the skill from the `/skills` repo into Claude Code, configure a mission statement, and begin a learning session on building LLMs.
5. **The independence goal** — He closes by emphasizing that the skill's endgame is making you *not* need it anymore — passing you to human communities rather than keeping you dependent on the model.

## Key passages

**"Instead of having a substitute who forgets you every single time, you have a teacher that remembers you and knows exactly where you're at."**
— The substitute-teacher analogy that frames the entire video. Statelessness is the problem; statefulness is the solution.

**"There's knowledge, skills, and wisdom."**
— The three-level teaching philosophy. Knowledge comes from vetted sources; skills are built through active practice; wisdom means pointing learners toward human communities.

**"You can outsource your research and all that stuff to AI, but you can never really outsource your understanding."**
— A repeated refrain (appears twice in the video). AI can gather information, but genuine understanding requires human engagement and community.

**"Every lesson that this skill builds is pitched right at your level. So, they're not so easy that you're bored, and they're not so hard that you want to quit."**
— Describes the zone of proximal development in practice: calibration based on accumulated learner data.

**"It's not trying to keep you dependent on the model to keep learning. It's trying to make you good enough at the topic so that you don't need it anymore."**
— The skill's endgame is learner independence, not perpetual AI reliance.

## Themes

- **Statefulness as the differentiator** — The video's central contrast: stateless tools forget you; stateful tools teach you. Persistence (lesson records, glossaries, progress tracking) is what turns an AI from an encyclopedia into a tutor.
- **The zone of proximal development** — Borrowed from educational psychology (Vygotsky), this is the sweet spot where material is challenging but not overwhelming. The skill implements it by calibrating each lesson based on prior interactions.
- **AI as a bridge, not a destination** — The skill is designed to eventually step aside and pass you to human communities. The goal is independence, not dependency.
- **Active learning over passive consumption** — Lessons emphasize taking action and self-checks rather than just reading. Knowledge without practice is incomplete.
- **Trust in sources matters** — The skill vets information from high-trusted sources before embedding it into lesson plans, rather than pulling from "any random sources."

## Tensions & objections

**The strongest objection: the "10x" claim is unverified and likely hyperbolic.** Rutkiewicz titles the video "10x'd My Ability To Learn" but provides no evidence — no before/after metrics, no comparison to other learning methods, no controlled test. The video is essentially a positive demo of a tool he likes, not a measured claim about learning acceleration.

**Statefulness is a feature of Claude Code generally, not unique to `/teach`.** Any Claude Code project with persistent files (notes, progress trackers, custom prompts) can achieve similar stateful tutoring. The `/teach` skill packages this nicely, but the underlying mechanism — file-based persistence — is not novel.

**The zone of proximal development is hard to calibrate without a human teacher.** The skill infers difficulty from your answers to its questions, but an AI has no true understanding of whether you *actually* grasped a concept or just guessed correctly. Mis-calibration could lead to lessons that are too easy (boredom) or too hard (frustration) — the very outcomes the video claims to avoid.

**The "independence" goal may be aspirational rather than achieved.** Rutkiewicz says the skill passes you to communities, but community engagement requires motivation the AI cannot provide. Many learners drop off after the AI stops guiding them — the skill cannot solve the motivation problem.

**Token cost and effort level matter.** The video recommends running on "high" or "max" effort for better research, which significantly increases cost. This is not a free or cheap learning tool — the quality depends on spending more tokens.

## What's worth learning

1. **Use file-based persistence for stateful AI tutoring.** Even without `/teach`, you can create a learning workspace with progress notes, a glossary, and a mission statement. The pattern — persistent files that the AI reads each session — is the real innovation, not the specific skill.
2. **Learn the zone of proximal development.** This educational concept (material pitched just above your current ability) applies to any learning, not just AI-assisted. When choosing resources, aim for "challenging but doable" — not beginner tutorials you've outgrown and not expert texts that overwhelm.
3. **Vet your sources before building lesson plans.** The skill's practice of pulling from high-trusted sources rather than random internet content is a good habit for any self-directed learner. Curate your inputs.
4. **Plan for the handoff to community.** If you use AI to learn something, identify a human community (forum, Discord, local group) to join *before* you start. The AI is a bridge, not the destination.
5. **Active practice beats passive reading.** Structure your learning sessions around doing — building, answering, self-checking — not just consuming explanations. The `/teach` skill's emphasis on "taking action" is sound pedagogy.