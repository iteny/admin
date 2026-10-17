package model

import "time"

// ArticleStatus 是文章状态。
type ArticleStatus string

const (
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"

	// ArticleStatusDraft 仅用于兼容历史数据；新文章不再使用草稿状态。
	ArticleStatusDraft ArticleStatus = "draft"
)

// ContentType 是正文类型。
type ContentType string

const (
	ContentTypeHTML     ContentType = "html"
	ContentTypeMarkdown ContentType = "markdown"
)

// Article 是文章。
type Article struct {
	BaseModel
	Title          string        `gorm:"size:200;not null" json:"title"`
	Slug           string        `gorm:"size:200;uniqueIndex;not null" json:"slug"`
	Summary        string        `gorm:"size:500" json:"summary"`
	Content        string        `gorm:"type:longtext" json:"content"`
	SubContent     string        `gorm:"type:longtext" json:"sub_content"`
	ContentType    ContentType   `gorm:"size:10;not null;default:'html'" json:"content_type"`
	CoverURL       string        `gorm:"size:255" json:"cover_url"`
	CategoryID     uint64        `gorm:"index;not null;default:0" json:"category_id"`
	Category       Category      `gorm:"foreignKey:CategoryID" json:"category"`
	AuthorID       uint64        `gorm:"index;not null;default:0" json:"author_id"`
	Author         User          `gorm:"foreignKey:AuthorID" json:"author"`
	Status         ArticleStatus `gorm:"size:20;not null;default:'published';index:idx_articles_status_published,priority:1" json:"status"`
	IsTop          bool          `gorm:"not null;default:false;index" json:"is_top"`
	IsRecommend    bool          `gorm:"not null;default:false;index" json:"is_recommend"`
	ViewCount      int           `gorm:"not null;default:0" json:"view_count"`
	CommentCount   int           `gorm:"not null;default:0" json:"comment_count"`
	PublishedAt    *time.Time    `gorm:"index:idx_articles_status_published,priority:2" json:"published_at"`
	Tags           []Tag         `gorm:"many2many:article_tags;" json:"tags"`
	SeoTitle       string        `gorm:"size:100" json:"seo_title"`
	SeoKeywords    string        `gorm:"size:200" json:"seo_keywords"`
	SeoDescription string        `gorm:"size:255" json:"seo_description"`
}
