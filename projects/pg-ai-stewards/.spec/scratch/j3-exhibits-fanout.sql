-- =====================================================================
-- J.3 — Apply fan-out to 8218aa77 — exhibits library
-- =====================================================================
-- Takes the survey from work_item 8218aa77 (everyday-science-to-exhibits)
-- and produces 6 actual exhibit briefs via the decompose-fanout pipeline.
--
-- Hand-crafts the decompose stage_results from the survey's review
-- output (saves the LLM cost of re-decomposing — we already know what
-- the 6 exhibits are). Triggers spawn via maturity=verified.
--
-- Each child runs research-write with a binding question that
-- explicitly asks for the 6-field exhibit-brief structure:
--   1. Story / Example
--   2. Real-World Application
--   3. Museum Demo
--   4. Underlying Science
--   5. History
--   6. Build Materials & Instructions
--
-- Children land at projects/space-center/exhibits/<slug>.md
-- Aggregator (synthesis=true) writes README.md at the same path.
--
-- Cost estimate: 6 × $0.70 + $0.30 aggregate = ~$4.50
-- =====================================================================

WITH brief_template AS (
    SELECT
        $TEMPLATE$Produce a complete exhibit brief for a {{EXHIBIT}} interactive exhibit, suitable for a local science center serving a rural Missouri community.

Structure the output with these 6 sections (use ## headings, in this exact order):

1. **Story / Example** — a narrative hook tying the science to something visitors can relate to. Specific people, specific places where possible.
2. **Real-World Application** — where this science shows up in everyday life, local industry, or current research.
3. **Museum Demo** — exactly what visitors will interact with on the floor. Be concrete about the physical setup, controls, and what visitors see/hear/touch.
4. **Underlying Science** — the physics/chemistry/biology of why it works. Calibrated to a curious adult; include a one-paragraph "for the deeper reader" sub-section if useful.
5. **History** — discovery, key inventors/scientists, dates. Source-verify every name and date.
6. **Build Materials & Instructions** — itemized parts list with sources (Amazon, hardware store, McMaster-Carr where applicable), target budget $100–$300, step-by-step assembly notes, signage suggestions, safety considerations, and known failure modes.

HARD CONSTRAINTS:
- Source-verify every direct quote and every historical claim. Cite primary sources where possible. Use markdown links: [Source](url).
- Honor uncertainty: where the literature is thin or there's a real safety/compliance concern, say so explicitly rather than glossing over.
- Target length: 1500-2500 words.
- Output STANDALONE markdown ready to file at projects/space-center/exhibits/{{SLUG}}.md. No code fences around the whole thing.

EXHIBIT-SPECIFIC CONTEXT:
{{CONTEXT}}$TEMPLATE$ AS tmpl
),
manifest AS (
    SELECT jsonb_build_object(
        'rationale', 'The 8218aa77 survey identified 6 buildable, cited candidates spanning physics, computer science, chemistry, biology, math, and local industry. Each child produces a full 6-field exhibit brief at its own file_destination; aggregator writes an index README with cross-cutting themes.',
        'children', jsonb_build_array(
            -- 1. Crystal Radio (Physics / EM)
            jsonb_build_object(
                'slug', 'exhibit-crystal-radio',
                'pipeline_family', 'research-write',
                'project_association', 'space-center',
                'file_destination', 'projects/space-center/exhibits/crystal-radio.md',
                'cost_cap_micro', 800000,
                'binding_question', replace(replace(replace(
                    (SELECT tmpl FROM brief_template),
                    '{{EXHIBIT}}', 'Crystal Radio (no-battery AM receiver)'),
                    '{{SLUG}}', 'crystal-radio'),
                    '{{CONTEXT}}',
                    'A diode + coil + earphone receiver that draws power only from incoming radio waves. The earliest mass-produced consumer electronics. Locally interesting as a foil to Webster Electric Cooperative grid-power exhibits. Per Free Science Project, reception is location-dependent — strong AM stations and good ground/antenna matter; verify station strength before committing to building, or include a fallback RF-detector LED demo. Sources to verify: freescienceproject.com/projects/CrystalRadio, history via primary biographies of Greenleaf Whittier Pickard (silicon carbide detector, 1906) and the broader cat-whisker era.')
            ),
            -- 2. CS Unplugged Sorting Network (Computer Science)
            jsonb_build_object(
                'slug', 'exhibit-cs-unplugged-sorting-network',
                'pipeline_family', 'research-write',
                'project_association', 'space-center',
                'file_destination', 'projects/space-center/exhibits/cs-unplugged-sorting-network.md',
                'cost_cap_micro', 800000,
                'binding_question', replace(replace(replace(
                    (SELECT tmpl FROM brief_template),
                    '{{EXHIBIT}}', 'CS Unplugged Sorting Network (floor-tape, no computer)'),
                    '{{SLUG}}', 'cs-unplugged-sorting-network'),
                    '{{CONTEXT}}',
                    'A walk-through sorting network where 2-6 visitors each pick a number card and physically traverse comparison nodes painted on the floor; they emerge sorted at the other end. Materials: floor tape, laminated number cards, signage. Total cost under $40. CC-licensed material from University of Canterbury (csunplugged.org). The exhibit teaches parallel algorithms, comparison-based sorting, and big-O intuition without any code. Verify: csunplugged.org/en/principles/, original Bell/Witten/Fellows curriculum, and the for-profit license terms (CC allows commercial use).')
            ),
            -- 3. Indicating Electrolysis (Chemistry)
            jsonb_build_object(
                'slug', 'exhibit-indicating-electrolysis',
                'pipeline_family', 'research-write',
                'project_association', 'space-center',
                'file_destination', 'projects/space-center/exhibits/indicating-electrolysis.md',
                'cost_cap_micro', 800000,
                'binding_question', replace(replace(replace(
                    (SELECT tmpl FROM brief_template),
                    '{{EXHIBIT}}', 'Indicating Electrolysis (water splitting with pH visualization)'),
                    '{{SLUG}}', 'indicating-electrolysis'),
                    '{{CONTEXT}}',
                    'A 9V battery + stainless-steel screws + Epsom salts + pH indicator setup that visibly splits water into hydrogen and oxygen with colorful pH-zone swirls. From Exploratorium Teacher Institute Snacks. Negligible material cost; the budget goes to enclosure, signage, and splash containment for public use. Safety/compliance angle is real: 2010-era classroom protocols predate current museum/fire-code chemical-handling guidelines. Address ADA-compliant height, child-safe concentrations, splash guards, MSDS for universal indicator. Verify: exploratorium.edu/snacks/indicating-electrolysis, history of Faraday electrolysis (1834), and current informal-science chemistry safety standards.')
            ),
            -- 4. Bacteriopolis Winogradsky Column (Biology)
            jsonb_build_object(
                'slug', 'exhibit-bacteriopolis-winogradsky',
                'pipeline_family', 'research-write',
                'project_association', 'space-center',
                'file_destination', 'projects/space-center/exhibits/bacteriopolis-winogradsky.md',
                'cost_cap_micro', 800000,
                'binding_question', replace(replace(replace(
                    (SELECT tmpl FROM brief_template),
                    '{{EXHIBIT}}', 'Bacteriopolis (Winogradsky Column living biology exhibit)'),
                    '{{SLUG}}', 'bacteriopolis-winogradsky'),
                    '{{CONTEXT}}',
                    'A clear column packed with pond mud + egg yolk + shredded newspaper that grows a multicolored bacterial ecosystem over 4-8 weeks. Each color band reveals a different metabolic niche (sulfate reducers, purple sulfur bacteria, cyanobacteria). Material cost negligible. The challenge is the maintenance protocol: light-cycle exposure, observation timeline, and what to do when the column eventually decays. Based on Sergei Winogradsky''s 1880s classical microbiology technique (founder of microbial ecology). Verify: Exploratorium Snacks (Bacteriopolis), Winogradsky biographical material, and protocols from contemporary classroom adaptations.')
            ),
            -- 5. Symmetry & Polyhedra (Math)
            jsonb_build_object(
                'slug', 'exhibit-symmetry-polyhedra',
                'pipeline_family', 'research-write',
                'project_association', 'space-center',
                'file_destination', 'projects/space-center/exhibits/symmetry-polyhedra.md',
                'cost_cap_micro', 800000,
                'binding_question', replace(replace(replace(
                    (SELECT tmpl FROM brief_template),
                    '{{EXHIBIT}}', 'Symmetry & Polyhedra Construction (snap-together geometry)'),
                    '{{SLUG}}', 'symmetry-polyhedra'),
                    '{{CONTEXT}}',
                    'A build-it-yourself table where visitors snap together polygons (Polydrons, Geoshapes, or 3D-printed equivalents) to construct Platonic solids, Archimedean solids, and other symmetric forms. Mirrors and pattern blocks support a parallel symmetry station. Based on the California Math Show (NSF/Eisenhower funded, ~1996). The exhibit is the most expensive of the six — commercial Polydron sets run $150-$250 for a class set. Tradeoff to address: durable commercial sets last a decade; 3D-printed equivalents shift labor to fabrication. Verify: geom.uiuc.edu/~addingto/mathshow.html, history of Felix Klein''s Erlangen program (1872) for the symmetry-as-group-action framing, and current pricing for Polydrons / Geoshapes / Magformers.')
            ),
            -- 6. Rural Electrification — Webster Electric Coop (Local Industry)
            jsonb_build_object(
                'slug', 'exhibit-rural-electrification-webster-coop',
                'pipeline_family', 'research-write',
                'project_association', 'space-center',
                'file_destination', 'projects/space-center/exhibits/rural-electrification-webster-coop.md',
                'cost_cap_micro', 800000,
                'binding_question', replace(replace(replace(
                    (SELECT tmpl FROM brief_template),
                    '{{EXHIBIT}}', 'Rural Electrification — Webster Electric Cooperative Partnership'),
                    '{{SLUG}}', 'rural-electrification-webster-coop'),
                    '{{CONTEXT}}',
                    'A 1949 farmhouse-kitchen vignette + interactive circuit board + pedal-powered generator that tells the story of the Rural Electric Administration (1935, REA) and modern grid science. Modeled on the South Dakota Agricultural Heritage Museum''s "Power to the People" exhibit, which partnered with Basin Electric, East River Cooperative, Sioux Valley Energy, and H&D Cooperative for sponsorship and artifact loans. The local parallel: Webster Electric Cooperative serves Webster County, MO; their archives may have similar photographs and equipment. Verify: sdstate.edu/south-dakota-agricultural-heritage-museum/power-people-electrifying-rural-south-dakota, REA history (1935 executive order, 1936 act), Webster Electric Cooperative history and contact information. Frame the build as a partnership concept note: what artifacts could WEC contribute, what stories from local farmers, what modern grid topics to cover (smart meters, demand response, distributed solar).')
            )
        ),
        'aggregate', jsonb_build_object(
            'destination', 'projects/space-center/exhibits/README.md',
            'synthesis', true
        )
    ) AS m
)
INSERT INTO stewards.work_items (
    pipeline_family, current_stage, slug, input, intent_id, actor,
    project_association, stage_results, maturity, status
)
SELECT
    'decompose-fanout',
    'decompose',
    'science-center-exhibits-batch-1',
    jsonb_build_object('binding_question',
        'Take the 6 buildable candidates from the 8218aa77 survey (Crystal Radio, CS Unplugged Sorting Network, Indicating Electrolysis, Bacteriopolis Winogradsky Column, Symmetry & Polyhedra, Rural Electrification + Webster Electric Cooperative) and produce a full 6-field exhibit brief for each, plus an index README. Each child writes to projects/space-center/exhibits/<slug>.md.'),
    (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
    'michael',
    'space-center',
    jsonb_build_object(
        'context_gather', jsonb_build_object('output',
            'Pre-populated context: the 8218aa77 survey already did the gathering. Skipping context_gather LLM call and going straight to a hand-crafted decompose manifest.'),
        'decompose', jsonb_build_object('output', (SELECT m FROM manifest))
    ),
    'planned',
    'completed'
RETURNING id, slug, pipeline_family;

-- Trigger spawn.
UPDATE stewards.work_items
   SET maturity = 'verified'
 WHERE slug = 'science-center-exhibits-batch-1'
RETURNING id, slug, maturity;

-- Inspect what got spawned.
SELECT slug, pipeline_family, status, maturity, file_destination
  FROM stewards.work_items
 WHERE slug = 'science-center-exhibits-batch-1'
    OR slug LIKE 'exhibit-%'
    OR slug = 'science-center-exhibits-batch-1-aggregator'
 ORDER BY created_at;
