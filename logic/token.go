package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

func Refresh(refreshToken string) (string, string, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	var stored models.RefreshToken
	err := mysql.DB.Where("token_hash = ?", tokenHash).First(&stored).Error
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if stored.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token expired")
	}

	// 删除旧 token（防止重放攻击）
	mysql.DB.Delete(&stored)

	// 生成新 token
	accessToken, _ := jwt.GenToken(stored.UserID)
	newRefresh, newHash, _ := jwt.GenerateRefreshToken()

	mysql.DB.Create(&models.RefreshToken{
		UserID:    stored.UserID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	return accessToken, newRefresh, nil
}
