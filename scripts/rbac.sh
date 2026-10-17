#!/usr/bin/env bash
# RBAC 冒烟：创建编辑角色用户，验证只能访问内容模块。
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
JAR="$(mktemp)"
EDITOR_JAR="$(mktemp)"
TMP="$(mktemp -d)"
trap 'rm -rf "$JAR" "$EDITOR_JAR" "$TMP"' EXIT

login() {
  local jar="$1" user="$2" pass="$3"
  curl -sS -c "$jar" "$BASE/admin/login" -o "$TMP/login.html"
  local token
  token=$(sed -n 's/.*name="_csrf" value="\([^"]*\)".*/\1/p' "$TMP/login.html" | head -1)
  curl -sS -b "$jar" -c "$jar" -o /dev/null -X POST "$BASE/admin/login" \
    --data-urlencode "_csrf=$token" --data-urlencode "username=$user" --data-urlencode "password=$pass" \
    --data-urlencode "redirect=/admin"
}

token_of() {
  local jar="$1"
  curl -sS -b "$jar" -c "$jar" "$BASE/admin" | sed -n 's/.*name="csrf-token" content="\([^"]*\)".*/\1/p' | head -1
}

login "$JAR" admin Admin@123456
admin_token=$(token_of "$JAR")
[ -n "$admin_token" ] || { echo "no admin csrf"; exit 1; }

suffix=$(date +%s)
username="editor$suffix"
password="Editor@123456"
headers="$TMP/headers"
code=$(curl -sS -b "$JAR" -c "$JAR" -D "$headers" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/users" \
  --data-urlencode "_csrf=$admin_token" --data-urlencode "username=$username" --data-urlencode "nickname=测试编辑" \
  --data-urlencode "email=$username@example.com" --data-urlencode "password=$password" \
  --data-urlencode "role_id=2" --data-urlencode "status=1")
[ "$code" = "302" ] || { echo "create editor failed: $code"; exit 1; }
user_id=$(tr -d '\r' < "$headers" | sed -n 's#.*/admin/users/\([0-9]*\)/edit.*#\1#p' | tail -1)
echo "editor id=$user_id"

login "$EDITOR_JAR" "$username" "$password"
for pair in "/admin:200" "/admin/articles:200" "/admin/settings:403" "/admin/users:403" "/admin/roles:403"; do
  path="${pair%%:*}"; want="${pair##*:}"
  got=$(curl -sS -b "$EDITOR_JAR" -c "$EDITOR_JAR" -o /dev/null -w '%{http_code}' "$BASE$path")
  if [ "$got" != "$want" ]; then echo "FAIL $path -> $got (want $want)"; exit 1; fi
  echo "ok   editor $path -> $got"
done

if [ -n "$user_id" ]; then
  admin_token=$(token_of "$JAR")
  code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/users/$user_id/delete" \
    --data-urlencode "_csrf=$admin_token" --data-urlencode "_method=DELETE")
  [ "$code" = "302" ] || { echo "cleanup editor failed: $code"; exit 1; }
  echo "ok   cleanup editor"
fi
