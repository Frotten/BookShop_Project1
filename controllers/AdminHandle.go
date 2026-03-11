package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"

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
