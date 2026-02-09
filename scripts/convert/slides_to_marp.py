#!/usr/bin/env python3
"""
Convert raw slides.md to Marp presentation format.

Handles:
- Marp YAML front matter
- Slide heading extraction from first content line
- Tab-separated data → markdown tables
- Image insertion (image-only slides and overlays)
- Shared images (220_and_222.png, 227_and_237.png)
- Zero-width space and artifact cleanup
"""
import re
import os
import sys

# --- Configuration ---

BASE = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
INPUT = os.path.join(BASE, "study", "yt", "morganphilpot", "original", "slides.md")
OUTPUT = os.path.join(BASE, "study", "yt", "morganphilpot", "original", "slides_marp.md")
ASSETS_REL = "assets"  # relative to the output file

# Image file → slide numbers it appears on
IMAGE_MAP = {
    "161.png": [161],
    "162.png": [162],
    "182.png": [182],
    "220_and_222.png": [220, 222],
    "224.png": [224],
    "225.png": [225],
    "226.png": [226],
    "227_and_237.png": [227, 237],
    "330.png": [330],
    "331.png": [331],
    "332.png": [332],
    "335.png": [335],
    "350.png": [350],
    "351.png": [351],
}

# Reverse mapping: slide number → image filename
SLIDE_IMG = {}
for fname, nums in IMAGE_MAP.items():
    for n in nums:
        SLIDE_IMG[n] = fname

# Slides that exist ONLY as images (no text in the source file)
IMAGE_ONLY = {161, 162, 182, 330, 331, 332, 350}

FRONT_MATTER = """\
---
marp: true
theme: default
paginate: true
header: "Signs of the Times — 2025 Chandler Presentation"
footer: "Morgan Philpot"
style: |
  section {
    font-size: 22px;
  }
  section.title {
    text-align: center;
    font-size: 32px;
  }
  section.image-slide {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 0;
  }
  blockquote {
    border-left: 4px solid #2196F3;
    padding-left: 16px;
    font-style: italic;
  }
  table {
    font-size: 18px;
    width: 100%;
  }
  img[alt~="center"] {
    display: block;
    margin: 0 auto;
  }
---"""


# --- Helpers ---

def clean(text: str) -> str:
    """Remove common copy-paste artifacts."""
    text = text.replace("\u200b", "")   # zero-width space
    text = text.replace("\ufeff", "")   # BOM
    text = text.replace("\r\n", "\n")   # normalize line endings
    # Collapse 3+ blank lines into 2
    text = re.sub(r"\n{3,}", "\n\n", text)
    return text.strip()


def collapse_google_slides_tables(text: str) -> str:
    """
    Google Slides table exports produce this pattern:
        Cell1
        (blank)
        \\t
        (blank)
        Cell2
    Collapse into: Cell1\\tCell2
    """
    # Normalize tab-separator lines (lines with only whitespace/tabs and at least one tab)
    text = re.sub(r"\n[ \t]*\t[ \t]*\n", "\n\t\n", text)
    # Iteratively collapse: [content] \\n+ \\t \\n+ [content] → [content]\\t[content]
    prev = None
    while text != prev:
        prev = text
        text = re.sub(
            r"(\S[^\n]*?)\s*\n(?:\s*\n)*\t\n(?:\s*\n)*(\S)",
            r"\1\t\2",
            text,
        )
    return text


def tabs_to_table(text: str) -> str:
    """
    Convert runs of tab-separated lines into markdown tables.
    A 'run' is 2+ lines (possibly separated by blank lines) each containing tabs.
    """
    lines = text.split("\n")
    result = []
    buf: list[str] = []

    def flush():
        if len(buf) >= 2:
            result.extend(_make_table(buf))
        else:
            result.extend(buf)
        buf.clear()

    i = 0
    while i < len(lines):
        stripped = lines[i].strip()
        if "\t" in stripped and stripped:
            buf.append(stripped)
            i += 1
        elif not stripped and buf:
            # Blank line while accumulating — peek ahead for more tab rows
            j = i
            while j < len(lines) and not lines[j].strip():
                j += 1
            if j < len(lines) and "\t" in lines[j] and lines[j].strip():
                i = j          # skip blank lines, continue collecting
            else:
                flush()
                result.append(lines[i])
                i += 1
        else:
            flush()
            result.append(lines[i])
            i += 1
    flush()
    return "\n".join(result)


def _make_table(rows_raw: list[str]) -> list[str]:
    """Convert tab-separated rows to a markdown pipe table."""
    rows = [r.split("\t") for r in rows_raw]
    max_cols = max(len(r) for r in rows)
    # Pad short rows
    for r in rows:
        while len(r) < max_cols:
            r.append("")
    # Clean cells
    rows = [[c.strip() for c in r] for r in rows]
    # Build table
    out = []
    out.append("| " + " | ".join(rows[0]) + " |")
    out.append("| " + " | ".join("---" for _ in range(max_cols)) + " |")
    for row in rows[1:]:
        out.append("| " + " | ".join(row) + " |")
    out.append("")  # trailing blank
    return out


