#!/bin/sh
# Verify Gemini's OpenAI-compatible chat endpoint works and determine the
# correct model-id format (bare vs "models/" prefix). Tiny calls on the
# cheapest model. Key stays in-container.
EP="https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"

echo "=== A) bare id: gemini-2.5-flash-lite ==="
curl -s -H "Authorization: Bearer $STEWARDS_PROVIDER_GOOGLE_GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash-lite","max_tokens":16,"messages":[{"role":"user","content":"Reply with the single word: pong"}]}' \
  "$EP" | head -c 700
echo ""
echo ""
echo "=== B) prefixed id: models/gemini-2.5-flash-lite ==="
curl -s -H "Authorization: Bearer $STEWARDS_PROVIDER_GOOGLE_GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"models/gemini-2.5-flash-lite","max_tokens":16,"messages":[{"role":"user","content":"Reply with the single word: pong"}]}' \
  "$EP" | head -c 700
echo ""
