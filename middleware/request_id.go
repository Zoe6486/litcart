package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "requestID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取或生成 ID
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}

		// 2. 存入 Context
		c.Set(RequestIDKey, rid)

		// 3. 设置给 Response Header，方便调试
		c.Writer.Header().Set("X-Request-ID", rid)

		c.Next()
	}
}
