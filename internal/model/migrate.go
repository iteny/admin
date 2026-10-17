package model

// AllModels 返回 AutoMigrate 需要的全部模型，顺序按依赖关系排列。
func AllModels() []any {
	return []any{
		&Role{},
		&Permission{},
		&User{},
		&Menu{},
		&Category{},
		&Tag{},
		&Article{},
		&Page{},
		&Media{},
		&Setting{},
		&OperationLog{},
	}
}
