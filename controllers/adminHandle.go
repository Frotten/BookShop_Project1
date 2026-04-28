package controllers

import (
	"Project1_Shop/dao/redis"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminRegisterHandler 管理员注册
// @Summary      管理员注册
// @Description  注册新管理员账号，注册成功后自动跳转至管理员登录页面
// @Tags         管理员认证
// @Accept       json
// @Produce      json
// @Param        body  body      models.Admin  true  "管理员注册参数（username、password）"
// @Success      200   {object}  models.ResponseData  "注册成功，跳转到登录页"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 用户已存在 / 服务器繁忙"
// @Router       /api/AdminRegister [post]
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

// AdminLoginHandler 管理员登录
// @Summary      管理员登录
// @Description  管理员登录，成功后通过 Cookie 下发 access_token 和 refresh_token，同时在响应体中返回 access_token
// @Tags         管理员认证
// @Accept       json
// @Produce      json
// @Param        body  body      models.Admin  true  "管理员登录参数（username、password）"
// @Success      200   {object}  models.ResponseData{data=map[string]string}  "登录成功，返回 access_token"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 用户名或密码错误 / 服务器繁忙"
// @Router       /api/AdminLogin [post]
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
		false,
		true,
	)
	c.SetCookie(
		"access_token",
		accessToken,
		int(jwt.AccessExpireDuration.Seconds()),
		"/",
		"",
		false,
		true,
	)
	HandleSuccess(c, gin.H{
		"access_token": accessToken,
	})
}
