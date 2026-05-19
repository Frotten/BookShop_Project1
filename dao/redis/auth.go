package redis

import (
	"Project1_Shop/pkg/jwt"
	"strconv"
	"strings"
)

const (
	AuthRoleUser  = "user"
	AuthRoleAdmin = "admin"
)

func formatAuthValue(role string, id int64) string {
	return role + ":" + strconv.FormatInt(id, 10)
}

func parseAuthValue(val string) (role string, id int64, err error) {
	if strings.Contains(val, ":") {
		parts := strings.SplitN(val, ":", 2)
		id, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", 0, err
		}
		return parts[0], id, nil
	}
	id, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return "", 0, err
	}
	return AuthRoleUser, id, nil
}

func SetUserAuth(tokenHash string, userID int64) error {
	pipe := RDB.Pipeline()
	pipe.Set(ctx, "auth:refresh:"+tokenHash, formatAuthValue(AuthRoleUser, userID), RandTTL(jwt.TokenExpireDuration))
	pipe.Set(ctx, "login:user:"+strconv.FormatInt(userID, 10), tokenHash, RandTTL(jwt.TokenExpireDuration))
	_, err := pipe.Exec(ctx)
	return err
}

func SetAdminAuth(tokenHash string, adminID int64) error {
	pipe := RDB.Pipeline()
	pipe.Set(ctx, "auth:refresh:"+tokenHash, formatAuthValue(AuthRoleAdmin, adminID), RandTTL(jwt.TokenExpireDuration))
	pipe.Set(ctx, "login:admin:"+strconv.FormatInt(adminID, 10), tokenHash, RandTTL(jwt.TokenExpireDuration))
	_, err := pipe.Exec(ctx)
	return err
}

func GetAuthByTokenHash(tokenHash string) (role string, id int64, err error) {
	val, err := RDB.Get(ctx, "auth:refresh:"+tokenHash).Result()
	if err != nil {
		return "", 0, err
	}
	return parseAuthValue(val)
}

func GetSessionTokenHash(role string, id int64) string {
	var key string
	switch role {
	case AuthRoleAdmin:
		key = "login:admin:" + strconv.FormatInt(id, 10)
	default:
		key = "login:user:" + strconv.FormatInt(id, 10)
	}
	ans, err := RDB.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	return ans
}

func RotateAuth(oldHash, newHash, role string, id int64) error {
	pipe := RDB.Pipeline()
	pipe.Del(ctx, "auth:refresh:"+oldHash)
	pipe.Set(ctx, "auth:refresh:"+newHash, formatAuthValue(role, id), RandTTL(jwt.TokenExpireDuration))
	switch role {
	case AuthRoleAdmin:
		pipe.Set(ctx, "login:admin:"+strconv.FormatInt(id, 10), newHash, RandTTL(jwt.TokenExpireDuration))
	default:
		pipe.Set(ctx, "login:user:"+strconv.FormatInt(id, 10), newHash, RandTTL(jwt.TokenExpireDuration))
	}
	_, err := pipe.Exec(ctx)
	return err
}

func RefreshTokenExists(tokenHash string) bool {
	n, err := RDB.Exists(ctx, "auth:refresh:"+tokenHash).Result()
	return err == nil && n > 0
}
