package redis

import (
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"strconv"
)

func InsertUser(p *models.User) error {
	key := "user:" + strconv.FormatInt(p.UserID, 10)
	RDB.HSet(ctx, key, "username", p.Username, "email", p.Email, "gender", p.Gender)
	RDB.Del(ctx, GetEmpty(key))
	return RDB.Expire(ctx, key, RandTTL(UserTime)).Err()
}

func SetUserAuth(userTokenHash string, UserID int64) error {
	pipe := RDB.Pipeline()
	pipe.Set(ctx, "auth:refresh:"+userTokenHash, UserID, RandTTL(jwt.TokenExpireDuration))
	pipe.Set(ctx, "login:user:"+strconv.FormatInt(UserID, 10), userTokenHash, RandTTL(jwt.TokenExpireDuration))
	_, err := pipe.Exec(ctx)
	return err
}

func SetRefreshToken(tokenHash string) error {
	_, err := RDB.Get(ctx, "auth:refresh:"+tokenHash).Result()
	return err
}

func GetTokenHash(UserID int64) string {
	Ans, err := RDB.Get(ctx, "login:user:"+strconv.FormatInt(UserID, 10)).Result()
	if err != nil {
		return ""
	}
	return Ans
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
