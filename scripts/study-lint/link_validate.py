#!/usr/bin/env python3
"""link-validate — every relative .md link in a study resolves to a real file, and
every SCRIPTURE link's label matches the file it points to.

The second study-lint rule. Catches the broken/mis-pathed link class the walk
hit by hand — exo->ex, mic->micah, dc/76 labeled "D&C 109", links to a directory,
the podcast trio's dead paths. Unlike scripture-verbatim's findings, these are
OBJECTIVE (a path is right or wrong), so they're safe to just fix.

  BROKEN    — the target file doesn't exist (relative to the source file)
  DIR       — the link points at a directory, not a specific file (project rule:
              always link the specific file)
  MISMATCH  — a scripture link whose label resolves to a DIFFERENT file than the
              path points to (e.g. label "D&C 109:76" linking dc/76.md)

Reuses the quoter's resolver for the label->file map. Run manually; exit 1 on any
flag.
"""
import sys, os, re

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "quoter"))
import resolver

LINK = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')
SKIP = ("http://", "https://", "mailto:", "#", "tel:")

def check_file(path):
    text = open(path, encoding="utf-8").read()
    base = os.path.dirname(os.path.abspath(path))
    out = []
    for m in LINK.finditer(text):
        label, target = m.group(1), m.group(2).strip()
        if target.startswith(SKIP) or target.startswith("!"):
            continue
        tgt = target.split("#", 1)[0].strip()        # drop any #fragment
        if not tgt:
            continue                                  # pure in-page anchor
        is_md = tgt.endswith(".md")
        looks_dir = tgt.endswith("/") or (not os.path.splitext(tgt)[1])
        if not (is_md or looks_dir):
            continue                                  # image / asset / other — skip
        abspath = os.path.normpath(os.path.join(base, tgt))
        if not os.path.exists(abspath):
            out.append(("BROKEN", label, target))
            continue
        if os.path.isdir(abspath):
            out.append(("DIR", label, target))
            continue
        # scripture label/path agreement
        try:
            p = resolver.parse_ref(label)
        except resolver.RefError:
            continue
        if "scriptures" not in abspath.replace("\\", "/"):
            continue                                  # label parses but link isn't scripture
        expect = os.path.normpath(resolver.file_path(p))
        if os.path.normcase(expect) != os.path.normcase(abspath):
            out.append(("MISMATCH", label,
                        f'links {target} but "{label}" is {os.path.relpath(expect, base).replace(os.sep,"/")}'))
    return out

def main(argv):
    files = argv[1:] or ([l.strip() for l in sys.stdin if l.strip()] if not sys.stdin.isatty() else [])
    files = [f for f in files if f.endswith(".md") and os.path.isfile(f)]
    if not files:
        print("usage: link_validate.py <file.md> ...", file=sys.stderr)
        return 2
    bad = 0
    for path in files:
        res = check_file(path)
        if not res:
            continue
        print(f"\n{path}  ({len(res)} link issue(s))")
        for level, label, msg in res:
            print(f"  [{level}] {label}: {msg}" if level != "BROKEN"
                  else f"  [BROKEN] {label} -> {msg}")
            bad += 1
    print(f"\n— {len(files)} file(s); {bad} link flag(s).")
    return 1 if bad else 0

if __name__ == "__main__":
    sys.exit(main(sys.argv))
