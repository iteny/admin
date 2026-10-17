#!/usr/bin/env bash
# 基础冒烟测试：登录后遍历后台页面，并检查前台页面。
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
JAR="$(mktemp)"
LOGIN_HTML="$(mktemp)"
trap 'rm -f "$JAR" "$LOGIN_HTML"' EXIT

fail=0
check() {
  local path="$1" expected="$2"
  local code
  code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' "$BASE$path")
  if [ "$code" != "$expected" ]; then
    echo "FAIL $path -> $code (want $expected)"
    fail=1
  else
    echo "ok   $path -> $code"
  fi
}

check "/healthz" 200
curl -sS -c "$JAR" "$BASE/admin/login" -o "$LOGIN_HTML"
token=$(sed -n 's/.*name="_csrf" value="\([^"]*\)".*/\1/p' "$LOGIN_HTML" | head -1)
if [ -z "$token" ]; then
  echo "FAIL: no CSRF token on login page"
  exit 1
fi

code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/login" \
  --data-urlencode "_csrf=$token" \
  --data-urlencode "username=admin" \
  --data-urlencode "password=Admin@123456" \
  --data-urlencode "redirect=/admin")
if [ "$code" != "302" ]; then
  echo "FAIL login -> $code"
  exit 1
fi
echo "ok   login -> 302"

for path in /admin /admin/articles /admin/articles/create /admin/categories /admin/tags /admin/pages \
  /admin/pages/create /admin/media /admin/media/picker /admin/users /admin/users/create \
  /admin/users/1/edit /admin/roles /admin/roles/create /admin/roles/2/edit /admin/menus /admin/settings /admin/logs /admin/profile \
  / /search /sitemap.xml /robots.txt; do
  check "$path" 200
done

exit "$fail"
