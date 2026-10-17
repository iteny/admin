package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cms/internal/model"
	"cms/internal/pkg/pagination"
	"cms/internal/service"
)

// PageHandler 处理单页管理。
type PageHandler struct {
	*Base
	Pages *service.PageService
}

// NewPageHandler 创建单页 handler。
func NewPageHandler(base *Base, pages *service.PageService) *PageHandler {
	return &PageHandler{Base: base, Pages: pages}
}

// Index 渲染单页列表。
func (h *PageHandler) Index(c *gin.Context) {
	q := pagination.Query{
		Page:     intParam(c, "page", 1),
		PageSize: intParam(c, "page_size", 15),
		Keyword:  c.Query("keyword"),
	}
	list, total, err := h.Pages.List(c.Request.Context(), q)
	if err != nil {
		h.Fail(c, err, "/admin/pages")
		return
	}
	data := h.Data(c, "单页管理", "内容管理", "page")
	data["Data"] = pagination.New(list, total, q.Page, q.PageSize)
	data["QueryValues"] = c.Request.URL.Query()
	data["BaseURL"] = "/admin/pages"
	c.HTML(http.StatusOK, "admin/page/index.html", data)
}

// CreatePage 渲染新增页。
func (h *PageHandler) CreatePage(c *gin.Context) {
	h.renderForm(c, nil, "")
}

// Create 创建单页。
func (h *PageHandler) Create(c *gin.Context) {
	in := pageInput(c)
	page, err := h.Pages.Create(c.Request.Context(), in)
	if err != nil {
		h.renderForm(c, pageFromInput(0, in), h.ErrorMessage(err))
		return
	}
	h.Success(c, "单页已创建", "/admin/pages/"+strconv.FormatUint(page.ID, 10)+"/edit")
}

// EditPage 渲染编辑页。
func (h *PageHandler) EditPage(c *gin.Context) {
	item, err := h.Pages.Get(c.Request.Context(), uint64Param(c, "id"))
	if err != nil {
		h.Fail(c, err, "/admin/pages")
		return
	}
	h.renderForm(c, item, "")
}

// Update 更新单页。
func (h *PageHandler) Update(c *gin.Context) {
	id := uint64Param(c, "id")
	in := pageInput(c)
	item, err := h.Pages.Update(c.Request.Context(), id, in)
	if err != nil {
		h.renderForm(c, pageFromInput(id, in), h.ErrorMessage(err))
		return
	}
	h.Success(c, "单页已保存", "/admin/pages/"+strconv.FormatUint(item.ID, 10)+"/edit")
}

// Delete 删除单页。
func (h *PageHandler) Delete(c *gin.Context) {
	if err := h.Pages.Delete(c.Request.Context(), uint64Param(c, "id")); err != nil {
		h.Fail(c, err, "/admin/pages")
		return
	}
	h.Success(c, "单页已删除", "/admin/pages")
}

func (h *PageHandler) renderForm(c *gin.Context, item *model.Page, message string) {
	if item == nil {
		item = &model.Page{Status: model.PageStatusPublished}
	}
	if item.Status == model.PageStatusDraft {
		// 兼容未迁移的历史草稿：编辑页按归档展示。
		item.Status = model.PageStatusArchived
	}
	title := "新增单页"
	if item.ID > 0 {
		title = "编辑单页"
	}
	data := h.Data(c, title, "内容管理", "page")
	data["Page"] = item
	data["Error"] = message
	data["IsEdit"] = item.ID > 0
	data["ActionURL"] = "/admin/pages"
	if item.ID > 0 {
		data["ActionURL"] = "/admin/pages/" + strconv.FormatUint(item.ID, 10)
	}
	c.HTML(http.StatusOK, "admin/page/form.html", data)
}

func pageInput(c *gin.Context) service.PageInput {
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	return service.PageInput{
		Title:          c.PostForm("title"),
		Slug:           c.PostForm("slug"),
		Content:        c.PostForm("content"),
		Status:         c.PostForm("status"),
		Sort:           sort,
		SeoTitle:       c.PostForm("seo_title"),
		SeoKeywords:    c.PostForm("seo_keywords"),
		SeoDescription: c.PostForm("seo_description"),
	}
}

func pageFromInput(id uint64, in service.PageInput) *model.Page {
	return &model.Page{
		BaseModel: model.BaseModel{ID: id},
		Title:     in.Title, Slug: in.Slug, Content: in.Content,
		Status: model.PageStatus(in.Status), Sort: in.Sort,
		SeoTitle: in.SeoTitle, SeoKeywords: in.SeoKeywords, SeoDescription: in.SeoDescription,
	}
}
