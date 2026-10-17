package model

// Setting 是键值配置。
type Setting struct {
	BaseModel
	GroupName string `gorm:"size:50;index" json:"group_name"`
	Key       string `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value     string `gorm:"type:text" json:"value"`
	Type      string `gorm:"size:20;not null;default:'string'" json:"type"`
	Remark    string `gorm:"size:255" json:"remark"`
}
