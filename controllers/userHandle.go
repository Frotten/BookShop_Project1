package controllers

import (
	"Project1_Shop/dao/redis"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SignUpHandler(c *gin.Context) {
	var p models.ParamSignUp
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("SignUpHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	err := logic.SignUp(&p)
	if err != models.CodeSuccess {
		zap.L().Error("models.SignUp failed")
		if err == models.CodeUserExist {
			HandleResponse(c, models.CodeUserExist)
			return
		}
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, nil)
}

func LoginHandler(c *gin.Context) {
	var p models.ParamLogin
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("LoginHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	User, code := logic.Login(&p)
	if code != models.CodeSuccess {
		zap.L().Error("models.Login failed")
		HandleResponse(c, code)
		return
	}
	accessToken, err := jwt.GenToken(User.UserID)
	if err != nil {
		zap.L().Error("jwt.GenToken failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	refreshToken, tokenHash, err := jwt.GenerateRefreshToken()
	if err != nil {
		zap.L().Error("jwt.GenerateRefreshToken failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}

	//mysql.DB.Create(&models.RefreshToken{
	//	UserID:    User.UserID,
	//	TokenHash: tokenHash,
	//	ExpiresAt: time.Now().Add(jwt.TokenExpireDuration),
	//})

	//redis.RDB.Set("login:token:{"+refreshToken+"}", User.UserID, jwt.TokenExpireDuration)

	redis.RDB.Set(c, "auth:refresh:"+tokenHash, User.UserID, jwt.TokenExpireDuration)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(jwt.TokenExpireDuration.Seconds()),
		"/",
		"",
		true,
		true,
	)
	HandleSuccess(c, gin.H{
		"access_token": accessToken,
	})
}

func RefreshHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		HandleResponse(c, models.CodeInvalidToken)
		return
	}
	newAccess, newRefresh, err := logic.Refresh(refreshToken, c)
	if err != nil {
		HandleResponse(c, models.CodeInvalidToken)
		return
	}
	c.SetCookie("refresh_token", newRefresh, int(jwt.TokenExpireDuration), "/", "", true, true)
	HandleSuccess(c, gin.H{
		"access_token": newAccess,
	})
}

func AdminRegisterHandler(c *gin.Context) {
	var A models.Admin
	if err := c.ShouldBind(&A); err != nil {
		zap.L().Error("AdminRegisterHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	err := logic.AdminRegister(&A)
	if err != models.CodeSuccess {
		zap.L().Error("logic.AdminRegister failed")
		if err == models.CodeUserExist {
			HandleResponse(c, models.CodeUserExist)
			return
		}
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, nil)
}

func AdminLoginHandler(c *gin.Context) {
	var A models.Admin
	if err := c.ShouldBind(&A); err != nil {
		zap.L().Error("AdminLoginHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	err := logic.AdminLogin(&A)
	if err != models.CodeSuccess {
		zap.L().Error("logic.AdminLogin failed")
		if err == models.CodeInvalidPassword || err == models.CodeUserNotExist {
			HandleResponse(c, models.CodeInvalidPassword)
			return
		}
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, nil)
}
