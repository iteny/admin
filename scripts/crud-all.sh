#!/usr/bin/env bash
# 全模块冒烟：标签、单页、文章更新、设置、媒体上传/删除。
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
JAR="$(mktemp)"
TMP="$(mktemp -d)"
trap 'rm -rf "$JAR" "$TMP"' EXIT

login_html=$(curl -sS -c "$JAR" "$BASE/admin/login")
token=$(printf '%s' "$login_html" | sed -n 's/.*name="_csrf" value="\([^"]*\)".*/\1/p' | head -1)
curl -sS -b "$JAR" -c "$JAR" -o /dev/null -X POST "$BASE/admin/login" \
  --data-urlencode "_csrf=$token" --data-urlencode "username=admin" --data-urlencode "password=Admin@123456"
token=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin" | sed -n 's/.*name="csrf-token" content="\([^"]*\)".*/\1/p' | head -1)
[ -n "$token" ] || { echo "no csrf"; exit 1; }

ts=$(date +%s)
cat_slug="all-cat-$ts"
tag_slug="all-tag-$ts"
art_slug="all-article-$ts"
page_slug="all-page-$ts"
cat_name="全模块分类$ts"
tag_name="全模块标签$ts"

post() { curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE$1" "${@:2}"; }

code=$(post /admin/categories --data-urlencode "_csrf=$token" --data-urlencode "name=$cat_name" --data-urlencode "slug=$cat_slug" --data-urlencode "status=1")
[ "$code" = "302" ] || { echo "category: $code"; exit 1; }
cat_html=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin/categories")
cat_id=$(printf '%s' "$cat_html" | sed -n "s/.*data-id=\"\([0-9]*\)\" data-parent=.*data-slug=\"$cat_slug\".*/\1/p" | head -1)
echo "category=$cat_id"

code=$(post /admin/tags --data-urlencode "_csrf=$token" --data-urlencode "name=$tag_name" --data-urlencode "slug=$tag_slug")
[ "$code" = "302" ] || { echo "tag: $code"; exit 1; }
tag_html=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin/tags")
tag_id=$(printf '%s' "$tag_html" | sed -n "s/.*data-id=\"\([0-9]*\)\" data-name=.*data-slug=\"$tag_slug\".*/\1/p" | head -1)
echo "tag=$tag_id"

code=$(post /admin/pages --data-urlencode "_csrf=$token" --data-urlencode "title=全模块单页$ts" --data-urlencode "slug=$page_slug" --data-urlencode "content=<p>单页正文</p>" --data-urlencode "status=published" --data-urlencode "sort=0")
[ "$code" = "302" ] || { echo "page: $code"; exit 1; }
page_code=$(curl -sS -o "$TMP/page.html" -w '%{http_code}' "$BASE/p/$page_slug")
[ "$page_code" = "200" ] || { echo "front page: $page_code"; exit 1; }
grep -q "全模块单页$ts" "$TMP/page.html"
echo "ok page"

headers="$TMP/art.headers"
code=$(curl -sS -b "$JAR" -c "$JAR" -D "$headers" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/articles" \
  --data-urlencode "_csrf=$token" --data-urlencode "title=全模块文章$ts" --data-urlencode "slug=$art_slug" \
  --data-urlencode "summary=摘要" --data-urlencode "content=<p>正文</p>" --data-urlencode "content_type=html" \
  --data-urlencode "category_id=$cat_id" --data-urlencode "status=published" --data-urlencode "tag_ids=$tag_id")
[ "$code" = "302" ] || { echo "article: $code"; exit 1; }
art_id=$(tr -d '\r' < "$headers" | sed -n 's#.*/admin/articles/\([0-9]*\)/edit.*#\1#p' | tail -1)
echo "article=$art_id"

tag_code=$(curl -sS -o "$TMP/tag.html" -w '%{http_code}' "$BASE/tag/$tag_slug")
[ "$tag_code" = "200" ] || { echo "tag page: $tag_code"; exit 1; }
grep -q "全模块文章$ts" "$TMP/tag.html"
echo "ok tag page"

code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/articles/$art_id" \
  --data-urlencode "_csrf=$token" --data-urlencode "_method=PUT" --data-urlencode "title=全模块文章更新$ts" \
  --data-urlencode "slug=$art_slug" --data-urlencode "summary=新摘要" --data-urlencode "content=<p>新正文</p>" \
  --data-urlencode "content_type=html" --data-urlencode "category_id=$cat_id" --data-urlencode "status=published" --data-urlencode "tag_ids=$tag_id")
[ "$code" = "302" ] || { echo "article update: $code"; exit 1; }
curl -sS -o "$TMP/art.html" "$BASE/article/$art_slug"
grep -q "全模块文章更新$ts" "$TMP/art.html"
echo "ok article update"

settings_code=$(curl -sS -b "$JAR" -c "$JAR" -o /dev/null -w '%{http_code}' -X POST "$BASE/admin/settings" \
  --data-urlencode "_csrf=$token" --data-urlencode "site_name=全模块CMS$ts" --data-urlencode "site_logo=" \
  --data-urlencode "site_favicon=" --data-urlencode "site_icp=" --data-urlencode "site_footer=footer" \
  --data-urlencode "site_description=desc" --data-urlencode "seo_title=seo" --data-urlencode "seo_keywords=kw" \
  --data-urlencode "seo_description=seo desc" --data-urlencode "upload_max_size=10485760" \
  --data-urlencode "upload_allowed_ext=jpg,jpeg,png,gif,webp,svg" --data-urlencode "mail_host=" \
  --data-urlencode "mail_port=587" --data-urlencode "mail_username=" --data-urlencode "mail_password=" --data-urlencode "mail_from=")
[ "$settings_code" = "302" ] || { echo "settings: $settings_code"; exit 1; }
curl -sS -o "$TMP/home.html" "$BASE/"
grep -q "全模块CMS$ts" "$TMP/home.html"
echo "ok settings"

printf 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Wl2n3sAAAAASUVORK5CYII=' | base64 -d > "$TMP/pixel.png"
upload=$(curl -sS -b "$JAR" -c "$JAR" -X POST "$BASE/admin/media/upload" -H "X-CSRF-Token: $token" -F "file=@$TMP/pixel.png")
echo "$upload" | grep -q '"code":0' || { echo "upload failed: $upload"; exit 1; }
media_url=$(printf '%s' "$upload" | sed -n 's/.*"url":"\([^"]*\)".*/\1/p')
media_id=$(printf '%s' "$upload" | sed -n 's/.*"id":\([0-9]*\).*/\1/p')
[ -n "$media_id" ] || { echo "media id missing"; exit 1; }
[ "$(curl -sS -o /dev/null -w '%{http_code}' "$BASE$media_url")" = "200" ] || { echo "media not served"; exit 1; }
echo "ok media upload $media_url"

# 清理
post /admin/articles/$art_id/delete --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE" >/dev/null
# 通过 slug 找到单页 ID（列表页）
page_html=$(curl -sS -b "$JAR" -c "$JAR" "$BASE/admin/pages?keyword=$page_slug")
page_id=$(printf '%s' "$page_html" | sed -n "s#.*href=\"/admin/pages/\([0-9]*\)/edit\".*#\1#p" | head -1)
[ -n "$page_id" ] && post /admin/pages/$page_id/delete --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE" >/dev/null
post /admin/categories/$cat_id/delete --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE" >/dev/null
post /admin/tags/$tag_id/delete --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE" >/dev/null
post /admin/media/$media_id/delete --data-urlencode "_csrf=$token" --data-urlencode "_method=DELETE" >/dev/null
echo "ok cleanup"
