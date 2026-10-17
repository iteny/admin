package web

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SEOHandler 生成 sitemap.xml 和 robots.txt。
type SEOHandler struct{ *Base }

// NewSEOHandler 创建 SEO handler。
func NewSEOHandler(base *Base) *SEOHandler { return &SEOHandler{Base: base} }

// Sitemap 输出 sitemap.xml。
func (h *SEOHandler) Sitemap(c *gin.Context) {
	articles, pages, categories, tags, err := h.Front.Sitemap(c.Request.Context())
	if err != nil {
		h.ServerError(c, "生成 sitemap 失败")
		return
	}
	base := h.baseURL(c)
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	writeURL(&buf, base+"/", time.Now())
	for _, item := range categories {
		writeURL(&buf, base+"/category/"+item.Slug, item.UpdatedAt)
	}
	for _, item := range tags {
		writeURL(&buf, base+"/tag/"+item.Slug, item.UpdatedAt)
	}
	for _, item := range articles {
		last := item.UpdatedAt
		if item.PublishedAt != nil {
			last = *item.PublishedAt
		}
		writeURL(&buf, base+"/article/"+item.Slug, last)
	}
	for _, item := range pages {
		writeURL(&buf, base+"/p/"+item.Slug, item.UpdatedAt)
	}
	buf.WriteString("</urlset>\n")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", buf.Bytes())
}

// Robots 输出 robots.txt。
func (h *SEOHandler) Robots(c *gin.Context) {
	base := h.baseURL(c)
	body := "User-agent: *\nAllow: /\nDisallow: /admin/\nSitemap: " + base + "/sitemap.xml\n"
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
}

func (h *SEOHandler) baseURL(c *gin.Context) string {
	base := strings.TrimRight(h.Config.App.BaseURL, "/")
	if base != "" {
		return base
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func writeURL(buf *bytes.Buffer, loc string, last time.Time) {
	buf.WriteString("<url><loc>")
	_ = xml.EscapeText(buf, []byte(loc))
	buf.WriteString("</loc>")
	if !last.IsZero() {
		buf.WriteString("<lastmod>")
		buf.WriteString(last.UTC().Format("2006-01-02"))
		buf.WriteString("</lastmod>")
	}
	buf.WriteString("</url>\n")
}
