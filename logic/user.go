package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"strconv"
)

func SignUp(p *models.ParamSignUp) models.ResCode {
	ok, res := mysql.CheckUserExist(p.Username) //查询用户名合法性（不能是已存在昵称，也不能是别人的邮箱）
	if ok == false {
		if res != models.CodeUserExist {
			return models.CodeServerBusy
		}
		return models.CodeUserExist
	}
	ok, res = mysql.CheckUserExist(p.Email) //查询邮箱合法性（不能是已存在昵称，也不能是别人的邮箱）
	if ok == false {
		if res != models.CodeUserExist {
			return models.CodeServerBusy
		}
		return models.CodeUserExist
	}
	User, res := mysql.InsertUser(p)
	if res != models.CodeSuccess || User.UserID <= 0 {
		return models.CodeServerBusy
	}
	err := redis.InsertUser(&User)
	if err != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func Login(p *models.ParamLogin) (*models.User, models.ResCode) {
	User, code := mysql.CheckUserLogin(p)
	if code != models.CodeSuccess {
		return nil, code
	}
	return User, models.CodeSuccess
}

func AdminRegister(p *models.Admin) models.ResCode {
	ok, err := mysql.CheckAdminExist(p.Username) //查询用户名合法性（不能是已存在昵称，也不能是别人的邮箱）
	if ok == false {
		if err != models.CodeUserExist {
			return models.CodeServerBusy
		}
		return models.CodeUserExist
	}
	err = mysql.InsertAdmin(p)
	if err != models.CodeSuccess {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func AdminLogin(p *models.Admin) models.ResCode {
	err := mysql.AdminLogin(p)
	if err != models.CodeSuccess {
		return err
	}
	return models.CodeSuccess
}

func GetUserInfo(UserID int64) (*models.UserView, models.ResCode) {
	UserIDStr := strconv.FormatInt(UserID, 10)
	z, err, _ := redis.G.Do(UserIDStr, func() (interface{}, error) {
		View, err := redis.GetUserInfo(UserID)
		if err == nil {
			return View, err
		}
		User, err := mysql.GetUserInfo(UserID)
		if err != nil {
			return nil, err
		}
		err = redis.InsertUser(User)
		if err != nil {
			return nil, err
		}
		return &models.UserView{
			UserID:   User.UserID,
			Username: User.Username,
			Email:    User.Email,
			Gender:   User.Gender,
		}, nil
	})
	if err != nil {
		return nil, models.CodeServerBusy
	}
	return z.(*models.UserView), models.CodeSuccess
}
