import json, sys, io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

def find_text(d):
    # data may hold message text in various shapes
    if isinstance(d, str):
        return d
    if isinstance(d, dict):
        for k in ("text", "content", "message", "prompt"):
            v = d.get(k)
            if isinstance(v, str) and v.strip():
                return v
            if isinstance(v, list):
                parts = []
                for b in v:
                    if isinstance(b, dict) and isinstance(b.get("text"), str):
                        parts.append(b["text"])
                    elif isinstance(b, str):
                        parts.append(b)
                if parts:
                    return "\n".join(parts)
    return None

for path in sys.argv[1:]:
    print(f"\n\n######## COPILOT: {path} ########")
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except Exception:
                continue
            if obj.get("type") != "user.message":
                continue
            txt = find_text(obj.get("data"))
            if txt and txt.strip():
                t = txt.strip()
                if t.startswith("<") and t.endswith(">"):
                    continue
                print("\n---")
                print(t)
