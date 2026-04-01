package controllers

import (
	"Project1_Shop/dao/redis"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
	AdminLoginPageHandle(c)
}

func AdminLoginHandler(c *gin.Context) {
	var A models.Admin
	if err := c.ShouldBind(&A); err != nil {
		zap.L().Error("AdminLoginHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.AdminLogin(&A)
	if res != models.CodeSuccess {
		zap.L().Error("logic.AdminLogin failed")
		if res == models.CodeInvalidPassword || res == models.CodeUserNotExist {
			HandleResponse(c, models.CodeInvalidPassword)
			return
		}
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	accessToken, err := jwt.GenAdminToken(A.AdminID, A.Username)
	if err != nil {
		zap.L().Error("jwt.GenAdminToken failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}

	refreshToken, tokenHash, err := jwt.GenerateRefreshToken()
	if err != nil {
		zap.L().Error("jwt.GenerateAdminRefreshToken failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	redis.RDB.Set(c, "auth:refresh:"+tokenHash, A.AdminID, jwt.TokenExpireDuration)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(jwt.TokenExpireDuration.Seconds()),
		"/",
		"",
		true,
		true,
	)
	c.SetCookie(
		"access_token",
		accessToken,
		int(jwt.AccessExpireDuration.Seconds()),
		"/",
		"",
		true,
		true,
	)
	HandleSuccess(c, gin.H{
		"access_token": accessToken,
	})
}
