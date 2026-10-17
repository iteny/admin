package model

// PageStatus 是单页状态。
type PageStatus string

const (
	PageStatusPublished PageStatus = "published"
	PageStatusArchived  PageStatus = "archived"

	// PageStatusDraft 仅用于兼容历史数据；新单页不再使用草稿状态。
	PageStatusDraft PageStatus = "draft"
)

// Page 是单页（关于我们、联系方式等）。
type Page struct {
	BaseModel
	Title          string     `gorm:"size:100;not null" json:"title"`
	Slug           string     `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Content        string     `gorm:"type:longtext" json:"content"`
	Status         PageStatus `gorm:"size:20;not null;default:'published';index" json:"status"`
	Sort           int        `gorm:"not null;default:0;index" json:"sort"`
	SeoTitle       string     `gorm:"size:100" json:"seo_title"`
	SeoKeywords    string     `gorm:"size:200" json:"seo_keywords"`
	SeoDescription string     `gorm:"size:255" json:"seo_description"`
}