def extract_heading_and_body(content: str):
    """
    Split slide content into (heading_line, body).
    The first non-empty line becomes the heading.
    """
    lines = content.split("\n")
    heading_idx = None
    for i, line in enumerate(lines):
        if line.strip():
            heading_idx = i
            break
    if heading_idx is None:
        return None, ""
    heading = lines[heading_idx].strip()
    body = "\n".join(lines[heading_idx + 1:]).strip()
    return heading, body


# --- Parsing ---

def parse_slides(raw: str):
    """
    Split the raw file by '---' separators, then identify slide numbers.
    Returns a list of (slide_number | None, body_text).
    """
    # Use regex split to handle possible trailing whitespace around ---
    blocks = re.split(r"\n---\n", raw)
    slides = []
    for block in blocks:
        block = block.strip()
        if not block:
            continue
        m = re.match(r"^## Slide (\d+)\s*\n?(.*)", block, re.DOTALL)
        if m:
            num = int(m.group(1))
            body = m.group(2).strip()
            slides.append((num, body))
        else:
            # Title slide or non-numbered content
            slides.append((None, block))
    return slides


# --- Slide Formatting ---

def format_slide(num, body, img_path=None, img_only=False):
    """Build the content for a single Marp slide."""
    parts = []

    # Slide number comment
    if num is not None:
        parts.append(f"<!-- slide {num} -->")
        parts.append("")

    # Image-only slide
    if img_only:
        parts.append(f"![bg contain]({img_path})")
        return "\n".join(parts)

    body = clean(body)
    body = collapse_google_slides_tables(body)
    body = tabs_to_table(body)

    if not body:
        if img_path:
            parts.append(f"![bg contain]({img_path})")
        return "\n".join(parts)

    heading, rest = extract_heading_and_body(body)

    if heading is None:
        if img_path:
            parts.append(f"![bg contain]({img_path})")
        return "\n".join(parts)

    # Image as background-right for slides with both text and image
    if img_path:
        parts.append(f"![bg right:40%]({img_path})")
        parts.append("")

    # Format heading
    if heading.startswith("#"):
        parts.append(heading)
    else:
        parts.append(f"## {heading}")

    if rest:
        parts.append("")
        parts.append(rest)

    return "\n".join(parts)


# --- Main Assembly ---

def main():
    print(f"Reading: {INPUT}")
    with open(INPUT, "r", encoding="utf-8") as f:
        raw = f.read()

    slides = parse_slides(raw)
    print(f"Parsed {len(slides)} text blocks")

    # Collect slide numbers we have text for
    text_nums = {num for num, _ in slides if num is not None}

    # Build output parts
    out_parts = [FRONT_MATTER]

    prev_num = 0
    first_slide = True

    for num, body in slides:
        if num is not None:
            # Insert image-only slides that belong between prev_num and num
            for img_num in sorted(IMAGE_ONLY):
                if prev_num < img_num < num:
                    img_path = f"{ASSETS_REL}/{SLIDE_IMG[img_num]}"
                    out_parts.append("\n---\n")
                    out_parts.append(format_slide(img_num, "", img_path=img_path, img_only=True))
            prev_num = num

        # Determine image for this slide
        has_img = (num in SLIDE_IMG) and (num not in IMAGE_ONLY)
        img_path = f"{ASSETS_REL}/{SLIDE_IMG[num]}" if has_img else None

        # Slide separator
        if first_slide:
            out_parts.append("\n")
            first_slide = False
        else:
            out_parts.append("\n---\n")

        out_parts.append(format_slide(num, body, img_path=img_path))

    # Any remaining image-only slides after the last text slide
    for img_num in sorted(IMAGE_ONLY):
        if img_num > prev_num:
            img_path = f"{ASSETS_REL}/{SLIDE_IMG[img_num]}"
            out_parts.append("\n---\n")
            out_parts.append(format_slide(img_num, "", img_path=img_path, img_only=True))

    output = "\n".join(out_parts)
    # Final cleanup: collapse excessive blank lines
    output = re.sub(r"\n{4,}", "\n\n\n", output)

    with open(OUTPUT, "w", encoding="utf-8") as f:
        f.write(output)

    total = len(slides) + len(IMAGE_ONLY)
    print(f"Wrote {total} slides ({len(slides)} text + {len(IMAGE_ONLY)} image-only)")
    print(f"Output: {OUTPUT}")


if __name__ == "__main__":
    main()
