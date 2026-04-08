package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

func Refresh(refreshToken string, c *gin.Context) (string, string, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])
	ans, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if ans.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token expired")
	}
	err = redis.SetRefreshToken(tokenHash)
	if err != nil {
		return "", "", err
	}
	redis.RDB.Del(c, "auth:refresh:"+tokenHash)
	accessToken, _ := jwt.GenToken(ans.UserID, ans.Username)
	newRefresh, newHash, _ := jwt.GenerateRefreshToken()
	mysql.DB.Create(&models.RefreshToken{
		UserID:    ans.UserID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	return accessToken, newRefresh, nil
}
