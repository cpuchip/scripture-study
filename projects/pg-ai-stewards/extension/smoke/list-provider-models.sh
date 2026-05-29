#!/bin/sh
# Lists models from opencode_go + google_gemini live /models endpoints.
# Run inside the pg or bridge container so the API keys (in env) never
# leave the container. Prints model ids only, never the key.
#
#   docker cp .../list-provider-models.sh pg-ai-stewards-dev:/tmp/lpm.sh
#   docker exec pg-ai-stewards-dev sh /tmp/lpm.sh

echo "=== opencode_go (https://opencode.ai/zen/go/v1/models) ==="
curl -s -H "Authorization: Bearer $STEWARDS_PROVIDER_OPENCODE_GO_API_KEY" \
  https://opencode.ai/zen/go/v1/models \
  | grep -oE '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | sort -u

echo ""
echo "=== google_gemini openai-compat (.../v1beta/openai/models) ==="
curl -s -H "Authorization: Bearer $STEWARDS_PROVIDER_GOOGLE_GEMINI_API_KEY" \
  https://generativelanguage.googleapis.com/v1beta/openai/models \
  | grep -oE '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | sort -u
