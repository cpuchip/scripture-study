#!/bin/sh
# Capture opencode's Anthropic-format SSE shape for qwen3.7-max so the
# substrate's parse_anthropic_sse (Rust) can reassemble it. Key stays in-container.
KEY="$STEWARDS_PROVIDER_OPENCODE_GO_API_KEY"
curl -s -N -X POST "https://opencode.ai/zen/go/v1/messages" \
  -H "x-api-key: $KEY" -H "Content-Type: application/json" -H "anthropic-version: 2023-06-01" \
  -d '{"model":"qwen3.7-max","max_tokens":80,"stream":true,"messages":[{"role":"user","content":"Reply with exactly: OK"}]}' \
  | head -c 2500
