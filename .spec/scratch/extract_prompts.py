import json, sys, io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

def text_of(content):
    # content may be a string or a list of blocks
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for b in content:
            if isinstance(b, dict):
                if b.get("type") == "text" and isinstance(b.get("text"), str):
                    parts.append(b["text"])
                elif b.get("type") == "tool_result":
                    return None  # tool result, not a human prompt
            elif isinstance(b, str):
                parts.append(b)
        return "\n".join(parts) if parts else None
    return None

SKIP_PREFIXES = ("<system-reminder>", "<command-name", "<command-message",
                 "<local-command", "<command-args", "Caveat:", "[Request interrupted")

def is_human(txt):
    if not txt:
        return False
    t = txt.strip()
    if not t:
        return False
    for p in SKIP_PREFIXES:
        if t.startswith(p):
            return False
    # skip pure tool-result-ish noise
    if t.startswith("[{") or t.startswith("{\""):
        return False
    return True

for path in sys.argv[1:]:
    print(f"\n\n######## FILE: {path} ########")
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except Exception:
                continue
            if obj.get("type") != "user":
                continue
            msg = obj.get("message")
            if not isinstance(msg, dict):
                continue
            if msg.get("role") != "user":
                continue
            txt = text_of(msg.get("content"))
            if is_human(txt):
                print("\n---")
                print(txt.strip())
