// import "github.com/dgrijalva/jwt-go"
// // MyClaims 自定义声明结构体并内嵌jwt.StandardClaims
// // jwt包自带的jwt.StandardClaims只包含了官方字段
// // 我们这里需要额外记录一个username字段，所以要自定义结构体
// // 如果想要保存更多信息，都可以添加到这个结构体中
//
//	type MyClaims struct {
//		UserID   int64  `json:"user_id"`
//		Username string `json:"username"`
//		jwt.StandardClaims
//	}
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5" // ← 改这里！（v5 推荐）
	"github.com/spf13/viper"
)

// Claims 自定义声明结构体并内嵌 jwt.RegisteredClaims（新版标准 Claims）
// 注意：jwt.StandardClaims 已经 deprecated，推荐用 jwt.RegisteredClaims
type MyClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`

	// jwt.RegisteredClaims 替换原来的 StandardClaims
	// 包含 iss, sub, aud, exp, nbf, iat, jti 等标准字段
	jwt.RegisteredClaims
}

// 实战不直接在里面定义，这里暂且这样写，后面得改
var mySecret = []byte("夏天夏天悄悄过去")

// 注意我只写了access tokem,没写refresh token，后面可以再补?
// GenToken 生成 JWT
func GenToken(userID int64, username string) (string, error) {
	expireHours := viper.GetInt("auth.jwt_expire")
	if expireHours <= 0 {
		expireHours = 24 // 默认 24 小时
	}
	// 创建一个我们自己的声明的数据
	c := MyClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "litcart-api",
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(mySecret)
}

// ParseToken 解析并验证 JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法（只允许 HS256）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySecret, nil
	})

	if err != nil {
		return nil, err // 包含过期、签名无效等错误
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
