package model

// Tag 是文章标签。
type Tag struct {
	BaseModel
	Name     string `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug     string `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	UseCount int    `gorm:"not null;default:0;index" json:"use_count"`
}
