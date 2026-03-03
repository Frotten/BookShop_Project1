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

var mySecret = []byte("Doctor")

type MyClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func GenToken(userID int64) (Token string, err error) {
	c := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)), //过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       //签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                       //生效时间
			Issuer:    "Shop",                                               //签发人
		},
	}
	Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret) //使用HS256加密算法,并使用密钥进行签名
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
