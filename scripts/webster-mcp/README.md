# Webster MCP Server

An MCP (Model Context Protocol) server that provides access to **Noah Webster's 1828 American Dictionary**, **Webster's Revised Unabridged Dictionary (1913)**, and the **Free Dictionary API** for modern definitions.

## Purpose

This server is particularly useful for scripture study. The Webster 1828 dictionary was compiled in the same era as the King James Bible's language tradition and the early Latter-day Saint scriptures, providing insight into the original meanings of words.

The three-dictionary approach allows comparing definitions across time — 1828 → 1913 → today — revealing how word meanings have shifted.

## Data Provenance (read this — it matters)

**Webster 1828** (`data/webster1828.json.gz`, 63,280 words): Noah Webster's *American Dictionary of the English Language* (1828, public domain). Text chain: the [Ellen G. White Estate Archives full-text preservation](https://archive.org/details/noah-websters-1828-dictionary-ellen-g-white-estate) → [kayson-argyle/websters_1828](https://github.com/kayson-argyle/websters_1828) (raw text + parsing pipeline, built for KJV/LDS Standard Works study — thank you!) → our converter `tools/convert_1828.py` (grouping, OCR cleanup, scripture-ref fixes). Verified against [webstersdictionary1828.com](https://webstersdictionary1828.com) by anachronism probes and word-by-word text comparison.

**Webster 1913** (`data/webster1913.json.gz`, 98,828 words): *Webster's Revised Unabridged Dictionary* (1913), via Project Gutenberg → [ssvivian/WebstersDictionary](https://github.com/ssvivian/WebstersDictionary) (MIT). A fine general historical dictionary — but it is **not** the 1828.

> **History note (2026-06-09):** from 2026-02-04 to 2026-06-09 this server served the 1913 text *labeled as 1828*. The ssvivian data never claimed to be 1828 — we assumed the edition without verifying it (the 1913 defines "telephone"; the tell was always there). The incident, forensics, and remediation are documented in `.spec/proposals/webster-1828-data-integrity.md`. The durable lesson: **verify the edition of a source, not just the quote.**

## Tools

### `define` (Recommended)
Look up a word in Webster 1828, Webster 1913, AND the modern dictionary. Three points in time, side by side.

```json
{ "word": "charity" }
```

### `webster_define`
Look up a word in the genuine Webster 1828 dictionary only. Falls back through
the 1828 spelling-variant map (`allege` → **ALLEDGE**, `zinc` → **ZINK**; see
`data/variants1828.json`) and archaic-suffix stemming (`sleepeth` → **SLEEP**),
always labeling a non-exact match — the 1828 genuinely lacks some modern
headwords (e.g. *naughty*, *pestilence*: absent from three independent
digitizations).

```json
{ "word": "charity" }
```

### `webster1913_define`
Look up a word in Webster's Revised Unabridged (1913) only.

```json
{ "word": "telephone" }
```

### `modern_define`
Look up a word in the modern dictionary (Free Dictionary API).

```json
{ "word": "charity" }
```

### `webster_search`
Search for words by pattern (prefix, contains). Optional `edition`: `"1828"` (default) or `"1913"`.

```json
{ "query": "char", "max_results": 20, "edition": "1828" }
```

### `webster_search_definitions`
Find words whose definitions contain specific text. Optional `edition`: `"1828"` (default) or `"1913"`.

```json
{ "query": "love", "max_results": 10 }
```

## Installation

### Build from source

```bash
cd scripts/webster-mcp
go build -o webster-mcp.exe ./cmd/webster-mcp
```

### Dictionary data

Both dictionaries ship as gzip-compressed JSON in `data/`:

- `webster1828.json.gz` (~4.6 MB) — genuine 1828
- `webster1913.json.gz` (~8 MB) — 1913 Revised Unabridged

The server decompresses on load. The 1913 file is auto-discovered as a sibling of the 1828 file; point `-dict1913` somewhere else to override.

### Regenerating the 1828 data

```bash
git clone https://github.com/kayson-argyle/websters_1828 <clone-dir>
# PYTHONUTF8=1 matters on Windows: upstream scripts open files without an encoding
PYTHONUTF8=1 python tools/convert_1828.py --src <clone-dir> --out data/webster1828.json.gz --report report.txt
```

The converter rejects OCR-junk headwords, strips U+FFFD characters, fixes invalid numbered-book scripture references ("7 Corinthians" → "1 Corinthians" — a pervasive 1→7 OCR error, 414 instances), and logs every change to the report.

### OCR repair tooling (2026-06-12)

The EGW-derived text carries residual OCR damage; three tools in `tools/` find
and fix it, every change ledgered:

- `scan_1828.py` — detection passes: near-empty senses, encoding artifacts,
  run-together words, junk fragments, **scripture references validated against
  the KJV canon** (via strongs-concordance-mcp's verse data), doubled phrases,
  missing See-reference targets.
- `repair_1828.py` — mechanical repair: fuzzy book-name fixes ("Wark" → Mark),
  1↔7 digit-swap candidates **disambiguated by KJV verse-text overlap**
  ("Corinthians 8:12" → the candidate whose verse text matches the quoted
  context), junction restoration, fragment handling, plus an overlay of
  entries re-transcribed from webstersdictionary1828.com where our text was
  destroyed.
- `fetch_site_entries.py` — polite (1.2s, identified UA, cached) fetcher for
  specific webstersdictionary1828.com entries used by the overlay.

Known faithful-but-odd readings are left alone deliberately: where all three
digitizations agree (e.g. *fine* sense 3 "Thin; keep; smoothly sharp"), the
text stands until checked against the 1828 facsimile — correcting toward
modern expectation is how mislabeling happens.

## Usage

### MCP registration (`.mcp.json`)

```json
{
  "servers": {
    "webster": {
      "command": "c:/path/to/webster-mcp.exe",
      "args": ["-dict", "c:/path/to/data/webster1828.json.gz"]
    }
  }
}
```

### Command Line

```bash
# Show dictionary statistics (both editions)
./webster-mcp.exe -stats

# Start MCP server (stdio)
./webster-mcp.exe

# Specify dictionary paths explicitly
./webster-mcp.exe -dict data/webster1828.json.gz -dict1913 data/webster1913.json.gz
```

## Example Output

Looking up "charity" with `webster_define`:

```markdown
**CHARITY** (noun)

**Definitions:**
1. In a general sense, love, benevolence, good will; that disposition of heart which inclines men to think favorably of their fellow men, and to do them good. In a theological sense, it includes supreme love to God, and universal good will to men. 1 Corinthians 8:1; Colossians 3:14; 1 Timothy 1:5.
2. In a more particular sense, love, kindness, affection, tenderness, springing from natural relations; as the charities of father, son and brother.
...
```

## Scripture Study Tips

Words that have changed meaning since 1828:

| Word | 1828 Meaning | Modern Focus |
|------|--------------|--------------|
| **charity** | Pure love of Christ, benevolence | Giving to the poor |
| **virtue** | Moral excellence, power | Sexual purity only |
| **peculiar** | Special, belonging exclusively | Strange, odd |
| **suffer** | Allow, permit | Experience pain |
| **conversation** | Conduct, behavior | Verbal exchange |
| **prevent** | Go before, precede | Stop from happening |

Use `define` to see the full 1828 → 1913 → modern drift for any word.

## License

- Code: MIT License
- Webster 1828 content: public domain (1828 text; see provenance above)
- Webster 1913 content: Project Gutenberg License
- Modern definitions: Free Dictionary API (Creative Commons)
