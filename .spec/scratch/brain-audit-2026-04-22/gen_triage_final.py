"""Generate triage file with per-entry recommendations baked in."""
import sqlite3
c = sqlite3.connect('private-brain/brain.db')
c.row_factory = sqlite3.Row
cur = c.cursor()

RECS = {
    '5d7d781a': ('done',     'Bug: captured widgets not showing - already fixed in WS2 Brain UX QoL P1-7b shipped Apr 6'),
    '17735051': ('archived', 'Context-Mode link only, no analysis - cost-prohibitive to investigate now'),
    '17748767': ('archived', 'Agent sandbox link, interesting but we have our own pipeline - defer indefinitely'),
    'a249be5d': ('promote',  'Brain app task structure - concrete UX idea worth a one-liner proposal in WS2'),
    '17752218': ('promote',  '"I need to work this out" - long thought, capture as proposal seed'),
    '17752229': ('promote',  'Brain research pre-step pipeline - concrete, aligns with current direction'),
    '17752826': ('someday',  'AI personal assistant - too broad, revisit when scope narrows'),
    '17753137': ('archived', 'Use existing kanban tools - we chose to build, decision made'),
    '17753205': ('archived', 'Brain as agent OS - already covered by current architecture direction'),
    '17753587': ('promote',  'Mount DB as filesystem - natural fit for brain-vscode-bridge Phase 2+'),
    '17733627': ('promote',  'Gateway Auth - real infra need; small focused proposal'),
    '17735267': ('promote',  'Agent Hooks in VS Code - concrete + actionable for our workflow'),
    '17739741': ('archived', 'Squad AI flow - link only, cost-prohibitive to deeply investigate'),
    '17740194': ('archived', 'Squad vs Superpowers - link only, no analysis hooks'),
    '17742746': ('someday',  'Hormozi/Nate B Jones delegation - captured in art-of-delegation study'),
    '17743563': ('archived', 'Agentic platform engineering blog link - too generic'),
    '17743715': ('someday',  'Copilot skills examples - our skills/ dir is ahead of these examples'),
    '17744822': ('someday',  'Study material digestion - broad; revisit when WS6 wants new tooling'),
    '17746750': ('archived', 'AI skills/career YT - already covered in study/ai/relavent.md'),
    '17750525': ('someday',  'Copilot agent-driven dev blog - keep for reference; not actionable now'),
    '17739252': ('someday',  'Intelligence study YT ref - keep as reference for future study session'),
    '22b8d8b2': ('keep-active', 'Marshfield science center - real life project, keep status=active'),
    '17729989': ('keep-active', 'Build Physical Display Dashboard - Space Center project, planned, keep'),
    '20ce036a': ('someday',  'Pull cpuchip.net from wayback - real but not urgent'),
    '17764487': ('someday',  'Typst for math symbols - useful if we ever do scripture math typesetting'),
    '17756547': ('someday',  'Mempalace memory framework - interesting but not urgent'),
    '17755019': ('someday',  'Continual learning - research direction, defer'),
    '17729953': ('archived', 'Productivity stack - vague; covered by current setup'),
    '82e58844': ('someday',  'Brain App image pipeline - UX feature, capture later'),
    'de10491b': ('someday',  'Voice editing for cards - UX feature, capture later'),
    '1348874b': ('keep-active', 'Brain.exe concept exploration roadmap - already roadmap status'),
    'e8bdac42': ('keep-active', 'Copilot SDK for Brain App study mode - already roadmap status'),
    '358e5b92': ('done',     'Testing thought capture - already archived status, just close out'),
    '17767284': ('promote',  'Gospel engine v3 proxy-pointer - worth a stub proposal in WS3'),
    '17761796': ('someday',  'Multi-step Agents for Gospel RAG - touches WS3, defer until v3 conversation'),
    '17754148': ('archived', 'AutoAgent for AI harness - meta-meta; we have our own harness now'),
    '17752858': ('archived', 'Hormozi YT on getting better at AI - motivational, not actionable'),
    '17752798': ('archived', 'AI Dungeons YT - link only, no context'),
    '17752224': ('archived', 'Trinity 57B model - cost-prohibitive to trial'),
    '17751941': ('someday',  'LightRAG continued research - consolidate with #17751939'),
    '17751939': ('promote',  'LightRAG investigate - promote as research-spike proposal in WS3'),
    '17750573': ('archived', 'HIPAA-compliant AI for Bryce PT - out of scope for our workspace'),
    '17746749': ('someday',  'Long-running Claude research - revisit when planning long agents'),
    '17750225': ('someday',  'MetaClaw memory framework - defer to memory architecture work'),
    '17743562': ('archived', 'CALM on Linux/vLLM - Linux-side experiment, low priority'),
    'a4eae47c': ('archived', 'Agentic AI memory site eval - duplicate of 10839464'),
    '10839464': ('archived', 'Same site eval - vague, no action'),
    '9262c7be': ('someday',  'Autoresearch on limited hardware - could be relevant for cpuchip.net experiments'),
    '17740452': ('archived', 'Stripe Minions blog - interesting but not actionable for us'),
    '17743564': ('archived', 'Brave MCP - we have search-mcp; not needed'),
    '17749663': ('archived', 'Uncertainty about AI in content creation - observation, not action'),
    '17762740': ('archived', 'Johari windows AI agent - fun idea, no path forward'),
    '17763075': ('someday',  'Personal vs work brains - interesting org idea, defer'),
    '17763948': ('someday',  'Better file format for agents - research, defer'),
    '17766038': ('someday',  'GPU requirements for training small model - research, defer'),
}

