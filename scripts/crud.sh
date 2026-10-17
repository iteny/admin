#!/usr/bin/env bash
# 内容闭环冒烟测试：登录 -> 建分类 -> 建文章 -> 发布 -> 前台可见 -> 删除。
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
JAR="$(mktemp)"
TMP="$(mktemp -d)"
trap 'rm -rf "$JAR" "$TMP"' EXIT

meta=$(curl -sS -c "$JAR" "$BASE/admin/login")
token=$(printf '%s' "$meta" | sed -n 's/.*name="_csrf" value="\([^"]*\)".*/\1/p' | head -1)
[ -n "$token" ] || { echo "no csrf token"; exit 1; }

curl -sS -b "$JAR" -c "$JAR" -o /dev/null -X POST "$BASE/admin/login" \
  --data-urlencode "_csrf=$token" --data-urlencode "username=admin" --data-urlencode "password=Admin@123456"

# 取后台页面里的 meta token
dashboard=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin")
token=$(printf '%s' "$dashboard" | sed -n 's/.*name="csrf-token" content="\([^"]*\)".*/\1/p' | head -1)
[ -n "$token" ] || { echo "no meta csrf token"; exit 1; }

suffix=$(date +%s)
cat_slug="smoke-cat-$suffix"
art_slug="smoke-article-$suffix"

code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/categories" \
  --data-urlencode "_csrf=$token" --data-urlencode "name=冒烟分类" --data-urlencode "slug=$cat_slug" \
  --data-urlencode "status=1" --data-urlencode "sort=0")
[ "$code" = "302" ] || { echo "create category failed: $code"; exit 1; }

cat_page=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin/categories")
cat_id=$(printf '%s' "$cat_page" | sed -n "s/.*data-id=\"\([0-9]*\)\" data-parent=.*data-slug=\"$cat_slug\".*/\1/p" | head -1)
[ -n "$cat_id" ] || { echo "category id not found"; exit 1; }
echo "category id=$cat_id"

headers=$(mktemp)
code=$(curl -sS -b "$JAR" -c "$JAR" -D "$headers" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/articles" \
  --data-urlencode "_csrf=$token" --data-urlencode "title=冒烟文章 $suffix" --data-urlencode "slug=$art_slug" \
  --data-urlencode "summary=这是一篇冒烟测试文章" --data-urlencode "content=<p>Hello <strong>Tabler CMS</strong></p>" \
  --data-urlencode "content_type=html" --data-urlencode "category_id=$cat_id" --data-urlencode "status=published" \
  --data-urlencode "is_top=on" --data-urlencode "is_recommend=on")
[ "$code" = "302" ] || { echo "create article failed: $code"; exit 1; }
loc=$(tr -d '\r' < "$headers" | sed -n 's/^[Ll]ocation: \(.*\)$/\1/p' | tail -1)
echo "created: $loc"

code=$(curl -sS -b "$JAR" -c "$JAR" -o "$TMP/article.html" -w '%{http_code}' "$BASE/article/$art_slug")
[ "$code" = "200" ] || { echo "front article failed: $code"; exit 1; }
grep -q "冒烟文章 $suffix" "$TMP/article.html" || { echo "article title not rendered"; exit 1; }
echo "ok front article"

art_id=$(printf '%s' "$loc" | sed -n 's#.*/admin/articles/\([0-9]*\)/edit.*#\1#p')
if [ -n "$art_id" ]; then
  code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/articles/$art_id/delete" \
    --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE")
  [ "$code" = "302" ] || { echo "delete article failed: $code"; exit 1; }
  code=$(curl -sS -o /dev/null -w '%{http_code}' "$BASE/article/$art_slug")
  [ "$code" = "404" ] || { echo "article still visible: $code"; exit 1; }
  echo "ok article delete"
fi

cat_code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/categories/$cat_id/delete" \
  --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE")
[ "$cat_code" = "302" ] || { echo "delete category failed: $cat_code"; exit 1; }
echo "ok category delete"
