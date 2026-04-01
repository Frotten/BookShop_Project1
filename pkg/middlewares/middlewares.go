package middlewares

import (
	"Project1_Shop/controllers"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
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
		c.Set("username", mc.Username)
		c.Set("permission", mc.Permission)
		c.Next() // 后续的处理函数可以用过c.Get(ctxUserIDKey)来获取当前请求的用户信息
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("permission")
		if !exists || role != "admin" {
			c.JSON(http.StatusFound, gin.H{
				"error": "admin only",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CookieAuthMiddleware() gin.HandlerFunc { //页面跳转无法将token写入消息头，故改用cookie
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(302, "/page/LoginPage")
			c.Abort()
			return
		}
		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.Redirect(302, "/page/LoginPage")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("permission", claims.Permission)
		c.Next()
	}
}
