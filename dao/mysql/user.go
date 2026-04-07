package mysql

import (
	"Project1_Shop/models"
	"Project1_Shop/pkg/md5"
)

const secret = "FrostNova"

func CheckUserExist(account string) (bool, models.ResCode) {
	var u models.User
	result := DB.Where("username = ? OR email = ?", account, account).First(&u) //这里改用account来查询，既可以通过用户名查询，也可以通过邮箱查询，因为在注册时，用户名和邮箱都必须唯一。
	if result.RowsAffected == 0 {
		return true, models.CodeSuccess
	}
	if result.RowsAffected != 0 {
		return false, models.CodeUserExist
	}
	return false, models.CodeServerBusy
}

func CheckUserLogin(p *models.ParamLogin) (*models.User, models.ResCode) {
	var u models.User
	result := DB.Where("username = ?", p.Username).First(&u)
	if result.RowsAffected == 0 {
		return nil, models.CodeUserNotExist
	}
	if u.Password != md5.Md5(p.Password) {
		return nil, models.CodeInvalidPassword
	}
	return &u, models.CodeSuccess
}

func InsertUser(p *models.ParamSignUp) (models.User, models.ResCode) {
	var u models.User
	u.Username = p.Username
	u.Password = md5.Md5(p.Password)
	u.Email = p.Email
	u.Gender = p.Gender
	result := DB.Create(&u)
	if result.Error != nil {
		return models.User{}, models.CodeServerBusy
	}
	return u, models.CodeSuccess
}

func CheckAdminExist(account string) (bool, models.ResCode) {
	var u models.User
	result := DB.Where("username = ?", account).First(&u)
	if result.RowsAffected == 0 {
		return true, models.CodeSuccess
	}
	if result.RowsAffected != 0 {
		return false, models.CodeUserExist
	}
	return false, models.CodeServerBusy
}

func InsertAdmin(p *models.Admin) models.ResCode {
	var u models.Admin
	u.Username = p.Username
	u.Password = md5.Md5(p.Password)
	result := DB.Create(&u)
	if result.Error != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func AdminLogin(p *models.Admin) models.ResCode {
	var u models.Admin
	result := DB.Where("username = ? AND password = ?", p.Username, md5.Md5(p.Password)).First(&u)
	if result.RowsAffected == 0 {
		return models.CodeUserNotExist
	}
	return models.CodeSuccess
}

func GetUserInfo(UserID int64) (*models.User, error) {
	var u models.User
	result := DB.First(&u, UserID)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}
	return &u, result.Error
}
