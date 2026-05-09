package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageResponse 分页响应
type PageResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Total     int64  `json:"total"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

// OK 成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Created 创建成功
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "created", Data: data})
}

// Page 分页响应
func Page(c *gin.Context, data any, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, PageResponse{
		Code: 0, Message: "ok", Data: data,
		Total: total, Page: page, PageSize: pageSize,
	})
}

// Fail 业务错误
func Fail(c *gin.Context, status int, msg string) {
	c.JSON(status, Response{Code: status, Message: msg})
}

// Abort 中断请求（不继续执行后续 handler）
func Abort(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, Response{Code: status, Message: msg})
}
