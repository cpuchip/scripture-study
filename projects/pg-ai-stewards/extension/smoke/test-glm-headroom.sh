#!/bin/sh
# Confirm glm-5.1 is a reasoning model that needs token headroom: the empty
# content at max_tokens=80 should fill in at a larger cap, with reasoning in
# a separate channel.
KEY="$STEWARDS_PROVIDER_OPENCODE_GO_API_KEY"
EP="https://opencode.ai/zen/go/v1/chat/completions"
PROMPT="Name one cheap way to test an LLM integration. One sentence."

echo "=== glm-5.1 non-streaming, max_tokens=1500 ==="
r=$(curl -s -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d "{\"model\":\"glm-5.1\",\"max_tokens\":1500,\"messages\":[{\"role\":\"user\",\"content\":\"$PROMPT\"}]}" "$EP")
echo "  finish_reason: $(echo "$r" | grep -oE '"finish_reason":"[^"]*"' | head -1)"
echo "  usage: $(echo "$r" | grep -oE '"(prompt|completion|total)_tokens":[0-9]*' | tr '\n' ' ')"
# content length
cl=$(echo "$r" | grep -oE '"content":"([^"\\]|\\.)*"' | head -1 | wc -c)
echo "  content field length(chars incl quotes): $cl"
echo "  content sample: $(echo "$r" | grep -oE '"content":"([^"\\]|\\.)*"' | head -1 | head -c 200)"
# reasoning channel?
echo "  has reasoning_content key: $(echo "$r" | grep -oE '"reasoning_content"' | head -1)"
echo "  has reasoning key: $(echo "$r" | grep -oE '"reasoning"' | head -1)"
rl=$(echo "$r" | grep -oE '"reasoning_content":"([^"\\]|\\.)*"' | head -1 | wc -c)
echo "  reasoning_content length(chars): $rl"

echo ""
echo "=== glm-5.1 STREAMING, max_tokens=1500 — where do the tokens go? ==="
s=$(curl -s -N -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d "{\"model\":\"glm-5.1\",\"stream\":true,\"stream_options\":{\"include_usage\":true},\"max_tokens\":1500,\"messages\":[{\"role\":\"user\",\"content\":\"$PROMPT\"}]}" "$EP")
echo "  delta.content chars:          $(echo "$s" | grep -oE '"content":"([^"\\]|\\.)*"' | sed 's/.*"content":"//;s/"$//' | tr -d '\n' | wc -c)"
echo "  delta.reasoning_content chars: $(echo "$s" | grep -oE '"reasoning_content":"([^"\\]|\\.)*"' | sed 's/.*"reasoning_content":"//;s/"$//' | tr -d '\n' | wc -c)"
echo "  delta.reasoning chars:         $(echo "$s" | grep -oE '"reasoning":"([^"\\]|\\.)*"' | sed 's/.*"reasoning":"//;s/"$//' | tr -d '\n' | wc -c)"
echo "  finish_reason: $(echo "$s" | grep -oE '"finish_reason":"[^"]*"' | tail -1)"
echo "  usage: $(echo "$s" | grep -oE '"usage":\{[^}]*\}' | tail -1 | head -c 200)"
echo "  --- first data chunk with a delta (structure sample) ---"
echo "$s" | grep -E '"delta"' | head -1 | head -c 400
