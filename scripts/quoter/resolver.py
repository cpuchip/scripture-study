#!/usr/bin/env python3
"""resolver — scripture reference -> canonical gospel-library path + relative link.

The shared spine of the study-tooling suite (see .spec/proposals/study-tooling.md):
the quoter uses it for link-gen, the linter uses the same map for link-validate.
Build it once; both consume it.

A reference like "Alma 5:14" or "Romans 5:3-5" or "D&C 19:18" resolves to the
gospel-library file (bofm/alma/5.md, nt/rom/5.md, dc-testament/dc/19.md) and, given
a target file, the correct RELATIVE markdown link (../ depth computed per target —
the thing that breaks moving a quote between files at different directory depths).
"""
import os, re

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
SCRIPTURES = os.path.join("gospel-library", "eng", "scriptures")

# (volume, folder, [name aliases]) — aliases are matched on a normalized key
# (lowercased, internal whitespace collapsed, trailing '.' dropped).
BOOKS = [
    # Old Testament
    ("ot", "gen", ["genesis", "gen"]),
    ("ot", "ex", ["exodus", "ex", "exod"]),
    ("ot", "lev", ["leviticus", "lev"]),
    ("ot", "num", ["numbers", "num"]),
    ("ot", "deut", ["deuteronomy", "deut", "deu"]),
    ("ot", "josh", ["joshua", "josh"]),
    ("ot", "judg", ["judges", "judg"]),
    ("ot", "ruth", ["ruth"]),
    ("ot", "1-sam", ["1 samuel", "1 sam", "i samuel"]),
    ("ot", "2-sam", ["2 samuel", "2 sam", "ii samuel"]),
    ("ot", "1-kgs", ["1 kings", "1 kgs", "i kings"]),
    ("ot", "2-kgs", ["2 kings", "2 kgs", "ii kings"]),
    ("ot", "1-chr", ["1 chronicles", "1 chr", "1 chron"]),
    ("ot", "2-chr", ["2 chronicles", "2 chr", "2 chron"]),
    ("ot", "ezra", ["ezra"]),
    ("ot", "neh", ["nehemiah", "neh"]),
    ("ot", "esth", ["esther", "esth"]),
    ("ot", "job", ["job"]),
    ("ot", "ps", ["psalms", "psalm", "ps", "psa"]),
    ("ot", "prov", ["proverbs", "prov", "pr"]),
    ("ot", "eccl", ["ecclesiastes", "eccl", "eccles"]),
    ("ot", "song", ["song of solomon", "song of songs", "song", "sos"]),
    ("ot", "isa", ["isaiah", "isa"]),
    ("ot", "jer", ["jeremiah", "jer"]),
    ("ot", "lam", ["lamentations", "lam"]),
    ("ot", "ezek", ["ezekiel", "ezek", "eze"]),
    ("ot", "dan", ["daniel", "dan"]),
    ("ot", "hosea", ["hosea", "hos"]),
    ("ot", "joel", ["joel"]),
    ("ot", "amos", ["amos"]),
    ("ot", "obad", ["obadiah", "obad"]),
    ("ot", "jonah", ["jonah"]),
    ("ot", "micah", ["micah", "mic"]),
    ("ot", "nahum", ["nahum", "nah"]),
    ("ot", "hab", ["habakkuk", "hab"]),
    ("ot", "zeph", ["zephaniah", "zeph"]),
    ("ot", "hag", ["haggai", "hag"]),
    ("ot", "zech", ["zechariah", "zech", "zec"]),
    ("ot", "mal", ["malachi", "mal"]),
    # New Testament
    ("nt", "matt", ["matthew", "matt", "mt"]),
    ("nt", "mark", ["mark", "mrk", "mk"]),
    ("nt", "luke", ["luke", "lk"]),
    ("nt", "john", ["john", "jhn", "jn"]),
    ("nt", "acts", ["acts"]),
    ("nt", "rom", ["romans", "rom"]),
    ("nt", "1-cor", ["1 corinthians", "1 cor"]),
    ("nt", "2-cor", ["2 corinthians", "2 cor"]),
    ("nt", "gal", ["galatians", "gal"]),
    ("nt", "eph", ["ephesians", "eph"]),
    ("nt", "philip", ["philippians", "philip", "phil", "php"]),
    ("nt", "col", ["colossians", "col"]),
    ("nt", "1-thes", ["1 thessalonians", "1 thes", "1 thess"]),
    ("nt", "2-thes", ["2 thessalonians", "2 thes", "2 thess"]),
    ("nt", "1-tim", ["1 timothy", "1 tim"]),
    ("nt", "2-tim", ["2 timothy", "2 tim"]),
    ("nt", "titus", ["titus"]),
    ("nt", "philem", ["philemon", "philem", "phlm"]),
    ("nt", "heb", ["hebrews", "heb"]),
    ("nt", "james", ["james", "jas"]),
    ("nt", "1-pet", ["1 peter", "1 pet"]),
    ("nt", "2-pet", ["2 peter", "2 pet"]),
    ("nt", "1-jn", ["1 john", "1 jn", "1 jhn"]),
    ("nt", "2-jn", ["2 john", "2 jn"]),
    ("nt", "3-jn", ["3 john", "3 jn"]),
    ("nt", "jude", ["jude"]),
    ("nt", "rev", ["revelation", "rev"]),
    # Book of Mormon
    ("bofm", "1-ne", ["1 nephi", "1 ne", "1 nep"]),
    ("bofm", "2-ne", ["2 nephi", "2 ne", "2 nep"]),
    ("bofm", "jacob", ["jacob"]),
    ("bofm", "enos", ["enos"]),
    ("bofm", "jarom", ["jarom"]),
    ("bofm", "omni", ["omni"]),
    ("bofm", "w-of-m", ["words of mormon", "w of m", "wom"]),
    ("bofm", "mosiah", ["mosiah"]),
    ("bofm", "alma", ["alma"]),
    ("bofm", "hel", ["helaman", "hel"]),
    ("bofm", "3-ne", ["3 nephi", "3 ne", "3 nep"]),
    ("bofm", "4-ne", ["4 nephi", "4 ne", "4 nep"]),
    ("bofm", "morm", ["mormon", "morm"]),
    ("bofm", "ether", ["ether"]),
    ("bofm", "moro", ["moroni", "moro", "mor"]),
    # Doctrine and Covenants
    ("dc-testament", "dc", ["doctrine and covenants", "d&c", "dc", "d and c"]),
    ("dc-testament", "od", ["official declaration", "od"]),
    # Pearl of Great Price
    ("pgp", "moses", ["moses"]),
    ("pgp", "abr", ["abraham", "abr"]),
    ("pgp", "js-m", ["joseph smith matthew", "joseph smith-matthew", "js-m", "jsm"]),
    ("pgp", "js-h", ["joseph smith history", "joseph smith-history", "js-h", "jsh"]),
    ("pgp", "a-of-f", ["articles of faith", "a of f", "aof"]),
]

