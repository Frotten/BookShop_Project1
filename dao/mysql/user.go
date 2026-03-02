package mysql

import (
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"Project1_Shop/pkg/md5"
	"errors"

	"gorm.io/gorm"
)

const secret = "FrostNova"

func CheckUserExist(account string) (bool, models.ResCode) {
	var u models.User
	result := DB.Where("username = ? OR email = ?", account, account).First(&u) //这里改用account来查询，既可以通过用户名查询，也可以通过邮箱查询，因为在注册时，用户名和邮箱都必须唯一。
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return true, models.CodeSuccess
	} else if result.Error != nil {
		return false, models.CodeUserExist
	}
	return false, models.CodeServerBusy
}

func CheckUserLogin(p *models.ParamLogin) (*models.User, models.ResCode) {
	var u models.User
	result := DB.Where("username = ? AND password = ?", p.Username, md5.Md5(p.Password)).Find(&u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result = DB.Where("username = ?", p.Username).Find(&u)
	}
	token, err := jwt.GenToken(u.UserID, u.Username)
	if err != nil {
		return nil, models.CodeServerBusy
	}
	u.Token = token
	return &u, models.CodeSuccess
}

func InsertUser(p *models.ParamSignUp) models.ResCode {
	var u models.User
	u.Username = p.Username
	u.Password = md5.Md5(p.Password)
	u.Email = p.Email
	u.Gender = p.Gender
	result := DB.Create(&u)
	if result.Error != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}
