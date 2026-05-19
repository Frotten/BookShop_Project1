package authcookie

import (
	"Project1_Shop/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func SetTokenCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie(
		"access_token",
		accessToken,
		int(jwt.AccessExpireDuration.Seconds()),
		"/",
		"",
		false,
		true,
	)
	if refreshToken != "" {
		c.SetCookie(
			"refresh_token",
			refreshToken,
			int(jwt.TokenExpireDuration.Seconds()),
			"/",
			"",
			false,
			true,
		)
	}
}
