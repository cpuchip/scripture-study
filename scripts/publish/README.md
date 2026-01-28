# Publish Script

Converts local markdown study files with relative gospel-library links to public files with absolute Church website URLs, enabling you to share your study notes with working links to scriptures, conference talks, and manuals.

## Usage

Run from the workspace root:

```bash
# Build the script
cd scripts/publish
go build -o publish.exe ./cmd/main.go

# Run from workspace root
cd ../..
./scripts/publish/publish.exe
```

### Flags

- `-output <dir>` - Output directory (default: `public`)
- `-v` - Verbose output showing each file processed
- `-dry-run` - Show what would be done without making changes

### Example

```bash
# Preview what will be converted
./scripts/publish/publish.exe -dry-run -v

# Generate public files
./scripts/publish/publish.exe -v
```

## What It Does

1. **Scans** the `study/` and `lessons/` directories for markdown files
2. **Converts** relative links to `gospel-library/` into absolute Church URLs:
   - `../../gospel-library/eng/scriptures/pgp/moses/6.md` → `https://www.churchofjesuschrist.org/study/scriptures/pgp/moses/6?lang=eng`
3. **Extracts verse references** from link text for deep linking:
   - `[Moses 6:59-60](...)` → `...?lang=eng&id=p59-p60#p59`
4. **Outputs** to `public/{study,lessons}/` preserving the directory structure

## Link Conversion Examples

| Original Link | Converted URL |
|--------------|---------------|
| `[Moses 6:59](../../gospel-library/eng/scriptures/pgp/moses/6.md)` | `https://www.churchofjesuschrist.org/study/scriptures/pgp/moses/6?lang=eng&id=p59#p59` |
| `[Genesis 5](../../gospel-library/eng/scriptures/ot/gen/5.md)` | `https://www.churchofjesuschrist.org/study/scriptures/ot/gen/5?lang=eng` |
| `[Born Again](../../gospel-library/eng/general-conference/2008/04/born-again.md)` | `https://www.churchofjesuschrist.org/study/general-conference/2008/04/born-again?lang=eng` |
| `[D&C 93:36](../../gospel-library/eng/scriptures/dc-testament/dc/93.md)` | `https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p36#p36` |

## Supported Link Types

- ✅ Scripture references (all standard works)
- ✅ General Conference talks
- ✅ Manuals and other curriculum
- ✅ Verse ranges (e.g., `Moses 6:59-60` → `p59-p60`)
- ✅ Single verses (e.g., `John 3:16` → `p16`)
- ❌ Internal links (kept as-is for local navigation)
- ❌ External links (kept as-is)

## Output Structure

```
public/
├── study/
│   ├── cfm/
│   │   └── 20260126-teach-these-things-freely.md
│   ├── creation.md
│   └── talks/
│       └── 202510-24brown.md
└── lessons/
    └── teaching-in-the-saviors-way/
        └── yw/
            └── 01_focus-on-christ.md
```
