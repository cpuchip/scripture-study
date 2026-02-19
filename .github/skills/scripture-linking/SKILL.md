---
name: scripture-linking
description: "Format scripture and conference talk links using workspace-relative paths. Covers path conventions, link examples, and file verification. Load when creating or editing study documents, lessons, talks, evaluations, or any markdown that references scriptures."
user-invokable: false
---

# Scripture & Talk Linking

## Path Conventions

- **Lowercase with hyphens:** `1-ne`, `js-h`, `w-of-m`
- **Chapter files without leading zeros:** `1.md`, `138.md`
- **D&C path:** `dc-testament/dc/` (not just `dc/`)
- **Pearl of Great Price:** `pgp/` → `moses/`, `abr/`, `js-m/`, `js-h/`

## Link Format

### Scriptures
```markdown
[1 Nephi 3:7](../gospel-library/eng/scriptures/bofm/1-ne/3.md)
[Moses 3:5](../gospel-library/eng/scriptures/pgp/moses/3.md)
[D&C 93:36](../gospel-library/eng/scriptures/dc-testament/dc/93.md)
[Abraham 4:18](../gospel-library/eng/scriptures/pgp/abr/4.md)
[Matthew 5:48](../gospel-library/eng/scriptures/nt/matt/5.md)
[Isaiah 53:3](../gospel-library/eng/scriptures/ot/isa/53.md)
```

### Study Aids
```markdown
[TG Faith](../gospel-library/eng/scriptures/tg/faith.md)
[BD Aaronic Priesthood](../gospel-library/eng/scriptures/bd/aaronic-priesthood.md)
[JST Genesis 9:21-25](../gospel-library/eng/scriptures/jst/jst-gen/9.md)
```

### Conference Talks
Always link to the **specific talk file**, never a conference directory.
```markdown
[President Nelson, April 2025](../gospel-library/eng/general-conference/2025/04/57nelson.md)
[Elder Eyring, October 2024](../gospel-library/eng/general-conference/2024/10/25eyring.md)
```

### Manuals
```markdown
[Come, Follow Me: Genesis 1-2](../gospel-library/eng/manual/come-follow-me-for-home-and-church-old-testament-2026/01.md)
```

## Verification Rules

1. **Verify files exist** before linking — use `file_search` or `list_dir` to confirm the path is valid
2. **Never link to a directory** — always link to the specific file (chapter, talk, lesson)
3. **Never claim a file doesn't exist** without checking — the gospel-library is large and paths can surprise you
4. **Prefer local copies** over external website links — reference cached files in `/gospel-library/` over linking to churchofjesuschrist.org

## Common Path Patterns

| Volume | Base Path |
|--------|-----------|
| Old Testament | `gospel-library/eng/scriptures/ot/{book}/{chapter}.md` |
| New Testament | `gospel-library/eng/scriptures/nt/{book}/{chapter}.md` |
| Book of Mormon | `gospel-library/eng/scriptures/bofm/{book}/{chapter}.md` |
| D&C | `gospel-library/eng/scriptures/dc-testament/dc/{section}.md` |
| Pearl of Great Price | `gospel-library/eng/scriptures/pgp/{book}/{chapter}.md` |
| Topical Guide | `gospel-library/eng/scriptures/tg/{topic}.md` |
| Bible Dictionary | `gospel-library/eng/scriptures/bd/{topic}.md` |
| Guide to the Scriptures | `gospel-library/eng/scriptures/gs/{topic}.md` |
| Conference Talks | `gospel-library/eng/general-conference/{year}/{month}/{filename}.md` |
| Manuals | `gospel-library/eng/manual/{manual-name}/{lesson}.md` |
