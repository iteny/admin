#!/usr/bin/env bash
# 单二进制部署验证：只复制二进制、config.yaml 与空 storage 目录。
set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p bin
go build -trimpath -o bin/cms ./cmd/cms
work=$(mktemp -d)
trap 'kill "${pid:-0}" 2>/dev/null || true; rm -rf "$work"' EXIT
cp bin/cms "$work/cms"
mkdir -p "$work/storage"
sed -e 's/addr: ":8080"/addr: ":18082"/' \
    -e 's/debug: true/debug: false/' \
    -e 's#dsn: "storage/cms.db"#dsn: "storage/cms.db"#' \
    config/config.yaml > "$work/config.yaml"

pushd "$work" >/dev/null
./cms -config config.yaml > server.log 2>&1 &
pid=$!
popd >/dev/null
for _ in $(seq 1 30); do
  if curl -fsS -o /dev/null "http://127.0.0.1:18082/healthz" 2>/dev/null; then
    break
  fi
  sleep 0.5
done
for path in /healthz /admin/login /static/vendor/tabler/css/tabler.min.css; do
  code=$(curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:18082$path")
  [ "$code" = "200" ] || { echo "FAIL $path -> $code"; cat "$work/server.log"; exit 1; }
  echo "ok $path"
done
