-- =====================================================================
-- Batch L.1.1.11 — Judge prompt template
-- =====================================================================
-- Single canonical template + per-pipeline-family override row pattern.
-- The judge surface (L.1.1.8) reads this when constructing the corpus
-- overview message it returns to the consuming agent.
-- =====================================================================


CREATE TABLE IF NOT EXISTS stewards.judge_templates (
    scope          text NOT NULL,                -- 'canonical' | 'pipeline:<family>'
    template_text  text NOT NULL,
    description    text NOT NULL DEFAULT '',
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (scope)
);

COMMENT ON TABLE stewards.judge_templates IS
'Batch L.1.1.11: judge prompt templates. scope=''canonical'' is the universal default; scope=''pipeline:<family>'' overrides for a specific pipeline. Read by judge surface (L.1.1.8) when constructing the corpus overview surfaced to the consuming agent.';


INSERT INTO stewards.judge_templates (scope, template_text, description)
VALUES (
    'canonical',
    $TMPL$You have been delivered an oversized tool result from **{{tool_name}}** — {{source_bytes}} bytes, indexed into a per-message mini-corpus of {{parent_count}} parent chunks and {{leaf_count}} leaf chunks (~512 tokens each, embedded into pgvector for semantic search).

Your binding question: **{{binding_question}}**

Top-level overview of the source:
> {{top_overview}}

Within your stewardship over the binding question, judge:

1. **Is the fruit good?** Is this content credible, on-topic, and worth preserving? If not, you may discard the corpus.
2. **What is most precious to save?** Use `retrieve_from_corpus(corpus_msg_id={{message_id}}, query=...)` to pull specific chunks; mark anything you'll cite later via `mark_engram_important`. The contextual blurbs prepended to each leaf situate it within the document for retrieval quality.
3. **What should be discarded?** Anything noise / off-topic / suspect. Discarded chunks are not deleted but won't be surfaced again automatically.

You have full agency here. Surface only what matters; pass on what doesn't.$TMPL$,
    'Canonical judge prompt — surfaced to consuming agent when an oversized tool result has been indexed into per-message overflow corpus. Variables: tool_name, source_bytes, parent_count, leaf_count, binding_question, top_overview, message_id.'
)
ON CONFLICT (scope) DO UPDATE
   SET template_text = EXCLUDED.template_text,
       description   = EXCLUDED.description,
       updated_at    = now();


CREATE OR REPLACE FUNCTION stewards.judge_template_for_pipeline(p_pipeline_family text)
RETURNS text LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_template text;
BEGIN
    IF p_pipeline_family IS NOT NULL THEN
        SELECT template_text INTO v_template
          FROM stewards.judge_templates
         WHERE scope = 'pipeline:' || p_pipeline_family;
        IF v_template IS NOT NULL THEN
            RETURN v_template;
        END IF;
    END IF;
    SELECT template_text INTO v_template
      FROM stewards.judge_templates
     WHERE scope = 'canonical';
    RETURN v_template;
END;
$FN$;

COMMENT ON FUNCTION stewards.judge_template_for_pipeline(text) IS
'Batch L.1.1.11: resolve the judge template for a pipeline family. Returns the pipeline-specific override if set, else the canonical template.';


-- =====================================================================
-- End of l18-judge-template.sql
-- =====================================================================
