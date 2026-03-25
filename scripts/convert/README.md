# Convert — Scripture JSON to Markdown

Downloads the LDS scriptures JSON dataset from [beandog/lds-scriptures](https://github.com/beandog/lds-scriptures) and converts all verses into organized markdown files — one file per chapter, grouped by volume and book.

This was used for initial setup. The [gospel-library downloader](../gospel-library/README.md) is the preferred method for getting content now, as it pulls from the Church's official API and includes conference talks, manuals, and study aids.

## Usage

```bash
go run main.go
```

Downloads the JSON, converts to markdown, and writes to `gospel-library/eng/scriptures/`.
