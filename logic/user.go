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
		if err == nil && View != nil {
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

func GetCommentsByUser(UserID int64, UserName string) ([]*models.CommentBook, models.ResCode) {
	z, err, _ := redis.G.Do(strconv.FormatInt(UserID, 10), func() (interface{}, error) {
		ids, err := redis.GetCommentIDsByUserID(UserID)
		if err != nil {
			res, err := mysql.GetCommentsByUserID(UserID)
			if err != nil {
				return nil, err
			}
			for _, c := range res {
				_ = redis.SetCommentsToCache(c)
				c.Username = UserName
				Book, _ := GetBookByID(c.BookID)
				c.BookTitle = Book.Title
			}
			return res, err
		}
		if len(ids) == 0 {
			return nil, err
		}
		list, missIDs, err := redis.MGetComments(ids)
		if err != nil {
			return nil, err
		}
		if len(missIDs) > 0 {
			dbList, err := mysql.GetCommentsByIDs(missIDs)
			if err != nil {
				return nil, err
			}
			for _, c := range dbList {
				_ = redis.SetCommentsToCache(c)
				list = append(list, c)
			}
		}
		var res []*models.CommentBook
		for _, c := range list {
			res = append(res, c)
		}
		for _, c := range res {
			c.Username = UserName
			Book, _ := GetBookByID(c.BookID)
			c.BookTitle = Book.Title
		}
		return res, err
	})
	if err != nil {
		return nil, models.CodeServerBusy
	}
	if z == nil {
		return nil, models.CodeSuccess
	}
	return z.([]*models.CommentBook), models.CodeSuccess
}

func GetRatingsByUser(UserID int64) ([]*models.UserRating, models.ResCode) {
	z, err, _ := redis.G.Do(strconv.FormatInt(UserID, 10), func() (interface{}, error) {
		Res, err := redis.GetUserRatingsByUserID(UserID)
		if err == nil {
			return Res, nil
		}
		URB, err := mysql.GetUserRatingByUserID(UserID)
		if err != nil {
			return nil, err
		}
		var ResList []*models.UserRating
		for _, ur := range URB {
			Book, _ := GetBookByID(ur.BookID)
			ResList = append(ResList, &models.UserRating{
				UserID: ur.UserID,
				BookID: ur.BookID,
				Score:  ur.Score,
				Title:  Book.Title,
			})
		}
		_ = redis.SetUserRatings(UserID, ResList)
		return ResList, nil
	})
	if err != nil {
		return nil, models.CodeServerBusy
	}
	if z == nil {
		return nil, models.CodeSuccess
	}
	return z.([]*models.UserRating), models.CodeSuccess
}
