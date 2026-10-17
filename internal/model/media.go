package model

// Media 是媒体库文件。
type Media struct {
	BaseModel
	OriginalName string `gorm:"size:255;not null" json:"original_name"`
	StoredName   string `gorm:"size:255;not null" json:"stored_name"`
	Path         string `gorm:"size:255;not null" json:"path"`
	URL          string `gorm:"size:255;not null" json:"url"`
	MimeType     string `gorm:"size:100" json:"mime_type"`
	Ext          string `gorm:"size:20" json:"ext"`
	Size         int64  `gorm:"not null;default:0" json:"size"`
	Width        int    `gorm:"not null;default:0" json:"width"`
	Height       int    `gorm:"not null;default:0" json:"height"`
	Hash         string `gorm:"size:64;index" json:"hash"`
	UploaderID   uint64 `gorm:"index;not null;default:0" json:"uploader_id"`
}
