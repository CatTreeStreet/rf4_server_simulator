#!/usr/bin/env bash
set -Eeuo pipefail
cd /app
if [[ ! -f /app/config.yaml ]]; then
  echo "ERROR: /app/config.yaml is missing. Mount a configured config.yaml (see docs/docker.md)." >&2
  exit 2
fi
if [[ ! -f /app/server_data/static.sqlite3 ]]; then
  echo "ERROR: bundled server_data/static.sqlite3 is missing." >&2
  exit 2
fi
mkdir -p /app/data /app/certs
echo "[rf4] Initializing database..."
seed -config /app/config.yaml
echo "[rf4] Starting gameserver..."
gameserver -config /app/config.yaml &
GAME_PID=$!
cleanup() {
  echo "[rf4] Stopping services..."
  kill "$GAME_PID" "${GATEWAY_PID:-}" 2>/dev/null || true
  wait "$GAME_PID" 2>/dev/null || true
  if [[ -n "${GATEWAY_PID:-}" ]]; then wait "$GATEWAY_PID" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM
echo "[rf4] Waiting for gameserver readiness..."
ready=0
for _ in $(seq 1 60); do
  if curl -fsS http://127.0.0.1:9502/readyz >/dev/null; then ready=1; break; fi
  if ! kill -0 "$GAME_PID" 2>/dev/null; then
    echo "ERROR: gameserver exited before becoming ready." >&2
    wait "$GAME_PID" || true
    exit 1
  fi
  sleep 1
done
if [[ "$ready" != 1 ]]; then echo "ERROR: gameserver did not become ready within 60 seconds." >&2; exit 1; fi
echo "[rf4] Starting gateway..."
gateway -config /app/config.yaml &
GATEWAY_PID=$!
set +e
wait -n "$GAME_PID" "$GATEWAY_PID"
status=$?
set -e
echo "[rf4] A service exited (status=$status); stopping the container."
exit "$status"
