#!/usr/bin/env bash
# 仅在服务器上使用：编译 xyKitchen 后端并替换监听 5402 的进程。不操作 Docker。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$ROOT/logs" "$ROOT/run"
cd "$ROOT/backend"
export GOTOOLCHAIN=local
if command -v go >/dev/null 2>&1; then
  go build -o xykitchen-server ./cmd/server
elif command -v docker >/dev/null 2>&1; then
  # 一次性编译容器，不重启现有业务容器、不动 compose
  DOCKER="docker"
  if ! docker info >/dev/null 2>&1; then DOCKER="sudo docker"; fi
  $DOCKER run --rm \
    -e GOPROXY=https://goproxy.cn,direct \
    -e GOSUMDB=sum.golang.google.cn \
    -v "$ROOT/backend:/app" -w /app \
    golang:1.22-bookworm \
    bash -c 'go mod tidy && go build -o xykitchen-server ./cmd/server'
else
  echo "需要本机 Go 或 Docker 以编译后端" >&2
  exit 1
fi
PID_FILE="$ROOT/run/backend.pid"
if [[ -f "$PID_FILE" ]]; then
  OLD="$(cat "$PID_FILE" 2>/dev/null || true)"
  if [[ -n "${OLD}" ]] && kill -0 "$OLD" 2>/dev/null; then
    kill "$OLD" || true
    sleep 1
  fi
fi
# 若仍有进程占用 5402（旧启动方式），仅结束该端口上的进程
if command -v fuser >/dev/null 2>&1; then
  fuser -k 5402/tcp 2>/dev/null || true
elif command -v lsof >/dev/null 2>&1; then
  L="$(lsof -t -i:5402 2>/dev/null || true)"
  if [[ -n "$L" ]]; then kill $L || true; sleep 1; fi
fi
nohup ./xykitchen-server >>"$ROOT/logs/backend.log" 2>&1 &
echo $! >"$PID_FILE"
sleep 1
curl -sf "http://127.0.0.1:5402/api/health" | head -c 200 || { echo "health check failed"; exit 1; }
echo "xyKitchen backend restarted OK"
