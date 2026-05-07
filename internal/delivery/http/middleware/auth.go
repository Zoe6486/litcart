package middleware

// middleware 只懂 HTTP

// 不懂你的业务 code

// 不依赖 controller

import (
	"litcart/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// // 定义身份结构（middleware 层）
//
//	type AuthUser struct {
//		UserID   int64
//		Username string
//	}
//
// // 私有 key（不导出）
// type authUserKeyType struct{}
// var authUserKey = authUserKeyType{}
// 上面加了username其实不需要
// 只存userID
type userIDKeyType struct{}

var userIDKey = userIDKeyType{}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// Authorization: Bearer xxxxxxx.xxx.xxx  / X-TOKEN: xxx.xxx.xx
		// 这里的具体实现方式要依据你的实际业务情况决定
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
			})
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization header format",
			})
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// ⭐ 关键：把 identity 放进 context
		// c.Set(authUserKey, AuthUser{
		// 	UserID:   claims.UserID,
		// 	Username: claims.Username,
		// })
		c.Set(userIDKey, claims.UserID)

		c.Next()
	}
}

// 提供 accessor（唯一入口）
//
//	func CurrentUser(c *gin.Context) (AuthUser, bool) {
//		u, ok := c.Get(authUserKey)
//		if !ok {
//			return AuthUser{}, false
//		}
//		user, ok := u.(AuthUser)
//		return user, ok
//	}
func CurrentUserID(c *gin.Context) (int64, bool) {
	id, ok := c.Get(userIDKey)
	if !ok {
		return 0, false
	}
	uid, ok := id.(int64)
	return uid, ok
}
