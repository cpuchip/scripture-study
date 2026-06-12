#!/usr/bin/env python3
"""Fetch specific entries from webstersdictionary1828.com (the verification
authority) for OCR repair, politely (1.2s between requests, identified UA,
local cache so re-runs never refetch).

Page structure (learned 2026-06-12):
  <h3 class="dictionaryhead">Word</h3> ... <div><p><strong>HEADWORD</strong>,
  <em>pos</em> first sense...</p><p>1. second sense...</p>...</div>
A page may carry several headwords (the target plus following words); only
blocks whose <strong> headword matches the target (apostrophes stripped) are
taken. Scripture links become plain text. A page with no dictionaryhead block
or no matching headword records {"absent": true} — webstersdictionary1828.com
genuinely lacks some 1828 entries.

Usage:
  python fetch_site_entries.py --words words.txt --cache site-cache.json --out site-entries.json
"""

import argparse
import json
import os
import re
import time
import urllib.request

UA = "scripture-study-repair/1.0 (one-time dictionary OCR repair; contact cpuchip@gmail.com)"


def fetch(word):
    url = f"https://webstersdictionary1828.com/Dictionary/{word.lower()}"
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=30) as r:
        return r.read().decode("utf-8", errors="replace")


def strip_tags(html):
    t = re.sub(r"<[^>]+>", " ", html)
    t = t.replace("&#39;", "'").replace("&amp;", "&").replace("&quot;", '"')
    t = t.replace("&nbsp;", " ")
    return re.sub(r"\s+", " ", t).strip()


def norm_head(s):
    return re.sub(r"[^A-Z\-]", "", s.upper())


def parse_entry(html, target):
    """Extract the target word's entry blocks from a page."""
    i = html.find('class="dictionaryhead"')
    if i == -1:
        return None
    j = html.find('d-md-none', i)
    section = html[i:j if j > -1 else None]
    # blocks start at <p><strong>HEADWORD</strong>
    blocks = re.split(r"(?=<p><strong>)", section)
    target_n = norm_head(target)
    entries = []
    for b in blocks:
        m = re.match(r"<p><strong>([^<]+)</strong>(.*)", b, re.S)
        if not m:
            continue
        head = norm_head(m.group(1))
        if head != target_n:
            continue
        paras = re.findall(r"<p>(.*?)</p>", "<p>" + m.group(2), re.S)
        # first para: ", <em>pos</em> first sense text"
        first = strip_tags(paras[0]) if paras else ""
        pos = ""
        pm = re.match(r"<p>\s*,?\s*<em>([^<]+)</em>(.*)", "<p>" + (paras[0] if paras else ""), re.S)
        if pm:
            pos = pm.group(1).strip()
            first = strip_tags(pm.group(2))
        else:
            first = re.sub(r"^[,.\s]+", "", first)
        defs = []
        if first:
            defs.append(first)
        for p in paras[1:]:
            t = strip_tags(p)
            if t:
                defs.append(t)
        if defs:
            entries.append({"pos": pos, "definitions": defs})
    if not entries:
        return None
    return entries


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--words", required=True, help="file with one word per line")
    p.add_argument("--cache", required=True)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    with open(args.words, encoding="utf-8") as f:
        words = [w.strip().upper() for w in f if w.strip()]

    cache = {}
    if os.path.exists(args.cache):
        with open(args.cache, encoding="utf-8") as f:
            cache = json.load(f)

    out = {}
    fetched = 0
    for w in words:
        if w not in cache:
            try:
                cache[w] = fetch(w)
            except Exception as exc:  # noqa: BLE001 - record and continue
                cache[w] = f"__ERROR__ {exc}"
            fetched += 1
            if fetched % 25 == 0:
                print(f"  fetched {fetched}…")
                with open(args.cache, "w", encoding="utf-8") as f:
                    json.dump(cache, f)
            time.sleep(1.2)
        html = cache[w]
        if html.startswith("__ERROR__"):
            out[w] = {"absent": True, "error": html[:120]}
            continue
        entries = parse_entry(html, w)
        if entries:
            out[w] = {"entries": entries, "source": "webstersdictionary1828.com"}
        else:
            out[w] = {"absent": True}

    with open(args.cache, "w", encoding="utf-8") as f:
        json.dump(cache, f)
    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(out, f, ensure_ascii=False, indent=1)

    got = sum(1 for v in out.values() if "entries" in v)
    print(f"{len(words)} words requested; {got} entries parsed; {len(words) - got} absent/failed")
    print(f"-> {args.out}")


if __name__ == "__main__":
    main()
