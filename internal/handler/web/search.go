package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/pagination"
)

// SearchHandler 处理前台搜索。
type SearchHandler struct{ *Base }

// NewSearchHandler 创建搜索 handler。
func NewSearchHandler(base *Base) *SearchHandler { return &SearchHandler{Base: base} }

// Index 渲染搜索结果。
func (h *SearchHandler) Index(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("q"))
	page := pageParam(c)
	list, total, totalPages, err := h.Front.Search(c.Request.Context(), keyword, page, 9)
	if err != nil {
		h.ServerError(c, "搜索失败")
		return
	}
	data := h.Data(c, "搜索："+keyword, "搜索文章")
	data["Keyword"] = keyword
	data["Articles"] = list
	data["Total"] = total
	data["Pager"] = pagination.NewView(page, totalPages, func(p int) string {
		values := url.Values{"q": {keyword}}
		if p <= 1 {
			return "/search?" + values.Encode()
		}
		values.Set("page", itoa(p))
		return "/search?" + values.Encode()
	})
	c.HTML(http.StatusOK, "web/search.html", data)
}

func itoa(n int) string {
	if n <= 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
