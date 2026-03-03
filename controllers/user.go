package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"

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
	HandleResponse(c, models.CodeSuccess)
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
	HandleSuccess(c, User)
}

func LogoutHandler(c *gin.Context) {

}
