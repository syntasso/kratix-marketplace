#!/usr/bin/env sh
set -e

ollama serve &
pid=$!

# Wait up to 120s for API
for i in $(seq 1 120); do
  if curl -fsS http://127.0.0.1:11434/api/version >/dev/null; then
    break
  fi
  sleep 1
done

# Warm TinyDolphin
printf '%s' '{"model":"tinydolphin:latest","prompt":" "}' \
  | curl -sS -X POST http://127.0.0.1:11434/api/generate -d @- >/dev/null || true

# Keep container alive on the server process
wait "$pid"
