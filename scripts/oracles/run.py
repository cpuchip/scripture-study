#!/usr/bin/env python3
"""run — the oracle runner.

Reads registry.yaml, figures out which deterministic checks are in-scope for a set
of files, runs them, and reports which DECISION-CLASSES are now green — i.e. which
work the agent may leave in act-and-report under the CLAUDE.md "the oracle is the
switch" rule (.spec/proposals/oracle-floored-autonomy.md).

This is the manual form. The same runner is what a Stop hook would call to make the
check automatic — but that changes the session's feel, so it stays Michael's toggle.

Usage:
  python scripts/oracles/run.py FILE ...      # run in-scope oracles on these files
  python scripts/oracles/run.py               # default: git-changed files (porcelain)
  python scripts/oracles/run.py --list        # print the registry
  python scripts/oracles/run.py --all         # run against study/** lessons/**

Exit 0 if every in-scope HARD oracle passed (or none were in scope); 1 otherwise.
Advisory oracles never affect the exit code.
"""
import os
import re
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
REGISTRY = os.path.join(os.path.dirname(__file__), "registry.yaml")

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")  # Windows cp1252 → utf-8
except Exception:
    pass

try:
    import yaml
except ImportError:
    sys.exit("run.py needs pyyaml (pip install pyyaml) — the registry is YAML.")


def glob_to_re(pattern):
    """Translate a path glob (** crosses dirs, * does not) to a regex."""
    out, i = [], 0
    while i < len(pattern):
        c = pattern[i]
        if pattern[i:i + 2] == "**":
            out.append(".*")
            i += 2
            if i < len(pattern) and pattern[i] == "/":
                i += 1  # consume the slash after ** so study/**/x also matches study/x
        elif c == "*":
            out.append("[^/]*")
            i += 1
        elif c == "?":
            out.append("[^/]")
            i += 1
        else:
            out.append(re.escape(c))
            i += 1
    return re.compile("^" + "".join(out) + "$")


def matches(path, patterns):
    p = path.replace("\\", "/")
    return any(glob_to_re(g).match(p) for g in patterns)


def git_changed():
    try:
        out = subprocess.run(["git", "status", "--porcelain"], cwd=ROOT,
                             capture_output=True, text=True).stdout
    except Exception:
        return []
    files = []
    for line in out.splitlines():
        if len(line) < 4:
            continue
        path = line[3:].strip()
        if " -> " in path:        # rename
            path = path.split(" -> ")[-1]
        if path.endswith(".md"):
            files.append(path)
    return files


def default_all():
    found = []
    for base in ("study", "lessons", "becoming"):
        d = os.path.join(ROOT, base)
        for dirpath, _, names in os.walk(d):
            if os.sep + "." in dirpath:   # skip .audit/.scratch
                continue
            for n in names:
                if n.endswith(".md"):
                    rel = os.path.relpath(os.path.join(dirpath, n), ROOT)
                    found.append(rel.replace("\\", "/"))
    return found


def main(argv):
    oracles = yaml.safe_load(open(REGISTRY, encoding="utf-8"))["oracles"]

    if "--list" in argv:
        print("Oracle registry:\n")
        for o in oracles:
            print(f"  {o['name']:18} [{o.get('tier','hard')}] unlocks: {o['unlocks']}")
            print(f"  {'':18} scope: {', '.join(o['scope'])}")
            print(f"  {'':18} {o['guarantees']}\n")
        return 0

    if "--all" in argv:
        targets = default_all()
    else:
        files = [a for a in argv if not a.startswith("--")]
        targets = files or git_changed()

    if not targets:
        print("No target .md files (pass files, or have git-changed .md files, or --all).")
        return 0

    print(f"Oracle run over {len(targets)} file(s).\n")
    hard_fail = False
    unlocked, blocked, skipped = [], [], []

    for o in oracles:
        in_scope = [f for f in targets if matches(f, o["scope"])]
        if not in_scope:
            skipped.append(o["name"])
            continue
        cmd = []
        for tok in o["run"].split():
            cmd.extend(in_scope if tok == "{files}" else [tok])
        proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True)
        ok = proc.returncode == 0
        tier = o.get("tier", "hard")
        status = "PASS" if ok else ("FAIL" if tier == "hard" else "flags")
        print(f"  [{status:5}] {o['name']:18} ({len(in_scope)} file(s)) → {o['unlocks']}")
        if not ok:
            tail = (proc.stdout or proc.stderr or "").strip().splitlines()
            for ln in tail[-6:]:
                print(f"            {ln}")
            if tier == "hard":
                hard_fail = True
                blocked.append(o["unlocks"])
            else:
                unlocked.append(o["unlocks"] + " (advisory flags — review)")
        else:
            unlocked.append(o["unlocks"])

    print("\n— summary —")
    if unlocked:
        print("  green (act-and-report): " + ", ".join(unlocked))
    if blocked:
        print("  BLOCKED (surface-first until green): " + ", ".join(blocked))
    if skipped:
        print(f"  not in scope: {', '.join(skipped)}")
    return 1 if hard_fail else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
