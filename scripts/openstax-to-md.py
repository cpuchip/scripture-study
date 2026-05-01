"""OpenStax HTML → Markdown converter.

Fetches a single OpenStax book page (e.g.
https://openstax.org/books/university-physics-volume-3/pages/1-1-the-propagation-of-light)
and converts it to a clean Markdown file with:
  - Headings preserved
  - MathML converted to LaTeX ($...$ inline, $$...$$ display)
  - Images downloaded locally and referenced
  - data-type sections (note, example, exercise) styled as blockquotes
  - Links preserved
  - Source URL + license + fetch date in frontmatter

Usage:
    python openstax-to-md.py <book-slug> <page-slug> [--out-dir books/openstax/<slug>/chapters]

Example:
    python openstax-to-md.py university-physics-volume-3 1-1-the-propagation-of-light

Per OpenStax CC BY-NC-SA 4.0 license, derivative works must be CC BY-NC-SA and attribute
the original. Output frontmatter includes the required attribution.
"""

from __future__ import annotations

import argparse
import re
import sys
import time
from pathlib import Path
from typing import Optional

import requests
from bs4 import BeautifulSoup, NavigableString, Tag
from mathml_to_latex.converter import MathMLToLaTeX

REPO_ROOT = Path(__file__).resolve().parent.parent
BOOKS_OPENSTAX = REPO_ROOT / "books" / "openstax"

USER_AGENT = "Mozilla/5.0 (scripture-study, personal educational use)"
HEADERS = {"User-Agent": USER_AGENT}

MML = MathMLToLaTeX()

# ---------------------------------------------------------------------------
# Math handling
# ---------------------------------------------------------------------------

def convert_math(el: Tag) -> str:
    """Convert a <math> element to LaTeX, wrapped per display attribute.

    OpenStax wraps math in <semantics> with both presentation MathML and an
    <annotation-xml> content-MathML duplicate. We must strip the annotations
    before conversion or every formula gets duplicated.
    """
    # Work on a copy so we don't mutate the soup
    from copy import copy
    el2 = copy(el)
    for ann in el2.find_all(["annotation", "annotation-xml"]):
        ann.decompose()
    raw = str(el2)
    try:
        latex = MML.convert(raw).strip()
    except Exception as e:  # pragma: no cover
        return f"<!-- math conversion failed: {e} -->"
    display = el.get("display", "inline")
    if display == "block":
        return f"\n\n$$\n{latex}\n$$\n\n"
    return f"${latex}$"


# ---------------------------------------------------------------------------
# Image handling
# ---------------------------------------------------------------------------

def download_image(src: str, img_dir: Path, timeout: int = 30) -> Optional[str]:
    """Download an image and return a relative path. Returns None on failure."""
    img_dir.mkdir(parents=True, exist_ok=True)
    # Use last path segment as filename, strip query
    name = src.split("?")[0].rsplit("/", 1)[-1]
    if not name or "." not in name:
        # Use a hash-based fallback
        import hashlib
        name = hashlib.sha1(src.encode()).hexdigest()[:12] + ".png"
    dest = img_dir / name
    if dest.exists():
        return f"images/{name}"
    try:
        r = requests.get(src, headers=HEADERS, timeout=timeout)
        r.raise_for_status()
        dest.write_bytes(r.content)
        return f"images/{name}"
    except Exception as e:
        print(f"  image fail: {src} -> {e}", file=sys.stderr)
        return None


# ---------------------------------------------------------------------------
# Element-to-markdown conversion (recursive)
# ---------------------------------------------------------------------------

def text_of(node) -> str:
    """Get text content, normalizing whitespace."""
    if isinstance(node, NavigableString):
        return str(node)
    return node.get_text()


