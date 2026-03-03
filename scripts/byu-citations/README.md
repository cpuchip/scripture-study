# BYU Scripture Citation Index — MCP Server

An MCP server (and CLI tool) that queries the [BYU Scripture Citation Index](https://scriptures.byu.edu/) to find which General Conference talks, Journal of Discourses entries, and other sources cite a particular verse of scripture.

## Usage

### As MCP Server

The server starts automatically via VS Code MCP configuration. It provides three tools:

- **`byu_citations`** — Look up citations for a scripture reference
- **`byu_citations_bulk`** — Look up multiple references at once
- **`byu_citations_books`** — List all supported books and their IDs

### As CLI

```bash
# Single verse
byu-citations lookup "D&C 113:6"
byu-citations lookup "3 Nephi 21:10"
byu-citations lookup "Alma 32:21"

# Abbreviations work
byu-citations lookup "Isa 53:5"
byu-citations lookup "Matt 5:48"
byu-citations lookup "Hel 5:12"

# Full names work too
byu-citations lookup "Doctrine and Covenants 93:36"
```

## What It Returns

For each cited verse, you get:
- **Speaker** name
- **Talk title**
- **Reference** (year-season:page or JD volume:page)
- Internal BYU talk and reference IDs

## Supported Books

All standard works:
- **Old Testament**: Genesis through Malachi (IDs 101-139)
- **New Testament**: Matthew through Revelation (IDs 140-166)
- **Book of Mormon**: 1 Nephi through Moroni (IDs 205-219)
- **D&C**: Sections (302), Official Declarations (303)
- **Pearl of Great Price**: Moses, Abraham, Facsimile, JS-M, JS-H, Articles of Faith (IDs 401-406)

## API

This tool wraps the BYU Scripture Citation Index AJAX API:
```
https://scriptures.byu.edu/citation_index/citation_ajax/{speaker}/{startYear}/{endYear}/{source}/{sort}/{filter}/{bookId}/{chapter}?verses={verse}
```

Default parameters: `Any/1830/2026/all/s/f/{bookId}/{chapter}?verses={verse}`

## Building

```bash
cd scripts/byu-citations
go build -o byu-citations.exe ./cmd/byu-citations/
```
