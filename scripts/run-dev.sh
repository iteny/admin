#!/usr/bin/env bash
# 本地开发辅助：构建并后台启动服务（仅用于冒烟测试）。
set -euo pipefail
cd "$(dirname "$0")/.."

pkill -f '/tmp/cms-bin' 2>/dev/null || true
go build -o /tmp/cms-bin ./cmd/cms
nohup /tmp/cms-bin -config config/config.yaml > /tmp/cms.log 2>&1 &
echo $! > /tmp/cms.pid
for _ in $(seq 1 20); do
  if curl -fsS -o /dev/null http://127.0.0.1:8080/healthz 2>/dev/null; then
    echo "health: 200"
    exit 0
  fi
  sleep 0.5
done
echo "health: failed" >&2
exit 1
