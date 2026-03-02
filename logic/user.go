package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/models"

	"go.uber.org/zap"
)

func SignUp(p *models.ParamSignUp) models.ResCode {
	ok, err := mysql.CheckUserExist(p.Username) //查询用户名合法性（不能是已存在昵称，也不能是别人的邮箱）
	if ok == false {
		if err != models.CodeUserExist {
			zap.L().Error("logic.SignUp failed")
			return models.CodeServerBusy
		}
		zap.L().Error("logic.SignUp failed")
		return models.CodeUserExist
	}
	ok, err = mysql.CheckUserExist(p.Email) //查询邮箱合法性（不能是已存在昵称，也不能是别人的邮箱）
	if ok == false {
		if err != models.CodeUserExist {
			zap.L().Error("logic.SignUp failed")
			return models.CodeServerBusy
		}
		zap.L().Error("logic.SignUp failed")
		return models.CodeUserExist
	}
	err = mysql.InsertUser(p)
	if err != models.CodeSuccess {
		zap.L().Error("mysql.InsertUser(p) failed")
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func Login(p *models.ParamLogin) (*models.User, models.ResCode) {
	User, code := mysql.CheckUserLogin(p)
	if code != models.CodeSuccess {
		zap.L().Error("logic.Login failed")
		return nil, code
	}
	return User, models.CodeSuccess
}