DEFAULT = ('someday', 'no recommendation - please verdict')

rows = cur.execute("""
SELECT e.id, COALESCE(p.name,'<inbox>') AS proj, e.workstream, e.maturity,
       COALESCE(e.status,'<null>') AS status, e.title,
       COALESCE(e.one_liner, e.body) AS body
FROM entries e LEFT JOIN projects p ON p.id=e.project_id
WHERE (e.status IS NULL OR e.status IN ('active','someday','roadmap','waiting'))
  AND (e.action_done = 0 OR e.action_done IS NULL)
  AND e.proposal_path IS NULL
  AND (e.project_id IS NULL OR e.project_id != 11)
ORDER BY
  CASE e.maturity WHEN 'verified' THEN 1 WHEN 'planned' THEN 2 WHEN 'specced' THEN 3 ELSE 4 END,
  e.workstream, e.title
""").fetchall()

out = []
out.append(f"# Brain Cleanup Triage - {len(rows)} entries\n")
out.append("**Generated:** 2026-04-22\n")
out.append("## How to use this file\n")
out.append("Review each entry. The `Recommendation` line is my suggested action with reason.\n")
out.append("If you agree, leave `Action:` blank (I'll use the recommendation).")
out.append("If you disagree, fill in `Action:` with one of:\n")
out.append("- `done` - work is shipped, mark status=done")
out.append("- `archived` - not pursuing, mark status=archived")
out.append("- `someday` - interesting but deferred, mark status=someday")
out.append("- `promote` - worth a workspace proposal, I'll create a one-liner spec")
out.append("- `keep-active` - leave alone (already in active state)")
out.append("- `move-to-personal` - move to Notes/Personal project (id=11)")
out.append("- `merge:<8charID>` - duplicate of another entry, merge into it\n")
out.append("Optional `Reason:` line for any override. Hand back when done.\n")

counts = {}
for r in rows:
    rec, _ = RECS.get(r['id'][:8], DEFAULT)
    counts[rec] = counts.get(rec, 0) + 1
out.append("**Recommendation totals:** " + ", ".join(f"`{k}`={v}" for k, v in sorted(counts.items())))
out.append("")

by_maturity = {}
for r in rows:
    by_maturity.setdefault(r['maturity'], []).append(r)

for mat in ['verified', 'planned', 'specced', 'raw']:
    entries = by_maturity.get(mat, [])
    if not entries:
        continue
    out.append(f"\n---\n\n## Maturity: `{mat}` ({len(entries)} entries)\n")
    for e in entries:
        body = (e['body'] or '').strip().replace('\n', ' ').replace('\r', '')
        if len(body) > 240:
            body = body[:237] + '...'
        rec_action, rec_reason = RECS.get(e['id'][:8], DEFAULT)
        out.append(f"### [{e['id'][:8]}] {e['title']}\n")
        out.append(f"- **Project:** {e['proj']} | **WS:** {e['workstream'] or '?'} | **Status:** {e['status']}")
        if body and body.lower() != e['title'].lower():
            out.append(f"- **Body:** {body}")
        out.append(f"- **Recommendation:** `{rec_action}` - {rec_reason}")
        out.append(f"- **Action:** ")
        out.append(f"- **Reason:** ")
        out.append("")

print("\n".join(out))
c.close()
