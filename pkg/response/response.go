package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorObject 更标准的错误结构（Stripe / AWS 风格）
type ErrorObject struct {
	Type      string      `json:"type"`                 // 错误类型（机器识别）
	Message   string      `json:"message"`              // 给用户看的
	Errors    interface{} `json:"errors,omitempty"`     // 字段级错误
	RequestID string      `json:"request_id,omitempty"` // 链路追踪
}

// ErrorResponse 包一层 error
type ErrorResponse struct {
	Error ErrorObject `json:"error"`
}

// --- 成功 ---
// 保持 RESTful，不包装
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// TooManyRequests 返回 429 状态码
func TooManyRequests(c *gin.Context, msg string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"code": http.StatusTooManyRequests,
		"msg":  msg,
	})
}

// --- 错误核心 ---
// 最底层方法
// errs为错误细节，举例：
// 当错误需要细化时，就有用了。
// ✅ 例子：表单校验错误
// errs := map[string]string{
// 	"username": "不能为空",
// 	"password": "长度至少6位",
// }
// errorResponse(c, 400, "validation_error", "参数校验失败", errs)

func errorResponse(c *gin.Context, status int, errType, msg string, errs interface{}) {
	rid, _ := c.Get("requestID")

	c.JSON(status, ErrorResponse{
		Error: ErrorObject{
			Type:      errType,
			Message:   msg,
			Errors:    errs,
			RequestID: toString(rid),
		},
	})
}

// --- 常用错误封装（推荐保留） ---

func BadRequest(c *gin.Context, msg string) {
	errorResponse(c, http.StatusBadRequest, "invalid_request", msg, nil)
}

func Unauthorized(c *gin.Context, msg string) {
	errorResponse(c, http.StatusUnauthorized, "unauthorized", msg, nil)
}

func Forbidden(c *gin.Context, msg string) {
	errorResponse(c, http.StatusForbidden, "forbidden", msg, nil)
}

func NotFound(c *gin.Context, msg string) {
	errorResponse(c, http.StatusNotFound, "not_found", msg, nil)
}
func Conflict(c *gin.Context, msg string) {
	errorResponse(c, http.StatusConflict, "conflict", msg, nil)
}

//	func InternalError(c *gin.Context) {
//		errorResponse(c, http.StatusInternalServerError, "internal_error", "internal server error", nil)
//	}
func InternalError(c *gin.Context, err error) {
	zap.L().Error("internal server error",
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
	)

	errorResponse(c, http.StatusInternalServerError,
		"internal_error",
		"internal server error",
		nil,
	)
}

// --- 参数校验 ---
func ValidationError(c *gin.Context, errs interface{}) {
	errorResponse(c, http.StatusBadRequest, "validation_error", "validation failed", errs)
}

// --- 小工具 ---
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
