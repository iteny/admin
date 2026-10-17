#!/usr/bin/env bash
# 验证生产模式：模板与静态资源来自 //go:embed，不依赖工作目录中的 web/ 文件。
set -euo pipefail
cd "$(dirname "$0")/.."

sed -e 's/addr: ":8080"/addr: ":18080"/' \
    -e 's/debug: true/debug: false/' \
    -e 's#dsn: "storage/cms.db"#dsn: "storage/cms-embed.db"#' \
    -e 's#change-this-session-secret-in-production#embed-smoke-secret-please-change#' \
    config/config.yaml > /tmp/cms-embed.yaml

/tmp/cms-bin -version
pkill -f '/tmp/cms-bin' 2>/dev/null || true
nohup /tmp/cms-bin -config /tmp/cms-embed.yaml > /tmp/cms-embed.log 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true' EXIT
sleep 2

for path in / /admin/login /static/vendor/tabler/css/tabler.min.css /static/vendor/tabler/js/tabler.min.js /static/js/admin.js; do
  code=$(curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:18080$path")
  [ "$code" = "200" ] || { echo "FAIL $path -> $code"; exit 1; }
  echo "ok $path"
done
