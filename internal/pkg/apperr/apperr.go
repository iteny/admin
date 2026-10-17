// Package apperr 定义领域错误，避免 handler 用字符串或 panic 表达业务失败。
package apperr

import "errors"

var (
	ErrNotFound     = errors.New("资源不存在")
	ErrForbidden    = errors.New("没有权限执行此操作")
	ErrUnauthorized = errors.New("请先登录")
	ErrConflict     = errors.New("资源冲突")
	ErrValidation   = errors.New("参数校验失败")
)

// ValidationError 携带字段级校验信息。
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

// NewValidation 创建字段校验错误。
func NewValidation(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
