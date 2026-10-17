package model

import "time"

// OperationLog 是后台操作日志，不参与软删除。
type OperationLog struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"index;not null;default:0" json:"user_id"`
	Username   string    `gorm:"size:50;index" json:"username"`
	Module     string    `gorm:"size:50;index" json:"module"`
	Action     string    `gorm:"size:100" json:"action"`
	Method     string    `gorm:"size:10" json:"method"`
	Path       string    `gorm:"size:255" json:"path"`
	IP         string    `gorm:"size:45" json:"ip"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	Request    string    `gorm:"type:text" json:"request"`
	StatusCode int       `gorm:"not null;default:0" json:"status_code"`
	DurationMs int       `gorm:"not null;default:0" json:"duration_ms"`
	CreatedAt  time.Time `gorm:"precision:3;index" json:"created_at"`
}
