package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cms/internal/service"
)

// CategoryHandler 处理分类管理。
type CategoryHandler struct {
	*Base
	Categories *service.CategoryService
}

// NewCategoryHandler 创建分类 handler。
func NewCategoryHandler(base *Base, categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{Base: base, Categories: categories}
}

// Index 渲染分类树。
func (h *CategoryHandler) Index(c *gin.Context) {
	h.render(c, "", nil)
}

// Create 创建分类。
func (h *CategoryHandler) Create(c *gin.Context) {
	in := categoryInput(c)
	if _, err := h.Categories.Create(c.Request.Context(), in); err != nil {
		h.render(c, h.ErrorMessage(err), &in)
		return
	}
	h.Success(c, "分类已创建", "/admin/categories")
}

// Update 更新分类。
func (h *CategoryHandler) Update(c *gin.Context) {
	id := uint64Param(c, "id")
	in := categoryInput(c)
	if _, err := h.Categories.Update(c.Request.Context(), id, in); err != nil {
		h.render(c, h.ErrorMessage(err), &in)
		return
	}
	h.Success(c, "分类已保存", "/admin/categories")
}

// Delete 删除分类。
func (h *CategoryHandler) Delete(c *gin.Context) {
	if err := h.Categories.Delete(c.Request.Context(), uint64Param(c, "id")); err != nil {
		h.Fail(c, err, "/admin/categories")
		return
	}
	h.Success(c, "分类已删除", "/admin/categories")
}

// Reorder 拖拽排序。
func (h *CategoryHandler) Reorder(c *gin.Context) {
	ids := parseIDs(c.PostFormArray("ids"))
	if err := h.Categories.Reorder(c.Request.Context(), ids); err != nil {
		h.Fail(c, err, "/admin/categories")
		return
	}
	h.JSON(c, gin.H{"ok": true})
}

func (h *CategoryHandler) render(c *gin.Context, message string, form *service.CategoryInput) {
	tree, err := h.Categories.Tree(c.Request.Context(), false)
	if err != nil {
		h.Fail(c, err, "/admin/categories")
		return
	}
	options, _ := h.Categories.Options(c.Request.Context())
	data := h.Data(c, "分类管理", "内容管理", "category")
	data["Tree"] = tree
	data["Options"] = options
	data["Error"] = message
	if form != nil {
		data["Form"] = form
	}
	data["QueryValues"] = c.Request.URL.Query()
	data["BaseURL"] = "/admin/categories"
	c.HTML(http.StatusOK, "admin/category/index.html", data)
}

func categoryInput(c *gin.Context) service.CategoryInput {
	status := int8(1)
	if c.PostForm("status") == "0" {
		status = 0
	}
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	return service.CategoryInput{
		ParentID:       uint64Param(c, "parent_id"),
		Name:           c.PostForm("name"),
		Slug:           c.PostForm("slug"),
		Description:    c.PostForm("description"),
		CoverURL:       c.PostForm("cover_url"),
		Sort:           sort,
		Status:         status,
		Template:       c.PostForm("template"),
		SeoTitle:       c.PostForm("seo_title"),
		SeoKeywords:    c.PostForm("seo_keywords"),
		SeoDescription: c.PostForm("seo_description"),
	}
}
