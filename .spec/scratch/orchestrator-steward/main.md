# Orchestrator Steward — Research & Findings

**Binding problem:** The brain pipeline currently fires a single agent call per stage and either succeeds or fails. When execution fails (timeout, error, partial work), the human must manually diagnose, fix, and retry. Michael has demonstrated that the real power of AI delegation is the *steward loop* — watching, diagnosing, fixing, restarting, repeating until done. The question: can we build this loop into brain itself?

**Date started:** 2026-04-10
**Related:** [brain-pipeline-fixes-phase4.md](../../.spec/proposals/brain-pipeline-fixes-phase4.md), [brain-pipeline-evolution.md](../../.spec/proposals/brain-pipeline-evolution.md)

---

## Research Areas
1. How Squad (bradygaster) does coordinator/orchestration
2. How other agentic systems do orchestration (OpenAI Swarm, LangGraph, CrewAI, etc.)
3. Scriptural steward/shepherd/watchman patterns — what does faithful stewardship look like?
4. How these patterns already manifest in what we've built
5. How they point to Christ

---

## Findings

### Squad Coordinator Pattern

Source: `external_context/squad/` — Brady Gaster's agent orchestration system (TypeScript)

**Architecture:** Coordinator receives messages → Route Analysis stage determines strategy (direct response, single spawn, multi-spawn, fallback) → Fan-out to agents in parallel via `Promise.allSettled` → Collect results.

**Key design decisions:**
- `AgentSpawnConfig` = { agentName, task, priority, context, modelOverride } — each spawn is a self-contained work unit
- `SessionPool` tracks agent sessions through lifecycle: creating → active → idle → error → destroyed
- `EventBus` publishes events (session.*, coordinator.routing) for observability
- `CostTracker` maintains per-agent and per-session cost breakdowns
- Failure isolation: each spawn wrapped in try/catch, `Promise.allSettled` means one failure doesn't kill others
- Fallback strategy: if all spawns fail, coordinator can try alternative routing

**What's transferable to brain:**
- The spawn-as-work-unit concept maps well to pipeline stages
- Session lifecycle tracking maps to entry state tracking
- Cost tracking already exists in brain (model pricing)
- Failure isolation is critical — one stage failure shouldn't corrupt entry state

**What's different for brain:**
- Squad is conversational (chat routing); brain is pipeline (sequential stages)
- Brain has a predefined stage sequence; Squad routes dynamically
- Brain needs *persistence* across sessions (SQLite); Squad is session-scoped
- Brain's orchestrator needs to resume across process restarts

---

### Industry Orchestrator Patterns

Source: Exa search — "AI agent orchestrator pattern retry loop failure handling escalation design architecture 2024 2025"

**Resilience patterns:**
1. **Exponential backoff with jitter** — don't retry immediately; increase delay between retries to avoid thundering herd
2. **Circuit breakers** (CLOSED → OPEN → HALF-OPEN) — after N failures, stop trying for a cooldown period, then probe with a single attempt
3. **Multi-provider fallback chains** — if primary model fails, try secondary, then tertiary
4. **Checkpointing/resume** — save intermediate state so long-running work can resume after failure
5. **Dead letter queues** — entries that fail repeatedly get quarantined for human review
6. **Defense-in-depth layering** — Agent → Validation → Observability → Circuit Breaker → Fallback → Retry → LLM Provider

**Key insight:** Production systems layer these patterns. It's not "retry OR circuit breaker" — it's retry *inside* circuit breaker *inside* fallback *inside* observability. The orchestrator is the composition of these layers.

