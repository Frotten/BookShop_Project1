package middlewares

import (
	"Project1_Shop/controllers"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"crypto/sha256"
	"encoding/hex"
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
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controllers.HandleResponse(c, models.CodeInvalidToken)
			c.Abort()
			return
		}
		c.Set("userID", mc.UserID)
		c.Set("username", mc.Username)
		c.Set("permission", mc.Permission)
		c.Next()
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

func CookieAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err == nil {
			claims, parseErr := jwt.ParseToken(token)
			if parseErr == nil {
				c.Set("userID", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("permission", claims.Permission)
				c.Next()
				return
			}
		}

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.Redirect(302, "/page/LoginPage")
			c.Abort()
			return
		}

		hash := sha256.Sum256([]byte(refreshToken))
		tokenHash := hex.EncodeToString(hash[:])

		userID, lookupErr := redis.GetUserIDByTokenHash(tokenHash)
		if lookupErr != nil || userID <= 0 {
			c.Redirect(302, "/page/LoginPage")
			c.Abort()
			return
		}

		userInfo, infoErr := redis.GetUserInfo(userID)
		if infoErr != nil || userInfo == nil {
			c.Redirect(302, "/page/LoginPage")
			c.Abort()
			return
		}

		newAccessToken, tokenErr := jwt.GenToken(userID, userInfo.Username)
		if tokenErr != nil {
			c.Redirect(302, "/page/LoginPage")
			c.Abort()
			return
		}

		c.SetCookie(
			"access_token",
			newAccessToken,
			int(jwt.AccessExpireDuration.Seconds()),
			"/",
			"",
			false,
			true,
		)
		c.Set("userID", userID)
		c.Set("username", userInfo.Username)
		c.Set("permission", "user")
		c.Next()
	}
}

func CheckLoginOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			controllers.HandleResponse(c, models.CodeNeedLogin)
			c.Abort()
			return
		}
		hash := sha256.Sum256([]byte(refreshToken))
		tokenHash := hex.EncodeToString(hash[:])
		UserID, _ := c.Get("userID")
		ReTokenHash := redis.GetTokenHash(UserID.(int64))
		if tokenHash == ReTokenHash {
			c.Next()
			return
		}
		controllers.HandleResponse(c, models.CodeInvalidToken)
		c.Abort()
		return
	}
}
