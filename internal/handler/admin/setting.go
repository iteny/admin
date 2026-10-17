package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/service"
)

// SettingHandler 处理系统设置。
type SettingHandler struct {
	*Base
	Settings *service.SettingService
}

// NewSettingHandler 创建设置 handler。
func NewSettingHandler(base *Base, settings *service.SettingService) *SettingHandler {
	return &SettingHandler{Base: base, Settings: settings}
}

var settingKeys = map[string][]string{
	"site":   {"site_name", "site_logo", "site_favicon", "site_icp", "site_footer", "site_description"},
	"seo":    {"seo_title", "seo_keywords", "seo_description"},
	"upload": {"upload_max_size", "upload_allowed_ext"},
	"mail":   {"mail_host", "mail_port", "mail_username", "mail_password", "mail_from"},
}

// Index 渲染设置页。
func (h *SettingHandler) Index(c *gin.Context) {
	items, err := h.Settings.List(c.Request.Context())
	if err != nil {
		h.Fail(c, err, "/admin/settings")
		return
	}
	data := h.Data(c, "系统设置", "系统管理", "setting")
	data["Items"] = items
	data["Values"] = h.Settings.All()
	data["Keys"] = settingKeys
	c.HTML(http.StatusOK, "admin/setting/index.html", data)
}

// Save 保存设置。
func (h *SettingHandler) Save(c *gin.Context) {
	for group, keys := range settingKeys {
		values := make(map[string]string, len(keys))
		for _, key := range keys {
			values[key] = c.PostForm(key)
		}
		if err := h.Settings.Save(c.Request.Context(), values, group); err != nil {
			h.Fail(c, err, "/admin/settings")
			return
		}
	}
	h.Success(c, "设置已保存", "/admin/settings")
}
