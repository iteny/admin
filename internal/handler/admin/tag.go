package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/service"
)

// TagHandler 处理标签管理。
type TagHandler struct {
	*Base
	Tags *service.TagService
}

// NewTagHandler 创建标签 handler。
func NewTagHandler(base *Base, tags *service.TagService) *TagHandler {
	return &TagHandler{Base: base, Tags: tags}
}

// Index 渲染标签列表。
func (h *TagHandler) Index(c *gin.Context) {
	h.render(c, "", nil)
}

// Create 创建标签。
func (h *TagHandler) Create(c *gin.Context) {
	in := tagInput(c)
	if _, err := h.Tags.Create(c.Request.Context(), in); err != nil {
		h.render(c, h.ErrorMessage(err), &in)
		return
	}
	h.Success(c, "标签已创建", "/admin/tags")
}

// Update 更新标签。
func (h *TagHandler) Update(c *gin.Context) {
	in := tagInput(c)
	if _, err := h.Tags.Update(c.Request.Context(), uint64Param(c, "id"), in); err != nil {
		h.render(c, h.ErrorMessage(err), &in)
		return
	}
	h.Success(c, "标签已保存", "/admin/tags")
}

// Delete 删除标签。
func (h *TagHandler) Delete(c *gin.Context) {
	if err := h.Tags.Delete(c.Request.Context(), uint64Param(c, "id")); err != nil {
		h.Fail(c, err, "/admin/tags")
		return
	}
	h.Success(c, "标签已删除", "/admin/tags")
}

func (h *TagHandler) render(c *gin.Context, message string, form *service.TagInput) {
	tags, err := h.Tags.List(c.Request.Context())
	if err != nil {
		h.Fail(c, err, "/admin/tags")
		return
	}
	data := h.Data(c, "标签管理", "内容管理", "tag")
	data["Tags"] = tags
	data["Error"] = message
	if form != nil {
		data["Form"] = form
	}
	c.HTML(http.StatusOK, "admin/tag/index.html", data)
}

func tagInput(c *gin.Context) service.TagInput {
	return service.TagInput{Name: c.PostForm("name"), Slug: c.PostForm("slug")}
}
