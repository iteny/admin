package model

// Role 是 RBAC 角色。
type Role struct {
	BaseModel
	Name        string       `gorm:"size:50;not null" json:"name"`
	Slug        string       `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Description string       `gorm:"size:255" json:"description"`
	IsSystem    bool         `gorm:"not null;default:false" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Permission 是一个权限点，Code 形如 article:create。
type Permission struct {
	BaseModel
	Name      string `gorm:"size:50;not null" json:"name"`
	Code      string `gorm:"size:100;uniqueIndex;not null" json:"code"`
	GroupName string `gorm:"size:50;index" json:"group_name"`
	Sort      int    `gorm:"not null;default:0" json:"sort"`
}
