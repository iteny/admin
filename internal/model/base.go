package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 是所有业务表共用的基础字段。
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"precision:3" json:"created_at"`
	UpdatedAt time.Time      `gorm:"precision:3" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