_ALIAS = {}
for vol, folder, names in BOOKS:
    for n in names:
        _ALIAS[n] = (vol, folder)

def _nkey(s):
    s = s.lower().replace("—", "-").replace("–", "-").strip()
    s = re.sub(r"\.", "", s)
    s = re.sub(r"\s+", " ", s)
    return s.strip()

# "1 Nephi 11:21"  "Romans 5:3-5"  "D&C 19:18"  "Alma 5"  (chapter-only ok)
_REF = re.compile(
    r"^\s*(?P<book>(?:[1-4]\s+)?[A-Za-z&.—\- ]+?)\s+"
    r"(?P<chap>\d+)(?::(?P<v1>\d+)(?:[-–](?P<v2>\d+))?)?\s*$"
)

class RefError(ValueError):
    pass

def parse_ref(ref):
    """'Alma 5:14' -> dict(book_label, vol, folder, chap, v1, v2|None).
    v1 None means a whole-chapter ref (block use)."""
    m = _REF.match(ref)
    if not m:
        raise RefError(f"unparseable reference: {ref!r}")
    key = _nkey(m.group("book"))
    if key not in _ALIAS:
        # try collapsing 'i/ii/iii' style already handled; otherwise unknown
        raise RefError(f"unknown book: {m.group('book')!r} (in {ref!r})")
    vol, folder = _ALIAS[key]
    v1 = int(m.group("v1")) if m.group("v1") else None
    v2 = int(m.group("v2")) if m.group("v2") else None
    return {
        "label": _canonical_label(key, m.group("chap"), v1, v2),
        "vol": vol, "folder": folder,
        "chap": int(m.group("chap")), "v1": v1, "v2": v2,
    }

# Pretty book labels for the emitted link text (so "1 ne" -> "1 Nephi").
_LABEL = {folder: names[0].title().replace("And", "and") for vol, folder, names in BOOKS}
_LABEL["dc"] = "D&C"
_LABEL["a-of-f"] = "Articles of Faith"
_LABEL["js-h"] = "Joseph Smith—History"
_LABEL["js-m"] = "Joseph Smith—Matthew"
_LABEL["w-of-m"] = "Words of Mormon"
_LABEL["1-ne"] = "1 Nephi"; _LABEL["2-ne"] = "2 Nephi"
_LABEL["3-ne"] = "3 Nephi"; _LABEL["4-ne"] = "4 Nephi"
_LABEL["song"] = "Song of Solomon"

def _canonical_label(key, chap, v1, v2):
    vol, folder = _ALIAS[key]
    book = _LABEL.get(folder, key.title())
    if v1 is None:
        return f"{book} {chap}"
    if v2:
        return f"{book} {chap}:{v1}-{v2}"
    return f"{book} {chap}:{v1}"

def file_path(ref_or_parsed):
    """Absolute path to the chapter .md file for a ref."""
    p = ref_or_parsed if isinstance(ref_or_parsed, dict) else parse_ref(ref_or_parsed)
    return os.path.join(ROOT, SCRIPTURES, p["vol"], p["folder"], f"{p['chap']}.md")

def link(ref_or_parsed, target_file):
    """Relative markdown link path from `target_file` to the chapter file —
    re-based per target's directory depth (../ count). Forward slashes."""
    p = ref_or_parsed if isinstance(ref_or_parsed, dict) else parse_ref(ref_or_parsed)
    dest = file_path(p)
    start = os.path.dirname(os.path.abspath(target_file))
    rel = os.path.relpath(dest, start)
    return rel.replace(os.sep, "/")

def md_link(ref_or_parsed, target_file):
    """Full '[Label](rel/path.md)' markdown link."""
    p = ref_or_parsed if isinstance(ref_or_parsed, dict) else parse_ref(ref_or_parsed)
    return f"[{p['label']}]({link(p, target_file)})"

if __name__ == "__main__":
    import sys
    target = sys.argv[2] if len(sys.argv) > 2 else "study/x.md"
    p = parse_ref(sys.argv[1])
    print("label :", p["label"])
    print("file  :", os.path.relpath(file_path(p), ROOT).replace(os.sep, "/"),
          "(exists)" if os.path.exists(file_path(p)) else "(MISSING)")
    print("link  :", md_link(p, target))
