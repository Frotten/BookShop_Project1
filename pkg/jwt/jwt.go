package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TokenExpireDuration = 7 * 24 * time.Hour

var mySecret = []byte("Doctor")

type MyClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenToken(userID int64, username string) (string, error) {
	c := MyClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()), //签发时间
			NotBefore: jwt.NewNumericDate(time.Now()), //生效时间
			Issuer:    "BlueBell",                     //签发人
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c) //使用HS256加密算法
	return token.SignedString(mySecret)                   //使用密钥进行签名
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
