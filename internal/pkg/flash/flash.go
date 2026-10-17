// Package flash 在 session 中保存一次性提示消息。
package flash

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Message 是一条 flash 消息。
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

const sessionKey = "cms_flash"

// Add 写入一条提示。
func Add(c *gin.Context, typ, text string) {
	s := sessions.Default(c)
	var list []Message
	if v := s.Get(sessionKey); v != nil {
		if old, ok := v.([]Message); ok {
			list = old
		}
	}
	list = append(list, Message{Type: typ, Text: text})
	s.Set(sessionKey, list)
	_ = s.Save()
}

// Success 写入成功提示。
func Success(c *gin.Context, text string) { Add(c, "success", text) }

// Error 写入失败提示。
func Error(c *gin.Context, text string) { Add(c, "danger", text) }

// Warning 写入警告提示。
func Warning(c *gin.Context, text string) { Add(c, "warning", text) }

// Pop 取出并清空提示。
func Pop(c *gin.Context) []Message {
	s := sessions.Default(c)
	v := s.Get(sessionKey)
	if v == nil {
		return nil
	}
	list, _ := v.([]Message)
	s.Delete(sessionKey)
	_ = s.Save()
	return list
}