def convert_element(el, img_dir: Path, depth: int = 0) -> str:  # noqa: C901
    """Recursively convert an HTML element to Markdown."""
    if isinstance(el, NavigableString):
        return str(el)

    if not isinstance(el, Tag):
        return ""

    name = el.name.lower()
    data_type = el.get("data-type", "")

    # Skip nav, script, style, etc.
    if name in {"script", "style", "nav", "header", "footer", "noscript"}:
        return ""

    # Math: handle at the <math> level (do not recurse into MathML)
    if name == "math":
        return convert_math(el)

    # Skip the redundant title at top of page (we add our own from frontmatter)
    if data_type == "document-title":
        return ""

    # Headings
    if name in {"h1", "h2", "h3", "h4", "h5", "h6"}:
        level = int(name[1])
        # OpenStax sometimes wraps headings with section number prefix in <span>
        text = el.get_text(" ", strip=True)
        # Insert a space between section number and title if missing
        text = re.sub(r"^(\d+(?:\.\d+)*)([A-Z])", r"\1 \2", text)
        return f"\n\n{'#' * level} {text}\n\n"

    # Paragraphs
    if name == "p":
        inner = "".join(convert_element(c, img_dir, depth) for c in el.children).strip()
        if not inner:
            return ""
        return f"\n\n{inner}\n\n"

    # Lists
    if name == "ul":
        items = [convert_element(c, img_dir, depth) for c in el.children if isinstance(c, Tag) and c.name == "li"]
        return "\n\n" + "\n".join(f"- {it.strip()}" for it in items if it.strip()) + "\n\n"
    if name == "ol":
        items = [convert_element(c, img_dir, depth) for c in el.children if isinstance(c, Tag) and c.name == "li"]
        return "\n\n" + "\n".join(f"{i+1}. {it.strip()}" for i, it in enumerate(items) if it.strip()) + "\n\n"
    if name == "li":
        return "".join(convert_element(c, img_dir, depth) for c in el.children).strip()

    # Inline emphasis
    if name in {"strong", "b"}:
        return f"**{el.get_text()}**"
    if name in {"em", "i"}:
        return f"*{el.get_text()}*"
    if name == "code":
        return f"`{el.get_text()}`"
    if name == "br":
        return "  \n"

    # Links
    if name == "a":
        href = el.get("href", "")
        text = "".join(convert_element(c, img_dir, depth) for c in el.children).strip()
        if not href or href.startswith("#"):
            return text
        if href.startswith("/"):
            href = "https://openstax.org" + href
        return f"[{text}]({href})"

    # Images
    if name == "img":
        src = el.get("src", "")
        alt = el.get("alt", "").replace("\n", " ").strip()
        if not src:
            return ""
        if src.startswith("/"):
            src = "https://openstax.org" + src
        local = download_image(src, img_dir)
        if local:
            return f"\n\n![{alt}]({local})\n\n"
        return f"\n\n![{alt}]({src})\n\n"

    # Figures
    if name == "figure":
        inner = "".join(convert_element(c, img_dir, depth) for c in el.children)
        return f"\n\n{inner}\n\n"
    if name == "figcaption":
        text = "".join(convert_element(c, img_dir, depth) for c in el.children).strip()
        return f"\n\n*{text}*\n\n"

    # OpenStax data-type special blocks
    if data_type == "equation":
        # The <math> child will already be handled with display
        inner = "".join(convert_element(c, img_dir, depth) for c in el.children).strip()
        # Force display math regardless of inline marker
        if inner.startswith("$") and inner.endswith("$") and not inner.startswith("$$"):
            inner = "$$\n" + inner.strip("$") + "\n$$"
        return f"\n\n{inner}\n\n"

    if data_type in {"note", "abstract"}:
        title_el = el.find(attrs={"data-type": "title"})
        title_text = title_el.get_text(strip=True) if title_el else data_type.title()
        body_parts = []
        for c in el.children:
            if isinstance(c, Tag) and c.get("data-type") == "title":
                continue
            body_parts.append(convert_element(c, img_dir, depth))
        body = "".join(body_parts).strip()
        body_lines = "\n".join(f"> {line}" for line in body.splitlines())
        return f"\n\n> **{title_text}**\n>\n{body_lines}\n\n"

    if data_type == "example":
        title_el = el.find(attrs={"data-type": "title"})
        title_text = title_el.get_text(strip=True) if title_el else "Example"
        body_parts = []
        for c in el.children:
            if isinstance(c, Tag) and c.get("data-type") == "title":
                continue
            body_parts.append(convert_element(c, img_dir, depth))
        body = "".join(body_parts).strip()
        return f"\n\n---\n\n**{title_text}**\n\n{body}\n\n---\n\n"

    if data_type == "exercise":
        body = "".join(convert_element(c, img_dir, depth) for c in el.children).strip()
        return f"\n\n> **Exercise:**\n> {body}\n\n"

    if data_type == "term":
        return f"**{el.get_text()}**"

    if data_type == "solution":
        body = "".join(convert_element(c, img_dir, depth) for c in el.children).strip()
        return f"\n\n*Solution:* {body}\n\n"

    # Default: recurse into children
    return "".join(convert_element(c, img_dir, depth) for c in el.children)


