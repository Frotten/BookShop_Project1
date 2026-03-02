package middlewares

import (
	"Project1_Shop/controllers"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controllers.HandleResponse(c, models.CodeNeedLogin)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controllers.HandleResponse(c, models.CodeInvalidToken)
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controllers.HandleResponse(c, models.CodeInvalidToken)
			c.Abort()
			return
		}
		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set("userID", mc.UserID)
		c.Next() // 后续的处理函数可以用过c.Get(ctxUserIDKey)来获取当前请求的用户信息
	}
}
