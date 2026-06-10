#!/usr/bin/env bash
# CT2.4 RUN 2 driver — drive a treatment + control persona session through a
# multi-turn, research-heavy room conversation and observe whether the
# context-tools arm (codewright-ct2) actually uses the levers between turns.
# Both arms are kimi-k2.6 (Michael: still interesting to watch on a strong model).
set -u
PSQL=(docker exec pg-ai-stewards-dev psql -U stewards -d stewards -At)

QUESTIONS=(
  "how does the gateway authenticate a persona versus a human?"
  "how does the hub broadcast a message and avoid echoing it back to the sender?"
  "how does history replay work when a client subscribes to a room?"
  "how are persona keys minted, hashed, and validated?"
  "how does presence and the roster get tracked and broadcast?"
  "how does room access control differ for personas versus humans?"
)

drive() {  # $1 = pipeline_family, $2 = slug-tag
  local pipe="$1" tag="$2" child sess wq status maturity
  echo "[$tag] turn 0 → ${QUESTIONS[0]}"
  child=$("${PSQL[@]}" -c "SELECT stewards.spawn_subagent_create('$pipe',
    E'You are Codewright in the #engineering room.\n[dev]: ${QUESTIONS[0]}\n\nAnswer it (research ai-chattermax).',
    NULL, 800000, NULL, '${tag}-t0', 'persona')::text;")
  # wait verified
  for _ in $(seq 1 40); do sleep 8
    read -r status maturity sess < <("${PSQL[@]}" -F' ' -c "SELECT status, maturity, coalesce(session_ids[array_length(session_ids,1)],'') FROM stewards.work_items WHERE id='$child';")
    [ "$maturity" = "verified" ] && break
    case "$status" in failed|cancelled) echo "[$tag] t0 ended $status"; return;; esac
  done
  [ -z "${sess:-}" ] && { echo "[$tag] no session"; return; }
  echo "[$tag] session=$sess"
  local i
  for i in 1 2 3 4 5; do
    echo "[$tag] turn $i → ${QUESTIONS[$i]}"
    wq=$("${PSQL[@]}" -c "SELECT stewards.consult_subagent_dispatch('$sess',
      E'[CONSULT] [dev]: ${QUESTIONS[$i]}\n\nReply in character, or SILENCE.')::text;")
    for _ in $(seq 1 40); do sleep 8
      status=$("${PSQL[@]}" -c "SELECT status FROM stewards.work_queue WHERE id=$wq;")
      case "$status" in done|error) break;; esac
    done
  done
  echo "[$tag] DONE session=$sess"
}

echo "=== TREATMENT (codewright-ct2, context tools on) ==="
drive persona-turn-code-ct2 ct2t
echo "=== CONTROL (codewright, no context tools) ==="
drive persona-turn-code ct2c
echo "=== ALL DONE ==="