def normalize_whitespace(s: str) -> str:
    """Collapse 3+ blank lines and trim."""
    s = re.sub(r"\n{3,}", "\n\n", s)
    s = re.sub(r"[ \t]+\n", "\n", s)
    return s.strip() + "\n"


# ---------------------------------------------------------------------------
# Top-level conversion
# ---------------------------------------------------------------------------

def convert_page(book_slug: str, page_slug: str, out_dir: Optional[Path] = None) -> Path:
    url = f"https://openstax.org/books/{book_slug}/pages/{page_slug}"
    print(f"Fetching: {url}")
    r = requests.get(url, headers=HEADERS, timeout=60)
    r.raise_for_status()
    # OpenStax serves text/html with NO charset; requests defaults to Latin-1
    # which mangles unicode (×, ≈, em-dash, smart quotes). Force UTF-8.
    r.encoding = "utf-8"

    soup = BeautifulSoup(r.text, "lxml")
    page_el = soup.select_one('[data-type="page"]')
    if page_el is None:
        page_el = soup.select_one("main") or soup
        print("  warning: [data-type=\"page\"] not found, falling back to <main>")

    title_el = page_el.find(attrs={"data-type": "document-title"}) or page_el.find("h1")
    title = title_el.get_text(" ", strip=True) if title_el else page_slug

    if out_dir is None:
        out_dir = BOOKS_OPENSTAX / book_slug / "chapters"
    out_dir.mkdir(parents=True, exist_ok=True)
    img_dir = out_dir / "images"

    body_md = convert_element(page_el, img_dir)
    body_md = normalize_whitespace(body_md)

    # Frontmatter and footer attribution
    fetched = time.strftime("%Y-%m-%d")
    frontmatter = (
        "---\n"
        f"book: {book_slug}\n"
        f"page: {page_slug}\n"
        f"title: \"{title.replace(chr(34), chr(39))}\"\n"
        f"source_url: {url}\n"
        f"fetched: {fetched}\n"
        f"license: CC BY-NC-SA 4.0 (OpenStax / Rice University)\n"
        "---\n\n"
    )
    footer = (
        "\n\n---\n\n"
        f"*Source: [OpenStax — {title}]({url}). "
        "Licensed under [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/). "
        "© Rice University.*\n"
    )

    out_path = out_dir / f"{page_slug}.md"
    out_path.write_text(frontmatter + body_md + footer, encoding="utf-8")
    print(f"  wrote: {out_path.relative_to(REPO_ROOT)}")
    return out_path


def main(argv=None):
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("book_slug", help="e.g. university-physics-volume-3")
    p.add_argument("page_slug", help="e.g. 1-1-the-propagation-of-light")
    p.add_argument("--out-dir", type=Path, default=None)
    args = p.parse_args(argv)
    convert_page(args.book_slug, args.page_slug, args.out_dir)


if __name__ == "__main__":
    main()
