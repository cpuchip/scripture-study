# Webster MCP Server

An MCP (Model Context Protocol) server that provides access to the **Webster 1828 Dictionary** and the **Free Dictionary API** for modern definitions.

## Purpose

This server is particularly useful for scripture study, as the Webster 1828 dictionary was compiled during the same era as the King James Bible and early Latter-day Saint scriptures, providing insight into the original meanings of words.

The dual-dictionary approach allows comparing historical and modern definitions side-by-side, revealing how word meanings have shifted over time.

## Features

- **98,000+ word definitions** from Noah Webster's 1828 American Dictionary
- **Modern definitions** via the Free Dictionary API
- **Combined lookup** showing both historical and contemporary meanings
- **Search by word pattern** - find words containing a query
- **Search within definitions** - find words whose definitions mention a topic

## Tools

### `define` (Recommended)
Look up a word in both dictionaries. Returns historical and modern definitions side by side.

```json
{
  "word": "charity"
}
```

### `webster_define`
Look up a word in the Webster 1828 dictionary only.

```json
{
  "word": "charity"
}
```

### `modern_define`
Look up a word in the modern dictionary (Free Dictionary API).

```json
{
  "word": "charity"
}
```

### `webster_search`
Search for words by pattern (prefix, contains).

```json
{
  "query": "char",
  "max_results": 20
}
```

### `webster_search_definitions`
Find words whose definitions contain specific text.

```json
{
  "query": "love",
  "max_results": 10
}
```

## Installation

### Build from source

```bash
cd scripts/webster-mcp
go build -o webster-mcp.exe ./cmd/webster-mcp
```

### Download dictionary data

The Webster 1828 dictionary is included as a gzip-compressed file: `data/webster1828.json.gz` (~8 MB).

The server automatically decompresses the file on load. Both `.json` and `.json.gz` files are supported.

Source: https://github.com/ssvivian/WebstersDictionary (MIT License)

## Usage

### VS Code Configuration

Add to your `.vscode/mcp.json`:

```json
{
  "servers": {
    "webster": {
      "command": "c:/path/to/webster-mcp.exe",
      "args": ["-dict", "c:/path/to/webster1828.json.gz"]
    }
  }
}
```

Or use environment variable:

```json
{
  "servers": {
    "webster": {
      "command": "c:/path/to/webster-mcp.exe",
      "env": {
        "WEBSTER_DICT_PATH": "c:/path/to/webster1828.json.gz"
      }
      }
    }
  }
}
```

### Command Line

```bash
# Show dictionary statistics
./webster-mcp.exe -stats

# Start MCP server (stdio)
./webster-mcp.exe

# Specify custom dictionary path
./webster-mcp.exe -dict /path/to/webster1828.json
```

## Example Output

Looking up "charity":

```markdown
# Definitions for: charity

## Webster 1828 Dictionary
_Historical definitions from Noah Webster's 1828 dictionary, reflecting the language of scripture._

**CHARITY** (n.)
*Synonyms:* Love; benevolence; good will; affection; tenderness; beneficence; liberality; almsgiving.

**Definitions:**
1. Love; universal benevolence; good will.
2. Liberality in judging of men and their actions; a disposition which inclines men to put the best construction on the words and actions of others.
3. Liberality to the poor and the suffering...
...

---

## Modern Dictionary
_Contemporary definitions from the Free Dictionary API._

**charity** /ˈtʃærɪti/

**noun**
1. Provision of help or relief to the poor; almsgiving.
2. Something given to help the needy; alms.
3. An institution, organization, or fund established to help the needy.
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

## License

- Code: MIT License
- Webster 1828 Dictionary content: Project Gutenberg License
- Modern definitions: Free Dictionary API (Creative Commons)
