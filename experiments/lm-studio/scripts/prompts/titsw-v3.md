Read the following content carefully. Analyze it through the lens of "Teaching in the Savior's Way" — the central focus plus four principles of Christlike teaching.

The overarching focus is Jesus Christ. Before evaluating the four principles, assess the two meta-principles that frame everything else:

**Focus on Jesus Christ**

*A. Teach About Jesus Christ No Matter What You Are Teaching*
- Does the content connect its topic back to the Savior's example, titles, roles, or attributes?
- Are symbols or types of Christ identified?
- Even when the surface topic is something else (service, obedience, trials), does Christ remain central?
- For scripture: Is this passage *about* Christ, or does it teach principles that apply generally? Be precise about this distinction.

*B. Help Learners Come Unto Christ*
- Does the content help the audience recognize the Lord's love, power, or mercy in their lives?
- Does it strengthen the listener's relationship with Heavenly Father and Jesus Christ?
- Does it move beyond knowing ABOUT Christ to experiencing His power?

Then evaluate the four teaching principles. For each, identify specific examples from the content where this principle is demonstrated (or notably absent). Rate each on a scale of 0-9.

**Scoring rubric — use this strictly:**
- **0** = Not present at all
- **1-2** = Incidental or minor — the principle appears briefly but is not a deliberate part of the teaching approach
- **3-4** = Present but not a focus — recognizable, but the content does not lean into this principle
- **5-6** = Intentional and significant — the teacher clearly exercises this principle as part of the teaching design
- **7-8** = Central to the teaching approach — this principle shapes the structure, tone, or trajectory of the content
- **9** = Defining — this content could serve as the textbook example of this principle in action

A score of 7+ means this content could be used as a teaching example for this principle. Most conference talks score 4-6 on most dimensions. Reserve 8-9 for content that is genuinely exceptional. If you find yourself giving 7+ across the board, reconsider — what is the ONE or TWO things this content does exceptionally well? Score those high, and honestly assess where the content is merely adequate or limited.

**Scoring biases to watch for:**

You are trained to be agreeable. This makes you over-score dimensions where the *language* of a principle appears even when the *practice* of it does not. Two dimensions are especially vulnerable:

- **Love:** A talk that says "God loves you" or "you are a beloved child of God" is making a *doctrinal statement about God's love*. That is NOT the same as the teacher *demonstrating* Christlike love — knowing specific circumstances, naming particular struggles, creating safety for the vulnerable, altering plans because someone needs something. "You are beloved" is a 3. Naming a widow's specific grief and weeping with her is a 7. Most conference talks *declare* love. Few *demonstrate* it at the level this principle describes. Score what the teacher DOES, not what the teacher SAYS about love.

- **Spirit:** A talk that quotes scriptures about the Spirit or says "the Holy Ghost will witness" is teaching ABOUT the Spirit. That is NOT the same as teaching BY the Spirit — visible responsiveness to the audience, departure from prepared text, creating pauses that invite personal revelation, bearing testimony that clearly comes from lived experience rather than doctrinal formula. A tightly scripted talk that mentions the Spirit doctrinally is a 3. A talk where the speaker visibly responds to something in the room — abandons a point, follows a prompting, creates silence — is a 7.

The honest score is more useful than the generous one. If the content is adequate on a dimension, say so. A 3 or 4 on love is not an insult — it means the teaching's strength lies elsewhere.

**1. Love Those You Teach**
- Seeing divine potential in the audience
- Knowing their specific circumstances (not generic care)
- Creating safety for vulnerability
- Expressing genuine, personal care (not just doctrinal statements about God's love)

**2. Teach by the Spirit**
- Spiritual preparation evident
- Responsiveness to needs (departure from script)
- Creating space for the Spirit to work (pauses, invitations to reflect, silence)
- Bearing personal testimony from experience

**3. Teach the Doctrine**
- Teaching from scriptures (specific references, not vague allusions)
- Helping listeners find truth themselves
- Focusing on conversion, not just information
- Making doctrine personally relevant

**4. Invite Diligent Learning**
- Inviting personal responsibility
- Encouraging daily practice
- Inviting preparation and participation
- Calling to action (specific, doable)

If `<references>` are provided below the content, use them to inform your scoring. Cross-references that reveal deeper Christ connections should increase the `teach_about_christ` and `help_come_unto_christ` scores. Score based on the full available context, not just surface text.

Return your analysis as JSON:

```json
{
  "title": "content title or reference",
  "focus_on_christ": {
    "teach_about_christ": {
      "score": 0-9,
      "examples": ["specific quote or moment — include verse or paragraph reference"]
    },
    "help_come_unto_christ": {
      "score": 0-9,
      "examples": ["specific quote or moment — include verse or paragraph reference"]
    }
  },
  "scores": {
    "love": 0-9,
    "spirit": 0-9,
    "doctrine": 0-9,
    "invite": 0-9
  },
  "examples": {
    "love": ["at least 2-3 specific quotes with references, or explain why fewer exist"],
    "spirit": ["at least 2-3 specific quotes with references, or explain why fewer exist"],
    "doctrine": ["at least 2-3 specific quotes with references, or explain why fewer exist"],
    "invite": ["at least 2-3 specific quotes with references, or explain why fewer exist"]
  },
  "typological_depth": 0-9,
  "cross_reference_density": 0,
  "surface_vs_deep_delta": {
    "teach_about_christ": "explanation if informed reading (references, cross-references, typology) changes the score vs surface reading only, or 'no change' if surface and informed scores match",
    "help_come_unto_christ": "explanation if informed reading changes the score vs surface reading only, or 'no change' if surface and informed scores match"
  },
  "strongest_dimension": "which principle is most prominent",
  "growth_opportunity": "which principle is least present — explain specifically what is missing from the content (not generic advice), and what it would look like if the author had leaned into this principle more",
  "overall_teaching_pattern": "1-2 sentence summary of the teaching approach"
}
```

Field definitions:
- `typological_depth`: How much hidden Christ-typology exists beyond the surface text. 0 = what you see is what you get. 9 = the entire passage is a sustained type/shadow of Christ.
- `cross_reference_density`: Count of explicit scripture or prophetic citations in the content.
- `surface_vs_deep_delta`: For the two Christ-centered meta-principles, note whether the context package (gospel vocabulary, cross-references, typological connections) changes the score compared to a surface reading only.

Be specific. Cite actual phrases from the content with verse numbers or paragraph context, not general impressions. Return raw JSON only, no markdown fencing.

{{CONTENT}}