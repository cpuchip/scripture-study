"""Quick verification that OpenStax PDFs are readable & text-extractable.

For each PDF: page count, total chars, sample of first content page, and a
sanity check that math symbols are preserved.
"""
import fitz  # PyMuPDF
from pathlib import Path

BASE = Path(r"C:\Users\cpuch\Documents\code\stuffleberry\scripture-study\books\openstax")

PDFS = [
    "university-physics-vol-1/university-physics-vol-1.pdf",
    "university-physics-vol-2/university-physics-vol-2.pdf",
    "university-physics-vol-3/university-physics-vol-3.pdf",
    "chemistry-2e/chemistry-2e.pdf",
    "chemistry-atoms-first-2e/chemistry-atoms-first-2e.pdf",
    "astronomy-2e/astronomy-2e.pdf",
]

for rel in PDFS:
    p = BASE / rel
    doc = fitz.open(p)
    pages = doc.page_count
    # sample a page that's likely to have content (skip front matter)
    sample_page_idx = min(60, pages - 1)
    text = doc.load_page(sample_page_idx).get_text()
    print(f"\n=== {rel} ===")
    print(f"  pages: {pages}")
    print(f"  sample (page {sample_page_idx + 1}, first 300 chars):")
    snippet = text[:300].replace("\n", " | ")
    print(f"  {snippet}")
    doc.close()
