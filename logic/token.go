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
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
)

func hashRefreshToken(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}

func RotateRefreshSession(refreshToken string) (accessToken, newRefresh, role, username string, subjectID int64, err error) {
	if refreshToken == "" {
		return "", "", "", "", 0, ErrInvalidRefreshToken
	}
	tokenHash := hashRefreshToken(refreshToken)
	if !redis.RefreshTokenExists(tokenHash) {
		return "", "", "", "", 0, ErrInvalidRefreshToken
	}
	role, subjectID, err = redis.GetAuthByTokenHash(tokenHash)
	if err != nil || subjectID <= 0 {
		return "", "", "", "", 0, ErrInvalidRefreshToken
	}
	if redis.GetSessionTokenHash(role, subjectID) != tokenHash {
		return "", "", "", "", 0, ErrRefreshTokenRevoked
	}
	switch role {
	case redis.AuthRoleAdmin:
		admin, dbErr := mysql.GetAdminByID(subjectID)
		if dbErr != nil || admin == nil {
			return "", "", "", "", 0, ErrInvalidRefreshToken
		}
		username = admin.Username
		accessToken, err = jwt.GenAdminToken(subjectID, username)
	case redis.AuthRoleUser:
		username, err = resolveUsername(subjectID)
		if err != nil {
			return "", "", "", "", 0, err
		}
		accessToken, err = jwt.GenToken(subjectID, username)
	default:
		return "", "", "", "", 0, ErrInvalidRefreshToken
	}
	if err != nil {
		return "", "", "", "", 0, err
	}
	newRefresh, newHash, err := jwt.GenerateRefreshToken()
	if err != nil {
		return "", "", "", "", 0, err
	}
	if err = redis.RotateAuth(tokenHash, newHash, role, subjectID); err != nil {
		return "", "", "", "", 0, err
	}
	if role == redis.AuthRoleUser {
		_ = mysql.DB.Create(&models.RefreshToken{
			UserID:    subjectID,
			TokenHash: newHash,
			ExpiresAt: time.Now().Add(jwt.TokenExpireDuration),
		})
	}
	return accessToken, newRefresh, role, username, subjectID, nil
}

func resolveUsername(userID int64) (string, error) {
	if view, err := redis.GetUserInfo(userID); err == nil && view != nil {
		return view.Username, nil
	}
	user, err := mysql.GetUserInfo(userID)
	if err != nil || user == nil {
		return "", ErrInvalidRefreshToken
	}
	return user.Username, nil
}

func Refresh(refreshToken string) (accessToken, newRefresh string, err error) {
	accessToken, newRefresh, _, _, _, err = RotateRefreshSession(refreshToken)
	return accessToken, newRefresh, err
}
