package model

// Menu 是后台侧边栏菜单项。
type Menu struct {
	BaseModel
	ParentID       uint64 `gorm:"index;not null;default:0" json:"parent_id"`
	Name           string `gorm:"size:50;not null" json:"name"`
	Icon           string `gorm:"size:50" json:"icon"`
	Path           string `gorm:"size:255" json:"path"`
	PermissionCode string `gorm:"size:100" json:"permission_code"`
	Sort           int    `gorm:"not null;default:0;index" json:"sort"`
	IsVisible      bool   `gorm:"not null;default:true" json:"is_visible"`
	IsExternal     bool   `gorm:"not null;default:false" json:"is_external"`
}
