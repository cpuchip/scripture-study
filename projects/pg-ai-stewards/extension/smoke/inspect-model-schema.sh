#!/bin/sh
# Dumps the raw first model object from each provider's /models response
# so we can see whether pricing fields are included in the gateway schema.
echo "=== opencode_go raw (first 1200 chars) ==="
curl -s -H "Authorization: Bearer $STEWARDS_PROVIDER_OPENCODE_GO_API_KEY" \
  https://opencode.ai/zen/go/v1/models | head -c 1200
echo ""
echo ""
echo "=== gemini raw (first 1200 chars) ==="
curl -s -H "Authorization: Bearer $STEWARDS_PROVIDER_GOOGLE_GEMINI_API_KEY" \
  https://generativelanguage.googleapis.com/v1beta/openai/models | head -c 1200
echo ""
