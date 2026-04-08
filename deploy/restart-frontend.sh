#!/usr/bin/env bash
# 在服务器上启动 / 重启 xyKitchen 网页前端（Node + Express，默认 5401；页面请求后端 5402 /api）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEB="$ROOT/frontend-web"
mkdir -p "$ROOT/logs" "$ROOT/run"
cd "$WEB"
if [[ ! -f package.json ]]; then
  echo "missing frontend-web/package.json" >&2
  exit 1
fi
if [[ ! -d node_modules ]]; then
  npm ci --omit=dev 2>/dev/null || npm install --omit=dev
fi
PID_FILE="$ROOT/run/frontend.pid"
if [[ -f "$PID_FILE" ]]; then
  OLD="$(cat "$PID_FILE" 2>/dev/null || true)"
  if [[ -n "${OLD}" ]] && kill -0 "$OLD" 2>/dev/null; then
    kill "$OLD" || true
    sleep 1
  fi
fi
if command -v fuser >/dev/null 2>&1; then
  fuser -k 5401/tcp 2>/dev/null || true
elif command -v lsof >/dev/null 2>&1; then
  L="$(lsof -t -i:5401 2>/dev/null || true)"
  if [[ -n "${L}" ]]; then kill $L || true; sleep 1; fi
fi
export PORT="${PORT:-5401}"
nohup node server.js >>"$ROOT/logs/frontend.log" 2>&1 &
echo $! >"$PID_FILE"
sleep 1
curl -sf "http://127.0.0.1:${PORT}/" | head -c 80 >/dev/null || { echo "frontend health check failed"; exit 1; }
echo "xyKitchen frontend (port ${PORT}) OK"
