package redis

import (
	"Project1_Shop/models"
	"strconv"
)

func InsertUser(p *models.User) error {
	key := "user:" + strconv.FormatInt(p.UserID, 10)
	RDB.HSet(ctx, key, "username", p.Username, "email", p.Email, "gender", p.Gender)
	RDB.Del(ctx, GetEmpty(key))
	return RDB.Expire(ctx, key, RandTTL(UserTime)).Err()
}

func GetUserInfo(UserID int64) (*models.UserView, error) {
	key := "user:" + strconv.FormatInt(UserID, 10)
	data, err := RDB.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	genderInt, err := strconv.ParseInt(data["gender"], 10, 8)
	if err != nil {
		return nil, err
	}
	userView := &models.UserView{
		UserID:   UserID,
		Username: data["username"],
		Email:    data["email"],
		Gender:   int8(genderInt),
	}
	return userView, nil
}
