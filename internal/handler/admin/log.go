package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/pagination"
	"cms/internal/repository"
	"cms/internal/service"
)

// LogHandler 处理操作日志。
type LogHandler struct {
	*Base
	Logs  *repository.LogRepo
	Users *service.UserService
}

// NewLogHandler 创建日志 handler。
func NewLogHandler(base *Base, logs *repository.LogRepo, users *service.UserService) *LogHandler {
	return &LogHandler{Base: base, Logs: logs, Users: users}
}

// Index 渲染日志列表。
func (h *LogHandler) Index(c *gin.Context) {
	filter := repository.LogFilter{
		Query: pagination.Query{
			Page:     intParam(c, "page", 1),
			PageSize: intParam(c, "page_size", 20),
			Keyword:  c.Query("keyword"),
		},
		UserID: uint64Param(c, "user_id"),
		Module: c.Query("module"),
	}
	if raw := c.Query("from"); raw != "" {
		if t, err := time.ParseInLocation("2006-01-02", raw, h.Loc); err == nil {
			utc := t.UTC()
			filter.From = &utc
		}
	}
	if raw := c.Query("to"); raw != "" {
		if t, err := time.ParseInLocation("2006-01-02", raw, h.Loc); err == nil {
			utc := t.AddDate(0, 0, 1).UTC()
			filter.To = &utc
		}
	}
	list, total, err := h.Logs.List(c.Request.Context(), filter)
	if err != nil {
		h.Fail(c, err, "/admin/logs")
		return
	}
	users, _, _ := h.Users.List(c.Request.Context(), pagination.Query{Page: 1, PageSize: 100}, nil, 0)
	data := h.Data(c, "操作日志", "系统管理", "log")
	data["Data"] = pagination.New(list, total, filter.Page, filter.PageSize)
	data["Users"] = users
	data["Filter"] = filter
	data["QueryValues"] = c.Request.URL.Query()
	data["BaseURL"] = "/admin/logs"
	c.HTML(http.StatusOK, "admin/log/index.html", data)
}

// Clear 清空日志。
func (h *LogHandler) Clear(c *gin.Context) {
	if err := h.Logs.Clear(c.Request.Context()); err != nil {
		h.Fail(c, err, "/admin/logs")
		return
	}
	h.Success(c, "操作日志已清空", "/admin/logs")
}
