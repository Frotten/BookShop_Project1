package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenExpireDuration = 7 * 24 * time.Hour
const AccessExpireDuration = 15 * time.Minute

var mySecret = []byte("Doctor")

type MyClaims struct {
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	Permission string `json:"role"`
	jwt.RegisteredClaims
}

func GenToken(userID int64, username string) (Token string, err error) {
	c := MyClaims{
		UserID:     userID,
		Username:   username,
		Permission: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Shop",
		},
	}
	Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)
	return
}

func GenAdminToken(userID int64, username string) (Token string, err error) {
	c := MyClaims{
		UserID:     userID,
		Username:   username,
		Permission: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Shop",
		},
	}
	Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)
	return
}

func GenerateRefreshToken() (string, string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", "", err
	}
	token := base64.URLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	return token, tokenHash, nil
}

func ParseToken(tokenString string) (*MyClaims, error) {
	var mc = new(MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
