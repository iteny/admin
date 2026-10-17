# 移动端文章 API

面向 Android / iOS 的只读 JSON API，默认不需要登录。

## 基础信息

```text
Base URL: https://cms.example.com/api/v1
```

统一响应结构：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

`code = 0` 表示成功；失败时 HTTP 状态码和 `code` 都会体现错误，例如：

```json
{
  "code": 40400,
  "msg": "文章不存在或未发布"
}
```

所有时间均为 RFC3339，例如：

```text
2026-10-17T14:16:34Z
```

## 接口列表

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/site` | 站点名称、描述、版本 |
| GET | `/home` | App 首页聚合数据 |
| GET | `/articles` | 已发布文章列表 |
| GET | `/search` | 搜索文章，参数同文章列表 |
| GET | `/articles/:slug` | 按 slug 查询文章详情 |
| GET | `/articles/id/:id` | 按文章 ID 查询文章详情 |
| GET | `/categories` | 启用分类树 |
| GET | `/tags` | 标签列表 |

## 文章列表

```http
GET /api/v1/articles?page=1&page_size=10
```

列表默认只返回文章摘要，不返回正文，避免列表响应过大。需要正文时先调用文章详情接口，或者显式传入：

```http
GET /api/v1/articles?page=1&page_size=10&include_content=1
```

使用 `include_content=1` 时，每一项会额外包含：

```json
{
  "content_html": "<p>正文</p>",
  "sub_content_html": "<p>副文</p>",
  "content_base_url": "https://cms.example.com"
}
```

支持参数：

| 参数 | 说明 |
| --- | --- |
| `page` | 页码，默认 1 |
| `page_size` | 每页数量，默认 10，最大 50 |
| `category` | 分类 slug，包含子分类 |
| `category_id` | 分类 ID |
| `tag` | 标签 slug |
| `tag_id` | 标签 ID |
| `q` / `keyword` | 搜索标题或摘要 |
| `sort` | `latest`、`oldest`、`views`、`updated` |
| `featured` | `1` 只返回推荐文章 |
| `top` | `1` 只返回置顶文章 |
| `include_content` | `1` 同时返回 `content_html` 和 `sub_content_html` |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "文章标题",
        "slug": "article-slug",
        "summary": "文章摘要",
        "cover_url": "https://cms.example.com/uploads/2026/10/cover.png",
        "category": {
          "id": 1,
          "name": "文章",
          "slug": "wen-zhang"
        },
        "tags": [
          { "id": 1, "name": "Go", "slug": "go" }
        ],
        "author": {
          "id": 1,
          "nickname": "管理员"
        },
        "is_top": false,
        "is_recommend": true,
        "view_count": 12,
        "published_at": "2026-10-17T14:16:34Z",
        "url": "https://cms.example.com/article/article-slug"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

## 文章详情

按 slug 查询：

```http
GET /api/v1/articles/article-slug
```

按文章 ID 查询：

```http
GET /api/v1/articles/id/123
```

`id` 必须是正整数；不存在、已归档或未发布的文章返回 404。

详情在列表字段基础上增加：

| 字段 | 说明 |
| --- | --- |
| `content_html` | 已清洗的 HTML 正文 |
| `sub_content_html` | 正文下方的副文；没有副文时为空字符串 |
| `content_type` | `html` 或 `markdown` |
| `content_base_url` | WebView 加载正文时的 base URL |
| `seo_title` | SEO 标题 |
| `seo_keywords` | SEO 关键词 |
| `seo_description` | SEO 描述 |
| `prev` | 上一篇摘要，可能为空 |
| `next` | 下一篇摘要，可能为空 |
| `related` | 相关文章摘要列表 |

Android WebView 建议这样加载正文：

```kotlin
webView.loadDataWithBaseURL(
    article.contentBaseUrl,
    article.contentHtml,
    "text/html",
    "UTF-8",
    null
)
```

这样正文中的 `/uploads/...` 图片会自动解析为完整地址。

## 分类

```http
GET /api/v1/categories
```

返回树形结构，`children` 始终是数组：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "parent_id": 0,
        "name": "文章",
        "slug": "wen-zhang",
        "description": "",
        "sort": 0,
        "article_count": 12,
        "children": []
      }
    ]
  }
}
```

## 标签

```http
GET /api/v1/tags
```

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      { "id": 1, "name": "Go", "slug": "go", "use_count": 5 }
    ]
  }
}
```

## 首页聚合

```http
GET /api/v1/home?page=1&page_size=10
```

返回：

```json
{
  "recommended": [],
  "latest": { "list": [], "total": 0, "page": 1, "page_size": 10, "total_pages": 1 },
  "top": [],
  "hot": [],
  "categories": [],
  "tags": []
}
```

## 只读范围

- 只返回 `published` 状态文章；
- `archived` 和历史草稿不会出现在列表或详情中；
- 详情接口会增加一次浏览量；
- 公共接口不需要 token；如果 App 后续需要登录、收藏、评论或发布文章，建议新增 `/api/v1/auth/*` 或 `/api/v1/admin/*`，使用独立的 Bearer Token，而不是复用后台 Cookie Session。
