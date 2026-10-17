// Package response 提供统一的 JSON 响应结构。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 是统一响应体。
type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: data})
}

// Created 返回创建成功响应。
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Body{Code: 0, Msg: "created", Data: data})
}

// Fail 返回失败响应。
func Fail(c *gin.Context, status, code int, msg string) {
	c.AbortWithStatusJSON(status, Body{Code: code, Msg: msg})
}