**What's transferable to brain:**
- Backoff between retries (don't burn tokens on immediate retry of the same failure)
- Circuit breaker on specific stages (if research fails 3x, stop trying research and escalate)
- Checkpointing naturally exists — each pipeline stage writes its output to the entry, so stage completion IS the checkpoint
- Dead letter / quarantine = failure_count threshold already in brain (currently 3)

---

### Scriptural Steward/Shepherd/Watchman Patterns

#### 1. The Nobleman's Parable — Watchmen on the Tower (D&C 101:43-62)

The Lord gives a parable about servants who planted olive trees, built a hedge, set watchmen — but then *questioned the need for the tower* ("What need hath my lord of this tower?" v47-48). They became slothful and didn't finish it. The enemy came by night.

Key verses:
- v45: Purpose of the tower — "that one may overlook the land round about, to be a watchman upon the tower, that mine olive trees may not be broken down when the enemy shall come"
- v53: "Ought ye not to have done even as I commanded you... built the tower also, and set a watchman upon the tower, and watched for my vineyard, and not have fallen asleep?"
- v54: "The watchman upon the tower would have seen the enemy while he was yet afar off; and then ye could have made ready"
- v61: The faithful servant = "a faithful and wise steward in the midst of mine house, a ruler in my kingdom"

**Pattern: Observability is not optional.** The servants thought monitoring was unnecessary in peacetime. They were wrong. The tower gives foresight — seeing problems before they arrive.

#### 2. The Good Shepherd (John 10:1-18)

- v3: "He calleth his own sheep by name, and leadeth them out"
- v4: "He goeth before them, and the sheep follow him: for they know his voice"
- v11: "The good shepherd giveth his life for the sheep"
- v12-13: The hireling "seeth the wolf coming, and leaveth the sheep, and fleeth" — "careth not for the sheep"
- v14: "I am the good shepherd, and know my sheep, and am known of mine"
- v27-28: "My sheep hear my voice, and I know them... neither shall any man pluck them out of my hand"

**Pattern: The shepherd is invested in the outcome, not just the process.** Knows each one by name. Goes before them (doesn't just send them). Stays when danger comes. The hireling is the anti-pattern: present when convenient, absent when it costs something.

#### 3. Shepherds of Israel — Failed Stewardship (Ezekiel 34)

God indicts the shepherds of Israel for feeding themselves instead of the flock:
- v4: "The diseased have ye not strengthened, neither have ye healed that which was sick, neither have ye bound up that which was broken, neither have ye brought again that which was driven away, neither have ye sought that which was lost"
- v11: "I, even I, will both search my sheep, and seek them out"
- v16: "I will seek that which was lost, and bring again that which was driven away, and will bind up that which was broken, and will strengthen that which was sick"
- v23: "I will set up one shepherd... even my servant David" (Christ)
- v25: "I will make with them a covenant of peace"

**Pattern: The checklist of faithful stewardship is in v4/v16.** Strengthen diseased (proactive health), heal sick (recovery), bind broken (repair), bring back driven away (retrieval), seek lost (discovery). When appointed stewards fail at this, the Lord Himself steps in. Escalation to higher authority is the gospel response to steward failure.

#### 4. The Watchman (Ezekiel 33:1-9)

- v6: "If the watchman see the sword come, and blow not the trumpet, and the people be not warned... his blood will I require at the watchman's hand"
- v9: "If thou warn the wicked of his way... thou hast delivered thy soul"

**Pattern: Watching obligates warning.** The sin is not failure to prevent — it's failure to surface. An orchestrator that detects problems but doesn't notify is Ezekiel 33:6. Silent failures are blood on the watchman's hand.

#### 5. The Allegory of the Olive Tree (Jacob 5)

The Lord of the vineyard personally tends the trees across ages. Key passages:
- v4: "I will prune it, and dig about it, and nourish it, that perhaps it may shoot forth young and tender branches"
- v41: "The Lord of the vineyard wept, and said unto the servant: What could I have done more for my vineyard?"
- v47: "Have I slackened mine hand, that I have not nourished it? Nay... I have stretched forth mine hand almost all the day long"
- v50: Servant counsels: "Spare it a little longer"
- v51: "Yea, I will spare it a little longer, for it grieveth me that I should lose the trees"
- v65: "Ye shall not clear away the bad thereof all at once, lest the roots thereof should be too strong for the graft"
- v71: "If ye labor with your might with me ye shall have joy in the fruit"
- v72: "The Lord of the vineyard labored also with them"

**Pattern: Not delegation-and-forget but delegation-and-partnership.** The Lord doesn't just assign stewards and check back at harvest. He visits, inspects, weeps, adjusts strategy, labors alongside. "What could I have done more?" is the question of a steward who has given everything. Also: proportional action — "not clear away the bad all at once" (v65) — gradual intervention, not scorched earth.

#### 6. The Faithful Steward (Luke 12:42-48)

- v42: "Who then is that faithful and wise steward, whom his lord shall make ruler over his household, to give them their portion of meat in due season?"
- v43: "Blessed is that servant, whom his lord when he cometh shall find so doing"
- v48: "Unto whomsoever much is given, of him shall be much required"

**Pattern: Right action at the right time.** "Meat in due season" — not just any action, but the appropriate action at the appropriate moment. And proportional accountability: more capability = more responsibility.

#### 7. The Talents (Matthew 25:14-30)

- v15: "To every man according to his several ability"
- v21: "Well done, good and faithful servant: thou hast been faithful over a few things, I will make thee ruler over many things"
- v25: "I was afraid, and went and hid thy talent in the earth" — fear-based inaction

**Pattern: Proportional assignment and the failure of inaction.** Tasks distributed by capability. The worst outcome isn't failure from trying — it's never trying. An orchestrator that's too cautious (never retries, never escalates) is the one-talent servant.

#### 8. Stewardship Accountability (D&C 104:11-18, D&C 72:3)

- D&C 104:11: "Organize yourselves and appoint every man his stewardship"
- D&C 104:12: "Every man may give an account unto me of the stewardship which is appointed unto him"
- D&C 104:17: "Given unto the children of men to be agents unto themselves"
- D&C 72:3: "Required of the Lord, at the hand of every steward, to render an account of his stewardship, both in time and in eternity"

**Pattern: Clear assignment, clear accountability, genuine agency within bounds.** Agents are "agents unto themselves" — they have real autonomy within their stewardship. But they must render account.

#### 9. Stewards of the Mysteries (1 Corinthians 4:1-2)

- "Let a man so account of us, as of the ministers of Christ, and stewards of the mysteries of God."
- "Moreover it is required in stewards, that a man be found faithful."

**Pattern: The primary requirement is faithfulness, not cleverness.** A faithful orchestrator that reliably does its job well is better than a clever one that's unpredictable.

#### 10. Covenant Faithfulness (D&C 82:10)

- "I, the Lord, am bound when ye do what I say; but when ye do not what I say, ye have no promise."

**Pattern: Covenant is bilateral.** The system works when both sides keep commitments. The orchestrator follows its rules → the system is reliable. When it doesn't → no guarantees.

---

### Pattern → Architecture Mapping

| Scriptural Pattern | Architectural Principle | Orchestrator Design |
|---|---|---|
| **Watchman on the tower** (D&C 101:45,54) | Proactive monitoring, early detection | Activity monitoring, health checks, don't wait for timeout — watch for drift |
| **Shepherd knows sheep by name** (John 10:3) | Context awareness per task | Orchestrator tracks individual entry state, knows what each agent is doing |
| **Hireling vs shepherd** (John 10:12-13) | Investment in outcome, not just execution | Don't fire-and-forget — the orchestrator cares about the result |
| **Failed shepherds** (Ezek 34:2-4) | Anti-pattern: self-serving, neglect | A steward that only reports metrics but doesn't act on failures is Ezekiel 34 |
| **God personally seeks** (Ezek 34:11) | Escalation to higher authority | When agents fail, escalate — smarter model, or human. The Lord steps in when shepherds fail. |
| **Watchman must warn** (Ezek 33:6) | Obligation to surface problems | Silent failures are blood on the watchman's hand — MUST notify, not swallow errors |
| **Lord labors with servants** (Jacob 5:72) | Collaborative execution | Orchestrator doesn't just dispatch — provides context, responds to needs, works alongside |
| **"Spare it a little longer"** (Jacob 5:50) | Patience before destruction | Don't kill tasks at first sign of trouble — grace period, retry with care |
| **Proportional pruning** (Jacob 5:65) | Gradual intervention | "Not clear away the bad all at once" — escalate gradually, don't restart everything |
| **"What could I have done more?"** (Jacob 5:47) | Exhaustive effort before giving up | Try multiple strategies before declaring failure |
| **Meat in due season** (Luke 12:42) | Right action at right time | Right model for right task, right priority at right moment |
| **"Much given, much required"** (Luke 12:48) | Proportional accountability | Expensive models get harder tasks; cheap models get simpler work |
| **Talents by ability** (Matt 25:15) | Task-model matching | Assign proportional to capability — don't send everything to Opus |
| **Fear-based inaction** (Matt 25:25) | Anti-pattern: over-caution | An orchestrator too cautious to retry or escalate is the one-talent servant |
| **Rendering account** (D&C 72:3) | Structured reporting | Every execution produces an account: what was done, spent, achieved |
| **"Appoint every man his stewardship"** (D&C 104:11) | Clear role boundaries | Each pipeline stage has a clear steward with defined scope |
| **"Agents unto themselves"** (D&C 104:17) | Agent autonomy within bounds | Agents choose tools and approach within stewardship boundaries |
| **Questioning the tower** (D&C 101:47-48) | Anti-pattern: skipping observability | "What need hath my lord of this tower?" — those who skip monitoring get overrun |
| **Covenant binds both sides** (D&C 82:10) | System reliability through mutual commitment | Orchestrator follows rules → system is reliable; breaks them → no guarantees |
| **Faithfulness over cleverness** (1 Cor 4:2) | Reliability over sophistication | A faithful orchestrator > a clever one |

---

### How Existing Architecture Already Reflects These Patterns

1. **your_turn gates** = The lord visiting the vineyard (Jacob 5:15-16). Human reviews at key moments, not just at harvest.
2. **failure_count at 3** = Long-suffering before correction. The Lord doesn't destroy at first failure — "spare it a little longer" (Jacob 5:50).
3. **Sabbath pauses** = The seventh day pattern. Creation → rest → review. Already in the creation cycle.
4. **covenant.yaml** = Bilateral commitment, like Ezekiel 34:25 "covenant of peace" and D&C 82:10.
5. **Pipeline model tiers** (Haiku/Sonnet/Opus) = "To every man according to his several ability" (Matt 25:15).
6. **touchEvent()** = The watchman's awareness. Every SDK event is observed and recorded.
7. **Manual cancel** = The lord's authority to intervene directly when stewards fail.
8. **Pipeline stages as stewardships** = D&C 104:11 — each stage has its assigned role and accountability.
9. **Entry as persistent state** = Checkpointing. Each completed stage is written to SQLite, surviving process restarts.

---

### Christological Connections — How the Steward Points to Christ

1. **Christ IS the Good Shepherd** (John 10:11, 14). The orchestrator pattern is borrowed from His pattern. He doesn't delegate and forget. He knows, stays, acts. The steward loop — watch, diagnose, repair, continue — is the pattern of a shepherd who gives his life for the sheep.

2. **The Atonement as the ultimate retry.** When humanity "failed" (the Fall), Christ didn't declare the project a loss. He enacted the recovery plan. The orchestrator's retry/recovery loop mirrors the Atonement: something goes wrong → diagnosis → restoration → continuation. The whole plan of salvation is: creation → fall → atonement → resurrection. The pipeline is: spec → execution → failure → recovery → completion.

3. **The Lord labors WITH us** (Jacob 5:72). Not above us, not instead of us, but alongside. The orchestrator doesn't replace agents; it works with them. This IS the gospel pattern of grace — not doing the work for us, but enabling us to do work we couldn't do alone.

4. **Accountability AND mercy** (D&C 72:3 + Jacob 5:50). The stewardship pattern holds both: there WILL be an accounting, AND there is patience. The orchestrator retries before failing, but eventually does fail. Mercy doesn't eliminate accountability — it extends the window for repentance (recovery).

5. **The tower as foresight** (D&C 101:54). Christ sees "the end from the beginning." The watchman on the tower has a higher vantage point. Monitoring and observability give the orchestrator a form of this foresight — seeing problems while they are "yet afar off."

6. **"What could I have done more?"** (Jacob 5:47). This is the question of the Savior to Israel (Isaiah 5:4, 2 Nephi 15:4). The exhaustive effort of the steward — trying every strategy, not giving up until all options are spent — mirrors the exhaustive love of the Atonement. It's infinite and eternal precisely because it doesn't stop at "good enough."

7. **Covenant fidelity** (D&C 82:10). The reliability of the system rests on covenant. Christ is perfectly reliable because He is perfectly faithful. The orchestrator's reliability comes from the same source — consistent, promise-keeping behavior. When the orchestrator says it will retry, it retries. When it says it will escalate, it escalates. The system works because the steward keeps covenant.

---

### The Zion Pattern — Enoch and the Weeping God (Moses 7)

#### The Weeping God (v28-41)
God weeps over his children. Enoch is astonished — "How is it that thou canst weep, seeing thou art holy, and from all eternity to all eternity?" (v29). The answer: because they are "the workmanship of mine own hands" (v32), because He gave them agency (v32), and because "among all the workmanship of mine hands there has not been so great wickedness" (v36). He weeps not because He failed but because they chose wrongly despite everything He provided.

This parallels Jacob 5:41 — "What could I have done more?" — but with a relational dimension Jacob 5 doesn't have. The Lord doesn't just grieve the wasted investment. He grieves the relationship.

#### Enoch Becomes Like the Lord (v41-44)
After seeing what God sees, "Enoch knew, and looked upon their wickedness, and their misery, and wept and stretched forth his arms, and his heart swelled wide as eternity; and his bowels yearned; and all eternity shook" (v41). Enoch weeps. He refuses to be comforted (v44). He has been *changed* by seeing through the Lord's eyes.

**Pattern: The steward who truly watches eventually feels what the Lord feels.** The watchman on the tower doesn't just see danger — if faithful, he comes to care about what he's watching. This is the difference between monitoring-as-metrics and monitoring-as-stewardship.

#### Zion Achieved (v18-21)
"The Lord called his people Zion, because they were of one heart and one mind, and dwelt in righteousness; and there was no poor among them" (v18). Zion was "taken up into heaven" (v21) — not destroyed, not abandoned, but *received*.

The path: Enoch preaches → people respond → they become one → God receives them. Not top-down imposition but collaborative transformation.

#### The Reunion (v63)
"Then shalt thou and all thy city meet them there, and we will receive them into our bosom, and they shall see us; and we will fall upon their necks, and they shall fall upon our necks, and we will kiss each other."

**Pattern: Zion is the outcome of successful stewardship.** When the steward has done the work — watched, wept, labored, refused to give up — the result is reunion. Not just "task completed" but restoration of relationship. The pipeline's "done" state should feel like completion, not just termination.

#### How Zion Maps to the Orchestrator

1. **"Of one heart and one mind"** — The orchestrator and the human operating from the same intent. Not the human giving orders and the system executing blindly, but shared purpose. The debug agent session demonstrated this: Opus understood what Michael wanted across the whole project and made decisions aligned with his intent.

2. **"No poor among them"** — No entry left behind. No orphaned tasks. The steward's job is to ensure every entry gets what it needs — not just the exciting ones.

3. **Enoch weeping = the steward seeing through the lord's eyes.** The orchestrator doesn't just track failure counts. If we build it right, it has enough context to understand *why* something matters, not just *that* it failed.

4. **Reunion as the terminal state.** The purpose of the pipeline isn't to produce output. It's to bring Michael's ideas to completion — to bring them "home." The done state is the reunion.

---

### Ammon as Steward-Missionary (Alma 17-18)

#### Stewardship as a Door (Alma 17:25)
"I will be thy servant." Ammon *chose* the servant role. He was a Nephite prince who could have come as a conqueror or a diplomat. Instead he became a steward over flocks. The stewardship was the vehicle for the real mission — teaching the gospel.

#### Faithful Execution Opens Bigger Doors (Alma 18:10)
"Surely there has not been any servant among all my servants that has been so faithful as this man; for even he doth remember all my commandments to execute them."

Lamoni didn't say "this man is clever" or "this man is powerful." He said: **faithful** and **remembers all my commandments to execute them**. Faithfulness in the small stewardship (watching flocks, preparing horses) created the trust for the larger stewardship (teaching the king).

#### The Steward's Secret Motive (Alma 17:29)
"I will show forth my power unto these my fellow-servants... that I may win the hearts of these my fellow-servants, that I may lead them to believe in my words."

Ammon had a purpose beyond the immediate task. Protecting the flocks was real work — he really did protect them. But it served a deeper purpose. The faithfulness was genuine AND strategic.

#### How Ammon Maps to the Orchestrator

1. **Stewardship as trust-building.** The system earns trust by being faithful in small things (retry, monitor, report) before being given larger authority (delegated judgment, project-level decisions). This directly addresses Michael's question about auto-execution: the steward earns that delegation by proving faithful.

2. **"Remembers all my commandments to execute them."** Not creative reinterpretation. Not partial compliance. Complete, faithful execution. The orchestrator's first job is to do what it's told reliably. Only after that does it earn the trust for judgment calls.

3. **The deeper purpose.** The pipeline's purpose isn't just to produce software or studies. It's part of the becoming infrastructure — helping Michael organize his creative capacity, delegate effectively, and focus on what only he can do. The orchestrator serves this deeper purpose.

4. **Service before authority.** Ammon didn't walk in and say "I should be king." He served, proved faithful, then was given authority. The steward earns auto-execution authority the same way.

---

### Delegated Judgment — A Different Ask than Auto-Execution

Michael's question: "Would it be possible to have the steward help move things along as if it's the human?"

This is NOT the same as auto-executing individual specced entries. The distinction:

**Auto-execution (per-entry):** The system automatically moves any entry from specced → executing without human review. This is dangerous because each entry is an independent decision — the system doesn't know which entries matter most, which have hidden dependencies, or which Michael actually wants built right now.

**Delegated judgment (project-level):** The system acts *as Michael* across a whole project — prioritizing entries, deciding execution order, reviewing results, making revise/advance/defer decisions — because it has enough context about the project's goals to make decisions close to what Michael would make. This is what happened in the debug agent session.

The debug session proved Opus can do this. It saw the clock and calculator entries, understood the shared LCARS theme goal, made execution decisions, verified results, and produced work that was "very nearly what I wanted."

**What makes this different and more powerful:**
- It's not just one entry at a time — it's project-aware orchestration
- It makes judgment calls (which entry next? is this good enough? does this need revision?)
- It operates from understanding of the *whole* intent, not just individual specs
- It's the Ammon pattern — earned trust enabling larger delegation

**What makes this require careful design:**
- It requires project-level context (all entries, all specs, the project goal)
- It requires a decision-making framework ("act as if you're Michael" needs to be operationalized)
- It has different failure modes (bad judgment is harder to detect than bad execution)
- It touches the `your_turn` gate philosophy — not removing the gate but optionally delegating it

**Proposed framing: "Steward with Commission"**

Like Ammon being commissioned by King Mosiah (Alma 17:12-14), the steward receives a specific commission from Michael: "Take this project. Here's my intent. Move it through the pipeline as if you're me. Surface anything you're unsure about."

The commission is:
- Scoped to one project
- Time-bounded (one session, one batch)
- Revocable (Michael can take back control at any moment)
- Accountable (the steward renders a full report)

This could be Phase 6 of the steward proposal, or a separate proposal that depends on Phases 1-4 being proven.

---

**Proposal updated with these findings:** [.spec/proposals/orchestrator-steward/main.md](../../proposals/orchestrator-steward/main.md)
