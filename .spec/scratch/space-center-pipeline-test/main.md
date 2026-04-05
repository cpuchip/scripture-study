# Space Center: Automated Pipeline Test Plan

*Status:* proposed
*Binding problem:* We built a full pipeline (6 maturity stages, review bot, execution gates) but haven't tested the fully automated path end-to-end on a real project. Space Center (project_id=4) is a low-stakes, fun project where we can practice the fully automated pipeline and see how we like it.

---

## Why Space Center

- It's Michael's dream business — a planetarium, space/science center, starship bridge simulator
- Low stakes — no production code, no users, no deadlines
- Small enough to observe the full cycle without drowning in entries
- Fun subject matter — keeps it energizing rather than feeling like homework
- Has related repos on GitHub (`cpuchip`) but hasn't been developed in brain yet
- Only 1 brain entry currently mentions it (as a project differentiator)

## What We're Testing

The fully automated pipeline path with **auto-continuation** (conceptual — not yet built):

1. **Create entries** → Seed Space Center project with 3-5 ideas (exhibits, simulator concepts, business model)
2. **Watch automatic research** → Review bot picks them up as stale raw entries, nudges them. We reply with enough context for the pipeline to advance.
3. **Observe advancement** → Pipeline advances from raw → researched → planned → specced
4. **Execute one** → Pick the most interesting specced entry and run the execution phase
5. **Verify** → Review the output, verify, and see what the full cycle feels like

## Test Entries (Seeds)

These should be created as brain entries in the Space Center project:

1. **"Starship Bridge Simulator — Game Design"** (category: ideas)
   - What kind of experience? Cooperative crew gameplay? Solo captain? Competition between bridges?
   - Age range. Technology stack. VR vs physical controls.

2. **"Planetarium Show Themes — First Season"** (category: ideas)
   - What 4-6 shows would launch the planetarium?
   - Mix of education and entertainment. Target audience.

3. **"Space Center Business Model"** (category: ideas)
   - Tickets, memberships, school groups, birthday parties, corporate events?
   - Revenue mix. What makes it sustainable?

4. **"Interactive Space Exhibits — Core Set"** (category: ideas)
   - What 5-10 exhibits anchor the experience?
   - Hands-on vs digital. Age-appropriate design.

5. **"Space Center Location & Requirements"** (category: ideas)
   - Square footage. Ceiling height for planetarium dome. Dark room requirements.
   - Parking. Zoning. Proximity to schools.

## What We're Observing

| Question | How We'll Know |
|----------|---------------|
| Does the nudge bot ask useful questions? | Read the session messages — are they specific to the entry? |
| Does research produce valuable output? | Is the research synthesis actually useful, or generic? |
| Does plan quality match the investment? | Would we use the plan, or rewrite it from scratch? |
| How many premium requests per full cycle? | Track: 0.33 per nudge + 0.33 per research + 1.0 per plan = ~1.66 per entry minimum |
| How long does the full cycle take? | From creation to specced, measuring wall-clock time |
| Does the pipeline feel collaborative or bureaucratic? | Subjective — does it help or does it feel like paperwork? |
| Do the VS Code session zombies bother Michael? | Count sidebar sessions after a round of nudges |

## Success Criteria

- All 5 entries reach "specced" maturity through the pipeline (even if some need mid-course nudges)
- At least 1 entry goes through the full cycle including execution
- Michael's honest assessment: "this was helpful" vs "this was paperwork"
- We identify at least 2 concrete improvements to the pipeline from the experience

## How To Run

1. Create the 5 seed entries (manually or via brain-app capture)
2. Assign all to Space Center project
3. **Wait.** Let the nudge bot pick them up at the next scan (entries become stale after 24h for raw)
4. When nudged, reply with enough context to be useful (2-3 sentences each)
5. Watch the advancement happen
6. After specced: execute one, verify it
7. Write up observations in this scratch file

## Notes

- If auto-continuation is built before this test, use it on 2-3 entries and leave the others manual for comparison
- Haiku-class models need Space Center context in the prompt (not in training data) — verify the project context injection is working
- This test may reveal that the nudge bot asks too-generic questions for a topic it has no background on

---

## Observations (fill in during test)

*Not yet started.*
