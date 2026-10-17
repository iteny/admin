package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
	"cms/internal/pkg/pagination"
	"cms/internal/pkg/response"
	"cms/internal/service"
)

// MediaHandler 处理媒体库。
type MediaHandler struct {
	*Base
	Media *service.MediaService
}

// NewMediaHandler 创建媒体 handler。
func NewMediaHandler(base *Base, media *service.MediaService) *MediaHandler {
	return &MediaHandler{Base: base, Media: media}
}

// Index 渲染媒体库。
func (h *MediaHandler) Index(c *gin.Context) {
	h.renderIndex(c, "admin/media/index.html")
}

// Picker 渲染媒体选择器片段。
func (h *MediaHandler) Picker(c *gin.Context) {
	h.renderIndex(c, "admin/media/picker.html")
}

func (h *MediaHandler) renderIndex(c *gin.Context, template string) {
	page := intParam(c, "page", 1)
	pageSize := intParam(c, "page_size", 24)
	list, total, err := h.Media.List(c.Request.Context(), page, pageSize, c.Query("keyword"), c.Query("mime"))
	if err != nil {
		h.Fail(c, err, "/admin/media")
		return
	}
	data := h.Data(c, "媒体库", "内容管理", "media")
	data["Data"] = pagination.New(list, total, page, pageSize)
	data["QueryValues"] = c.Request.URL.Query()
	data["BaseURL"] = "/admin/media"
	data["Picker"] = template == "admin/media/picker.html"
	c.HTML(http.StatusOK, template, data)
}

// Upload 接收文件并返回 JSON。
func (h *MediaHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.Config.App.UploadMaxSize+1<<20)
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "请选择要上传的文件，且文件不能超过大小限制")
		return
	}
	user := middleware.CurrentUser(c)
	if user == nil {
		response.Fail(c, http.StatusUnauthorized, 40100, "登录已过期")
		return
	}
	item, err := h.Media.Upload(c.Request.Context(), file, user.ID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, h.ErrorMessage(err))
		return
	}
	response.OK(c, gin.H{
		"id":            item.ID,
		"url":           item.URL,
		"thumb_url":     thumbURL(item.URL),
		"original_name": item.OriginalName,
		"size":          item.Size,
	})
}

// Delete 删除媒体。
func (h *MediaHandler) Delete(c *gin.Context) {
	if err := h.Media.Delete(c.Request.Context(), uint64Param(c, "id")); err != nil {
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			response.Fail(c, http.StatusBadRequest, 40000, h.ErrorMessage(err))
			return
		}
		h.Fail(c, err, "/admin/media")
		return
	}
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		h.JSON(c, gin.H{"ok": true})
		return
	}
	h.Success(c, "媒体已删除", "/admin/media")
}

func thumbURL(url string) string {
	ext := ""
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '.' {
			ext = url[i:]
			break
		}
		if url[i] == '/' {
			break
		}
	}
	if ext == "" || strings.HasSuffix(url, ".svg") {
		return url
	}
	return url[:len(url)-len(ext)] + "_thumb" + ext
}
