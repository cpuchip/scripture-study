#!/bin/sh
# Diagnostic for the brainstorm-run empties (qwen3.7-max hard-failed; glm-5
# completed-but-empty). Tests each model directly at the opencode gateway,
# both non-streaming and streaming (the substrate streams), to isolate
# "model flaky" from "streaming-path issue". Key stays in-container.
KEY="$STEWARDS_PROVIDER_OPENCODE_GO_API_KEY"
EP="https://opencode.ai/zen/go/v1/chat/completions"
PROMPT="Reply with one short sentence about why testing matters."

echo "=== available glm / qwen models on the gateway ==="
curl -s -H "Authorization: Bearer $KEY" https://opencode.ai/zen/go/v1/models \
  | grep -oE '"id":"[^"]*"' | grep -iE 'glm|qwen' | sort

for M in qwen3.7-max glm-5 glm-5.1; do
  echo ""
  echo "########## $M ##########"

  echo "--- non-streaming ---"
  ns=$(curl -s -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
    -d "{\"model\":\"$M\",\"max_tokens\":80,\"messages\":[{\"role\":\"user\",\"content\":\"$PROMPT\"}]}" "$EP")
  echo "  content: $(echo "$ns" | grep -oE '"content":"[^"]*"' | head -1)"
  echo "  finish_reason: $(echo "$ns" | grep -oE '"finish_reason":"[^"]*"' | head -1)"
  echo "  usage: $(echo "$ns" | grep -oE '"(prompt|completion|total)_tokens":[0-9]*' | tr '\n' ' ')"
  err=$(echo "$ns" | grep -oE '"(error|message)":"[^"]*"' | head -2)
  [ -n "$err" ] && echo "  ERROR: $err"
  echo "  raw[0:160]: $(echo "$ns" | head -c 160)"

  echo "--- streaming (stream_options.include_usage, like the substrate) ---"
  st=$(curl -s -N -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
    -d "{\"model\":\"$M\",\"stream\":true,\"stream_options\":{\"include_usage\":true},\"max_tokens\":80,\"messages\":[{\"role\":\"user\",\"content\":\"$PROMPT\"}]}" "$EP")
  clen=$(echo "$st" | grep -oE '"content":"[^"]*"' | sed 's/.*"content":"//;s/"$//' | tr -d '\n' | wc -c)
  echo "  streamed content chars: $clen"
  echo "  finish_reason: $(echo "$st" | grep -oE '"finish_reason":"[^"]*"' | tail -1)"
  echo "  usage line: $(echo "$st" | grep -oE '"usage":\{[^}]*\}' | tail -1 | head -c 160)"
  serr=$(echo "$st" | grep -oiE 'error[^,}]*' | head -1)
  [ -n "$serr" ] && echo "  STREAM ERROR: $serr"
done
