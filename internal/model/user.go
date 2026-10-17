package model

import "time"

// 用户状态。
const (
	UserStatusDisabled int8 = 0
	UserStatusEnabled  int8 = 1
)

// User 是后台管理员/作者。
type User struct {
	BaseModel
	Username    string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password    string     `gorm:"size:100;not null" json:"-"`
	Nickname    string     `gorm:"size:50" json:"nickname"`
	Email       string     `gorm:"size:100;uniqueIndex" json:"email"`
	Avatar      string     `gorm:"size:255" json:"avatar"`
	Status      int8       `gorm:"type:tinyint;not null;default:1;index" json:"status"`
	RoleID      uint64     `gorm:"index;not null;default:0" json:"role_id"`
	Role        Role       `gorm:"foreignKey:RoleID" json:"role"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `gorm:"size:45" json:"last_login_ip"`
	Remark      string     `gorm:"size:255" json:"remark"`
}
