-- =====================================================================
-- Batch R.2 — server-side document read (the injection source)
-- =====================================================================
-- Pairs with the pg `/workspace:ro` mount added to docker-compose.yaml.
-- read_workspace_doc() reads document text from the repo via pg_read_file,
-- so start_panel_redline (R.4) can inject it into a panel model's prompt
-- WITHOUT the model ever touching the filesystem. This is what kills the
-- original failure mode (fs-read allow-list + fs_search budget-burn).
--
-- SECURITY (D-RL2): the pg superuser can now read the whole repo (ro). The
-- real boundary is in this function, not the mount:
--   - repo-relative only, no absolute paths, no ".." traversal
--   - doc extensions ONLY (.md/.markdown/.txt/.mdx) — you cannot read .env,
--     a .key, a .pem, or any binary; secrets simply don't match
--   - explicit deny for .env*, .git/, and obvious secret name patterns
-- A crafted redline call therefore cannot exfiltrate a secret into an
-- external model's prompt. The models still get only injected text.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. _doc_path_allowed — the gate predicate.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards._doc_path_allowed(p_rel text)
RETURNS boolean LANGUAGE sql IMMUTABLE AS $$
    SELECT p_rel IS NOT NULL
       AND p_rel <> ''
       AND p_rel !~ '^/'                          -- not absolute
       AND p_rel !~ '(^|/)\.\.(/|$)'              -- no ".." traversal
       AND p_rel ~* '\.(md|markdown|txt|mdx)$'    -- document extensions ONLY
       AND p_rel !~* '(^|/)\.env'                 -- no .env / .env.*
       AND p_rel !~* '(^|/)\.git(/|$)'            -- no .git internals
       AND p_rel !~* '(secret|password|\.key$|\.pem$|id_rsa)';  -- obvious secrets
$$;

COMMENT ON FUNCTION stewards._doc_path_allowed(text) IS
'R.2: true if a repo-relative path is safe for read_workspace_doc — a document file (.md/.markdown/.txt/.mdx), repo-relative, no traversal, not a secret. The doc-extension requirement is the strong guard: .env / .key / binaries never match.';


-- ---------------------------------------------------------------------
-- 2. read_workspace_doc(path_or_glob) -> rows of (rel_path, content)
-- ---------------------------------------------------------------------
-- A single file, or a single-directory glob (e.g.
-- 'projects/scripture-book/src/chapters/*.md'). Recursive '**' is not
-- supported in v1 (the book's chapters live in one directory). Reads via
-- pg_read_file from the /workspace ro mount.
CREATE OR REPLACE FUNCTION stewards.read_workspace_doc(p_path_or_glob text)
RETURNS TABLE(rel_path text, content text)
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_root text := '/workspace';
    v_dir  text;
    v_pat  text;
    v_name text;
    v_rel  text;
BEGIN
    IF p_path_or_glob IS NULL OR p_path_or_glob = '' THEN
        RAISE EXCEPTION 'read_workspace_doc: path is required';
    END IF;
    IF p_path_or_glob ~ '^/' OR p_path_or_glob ~ '(^|/)\.\.(/|$)' THEN
        RAISE EXCEPTION 'read_workspace_doc: must be repo-relative with no "..": %', p_path_or_glob;
    END IF;

    IF position('*' IN p_path_or_glob) > 0 THEN
        -- Single-directory glob: dir + filename-pattern.
        IF position('/' IN p_path_or_glob) = 0 THEN
            RAISE EXCEPTION 'read_workspace_doc: bare glob not allowed — qualify with a directory: %', p_path_or_glob;
        END IF;
        v_dir := regexp_replace(p_path_or_glob, '/[^/]*$', '');
        v_pat := regexp_replace(p_path_or_glob, '^.*/', '');
        IF position('*' IN v_dir) > 0 THEN
            RAISE EXCEPTION 'read_workspace_doc: glob only supported in the filename, not the directory: %', p_path_or_glob;
        END IF;
        FOR v_name IN SELECT f FROM pg_ls_dir(v_root || '/' || v_dir) AS f ORDER BY f
        LOOP
            v_rel := v_dir || '/' || v_name;
            IF v_name LIKE replace(v_pat, '*', '%')
               AND stewards._doc_path_allowed(v_rel) THEN
                rel_path := v_rel;
                content  := pg_read_file(v_root || '/' || v_rel);
                RETURN NEXT;
            END IF;
        END LOOP;
    ELSE
        IF NOT stewards._doc_path_allowed(p_path_or_glob) THEN
            RAISE EXCEPTION 'read_workspace_doc: path not allowed — must be a .md/.markdown/.txt/.mdx document, not a secret or non-doc: %', p_path_or_glob;
        END IF;
        rel_path := p_path_or_glob;
        content  := pg_read_file(v_root || '/' || p_path_or_glob);
        RETURN NEXT;
    END IF;
END;
$func$;

COMMENT ON FUNCTION stewards.read_workspace_doc(text) IS
'R.2: read document text from the /workspace ro mount via pg_read_file, for server-side injection into a redline panel (R.4). Accepts a single file or a single-dir filename glob (e.g. .../chapters/*.md). Gated by _doc_path_allowed — docs only, no secrets, no traversal. The model never touches fs; only this function reads.';


-- =====================================================================
-- Acceptance (R.2):
--   1. SELECT count(*), sum(length(content))>0 FROM
--        read_workspace_doc('projects/scripture-book/src/chapters/00_frontmatter.md'); → 1, true
--   2. SELECT count(*) FROM
--        read_workspace_doc('projects/scripture-book/src/chapters/*.md'); → N (all chapters)
--   3. read_workspace_doc('projects/pg-ai-stewards/extension/.env') → EXCEPTION (not a doc)
--   4. read_workspace_doc('../../../etc/passwd') → EXCEPTION (traversal)
--   5. _doc_path_allowed('a/.env') = false; _doc_path_allowed('x/foo.md') = true
-- =====================================================================
