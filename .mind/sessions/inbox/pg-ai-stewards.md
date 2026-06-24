## 📬 2026-06-16 (from general-workspace) — proposal: let the digester pipelines READ our repos — OPEN (needs council)

**Michael's ask:** give the ai/book/video digester pipelines the ability to *read the
work we're doing here* — a container with our repos checked out — so a digester can
compare what *it* produced against *our* studies and surface what to learn / incorporate.

**Motivation on disk:** the playlist digester digested the Euclid video the same week the
general lane wrote a human study of the *same* video — neither knows the other exists. A
"cross-reference our corpus" stage turns the digesters' §6 ("what could we do with this")
into "here's how this compares to what we've done, and what's worth folding in."

**~90% there:** the substrate ships read-only fs-read; the gap is making our repos visible
to the digester container. (a) read-only bind-mount scripture-study / scripture-book /
pg-ai-stewards-**oss** (NOT the private substrate repo with keys); or (b) a git-clone step
like code-pr. New tools-on read-only "cross-reference our corpus" stage. Caveats:
read-only always; mind secrets; gitignored content (gospel-library, /books, /yt) won't be
in a clean clone. **New standing capability → dominion_in_council: ratify before building.**
Pairs with book-digester.md §6 + study-pipeline.md. **Adjacent to the digester-steward
(curator) — a presiding curator that can read our corpus could pick books/videos that
fill gaps in what we've already studied.**

— filed by general-workspace; NOT yet acted — the next council item when Michael wants it.

<!-- RESOLVED + cleared 2026-06-24: the doc-extract sandbox shipped (P3e/f). The
"cross-reference our corpus" need is met by `doc_import_corpus` — zip a repo (a
repo is a folder), import it into the searchable docs pool tagged by project,
then doc_search it; all through the same hardened no-network extract lane. A
live read-only `git clone` into the sandbox (no zip step) is the noted future
enhancement (docs/rich-documents.md §deferred). See
.spec/journal/2026-06-24-rich-docs-p3-doc-extract.md. -->

<!-- cleared 2026-06-16: storytelling-craft-digest (done) + stuck-research-write diagnosis (done) -->
