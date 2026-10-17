#!/usr/bin/env bash
# 系统模块冒烟：角色、菜单、单页的创建/更新/删除。
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
JAR="$(mktemp)"
TMP="$(mktemp -d)"
trap 'rm -rf "$JAR" "$TMP"' EXIT

curl -sS -c "$JAR" "$BASE/admin/login" -o "$TMP/login.html"
token=$(sed -n 's/.*name="_csrf" value="\([^"]*\)".*/\1/p' "$TMP/login.html" | head -1)
curl -sS -b "$JAR" -c "$JAR" -o /dev/null -X POST "$BASE/admin/login" \
  --data-urlencode "_csrf=$token" --data-urlencode "username=admin" --data-urlencode "password=Admin@123456"
token=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin" | sed -n 's/.*name="csrf-token" content="\([^"]*\)".*/\1/p' | head -1)

ts=$(date +%s)
role_slug="smoke-role-$ts"
menu_name="冒烟菜单$ts"
page_slug="smoke-page-$ts"
headers="$TMP/headers"

code=$(curl -sS -b "$JAR" -c "$JAR" -D "$headers" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/roles" \
  --data-urlencode "_csrf=$token" --data-urlencode "name=冒烟角色$ts" --data-urlencode "slug=$role_slug" \
  --data-urlencode "description=smoke" --data-urlencode "permission_ids=1" --data-urlencode "permission_ids=2")
[ "$code" = "302" ] || { echo "role create: $code"; exit 1; }
role_id=$(tr -d '\r' < "$headers" | sed -n 's#.*/admin/roles/\([0-9]*\)/edit.*#\1#p' | tail -1)
[ -n "$role_id" ] || { echo "role id missing"; exit 1; }
echo "role=$role_id"

code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/roles/$role_id" \
  --data-urlencode "_csrf=$token" --data-urlencode "_method=PUT" --data-urlencode "name=冒烟角色更新$ts" \
  --data-urlencode "slug=$role_slug" --data-urlencode "description=updated" --data-urlencode "permission_ids=1")
[ "$code" = "302" ] || { echo "role update: $code"; exit 1; }
echo "ok role update"

code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/menus" \
  --data-urlencode "_csrf=$token" --data-urlencode "name=$menu_name" --data-urlencode "icon=test-pipe" \
  --data-urlencode "path=/admin" --data-urlencode "parent_id=0" --data-urlencode "sort=98" \
  --data-urlencode "permission_code=dashboard:view" --data-urlencode "is_visible=on")
[ "$code" = "302" ] || { echo "menu create: $code"; exit 1; }
menu_html=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin/menus")
menu_id=$(printf '%s' "$menu_html" | sed -n "s/.*data-id=\"\([0-9]*\)\" data-parent=.*data-name=\"$menu_name\".*/\1/p" | head -1)
[ -n "$menu_id" ] || { echo "menu id missing"; exit 1; }
echo "menu=$menu_id"

code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/menus/$menu_id" \
  --data-urlencode "_csrf=$token" --data-urlencode "_method=PUT" --data-urlencode "name=${menu_name}更新" \
  --data-urlencode "icon=test-pipe" --data-urlencode "path=/admin" --data-urlencode "parent_id=0" \
  --data-urlencode "sort=99" --data-urlencode "is_visible=on" --data-urlencode "permission_code=dashboard:view")
[ "$code" = "302" ] || { echo "menu update: $code"; exit 1; }
echo "ok menu update"

code=$(curl -sS -b "$JAR" -c "$JAR" -D "$headers" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/pages" \
  --data-urlencode "_csrf=$token" --data-urlencode "title=冒烟单页$ts" --data-urlencode "slug=$page_slug" \
  --data-urlencode "content=<p>v1</p>" --data-urlencode "status=published" --data-urlencode "sort=1")
[ "$code" = "302" ] || { echo "page create: $code"; exit 1; }
page_id=$(tr -d '\r' < "$headers" | sed -n 's#.*/admin/pages/\([0-9]*\)/edit.*#\1#p' | tail -1)
[ -n "$page_id" ] || { echo "page id missing"; exit 1; }
code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/pages/$page_id" \
  --data-urlencode "_csrf=$token" --data-urlencode "_method=PUT" --data-urlencode "title=冒烟单页更新$ts" \
  --data-urlencode "slug=$page_slug" --data-urlencode "content=<p>v2</p>" --data-urlencode "status=published" --data-urlencode "sort=1")
[ "$code" = "302" ] || { echo "page update: $code"; exit 1; }
curl -sS -o "$TMP/page.html" "$BASE/p/$page_slug"
grep -q "冒烟单页更新$ts" "$TMP/page.html" || { echo "page update not visible"; exit 1; }
echo "ok page update"

# 清理
curl -sS -b "$JAR" -c "$JAR" -o /dev/null -X POST "$BASE/admin/pages/$page_id/delete" --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE"
curl -sS -b "$JAR" -c "$JAR" -o /dev/null -X POST "$BASE/admin/menus/$menu_id/delete" --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE"
curl -sS -b "$JAR" -c "$JAR" -o /dev/null -X POST "$BASE/admin/roles/$role_id/delete" --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE"
echo "ok cleanup"
