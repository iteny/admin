package model

// 分类状态。
const (
	CategoryStatusDisabled int8 = 0
	CategoryStatusEnabled  int8 = 1
)

// Category 是无限级内容分类。
type Category struct {
	BaseModel
	ParentID       uint64 `gorm:"index;not null;default:0" json:"parent_id"`
	Name           string `gorm:"size:50;not null" json:"name"`
	Slug           string `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description    string `gorm:"size:255" json:"description"`
	CoverURL       string `gorm:"size:255" json:"cover_url"`
	Sort           int    `gorm:"not null;default:0;index" json:"sort"`
	Status         int8   `gorm:"type:tinyint;not null;default:1;index" json:"status"`
	Template       string `gorm:"size:50" json:"template"`
	SeoTitle       string `gorm:"size:100" json:"seo_title"`
	SeoKeywords    string `gorm:"size:200" json:"seo_keywords"`
	SeoDescription string `gorm:"size:255" json:"seo_description"`
}
