package controllers

import (
	"Project1_Shop/dao/redis"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SignUpHandler(c *gin.Context) {
	var p models.ParamSignUp
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("SignUpHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	err := logic.SignUp(&p)
	if err != models.CodeSuccess {
		zap.L().Error("models.SignUp failed")
		if err == models.CodeUserExist {
			HandleResponse(c, models.CodeUserExist)
			return
		}
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, nil)
}

func LoginHandler(c *gin.Context) {
	var p models.ParamLogin
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("LoginHandler", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	User, code := logic.Login(&p)
	if code != models.CodeSuccess {
		zap.L().Error("models.Login failed")
		HandleResponse(c, code)
		return
	}
	accessToken, err := jwt.GenToken(User.UserID, User.Username)
	if err != nil {
		zap.L().Error("jwt.GenToken failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	refreshToken, userTokenHash, err := jwt.GenerateRefreshToken()
	if err != nil {
		zap.L().Error("jwt.GenerateRefreshToken failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	err = redis.SetUserAuth(userTokenHash, User.UserID)
	if err != nil {
		zap.L().Error("redis.SetUserAuth failed", zap.Error(err))
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(jwt.TokenExpireDuration.Seconds()),
		"/",
		"",
		false,
		true,
	)
	c.SetCookie(
		"access_token",
		accessToken,
		int(jwt.AccessExpireDuration.Seconds()),
		"/",
		"",
		false,
		true,
	)
	HandleSuccess(c, gin.H{
		"access_token": accessToken,
	})
}

func RefreshHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		HandleResponse(c, models.CodeInvalidToken)
		return
	}
	newAccess, newRefresh, err := logic.Refresh(refreshToken, c)
	if err != nil {
		HandleResponse(c, models.CodeInvalidToken)
		return
	}
	c.SetCookie("refresh_token", newRefresh, int(jwt.TokenExpireDuration), "/", "", false, true)
	HandleSuccess(c, gin.H{
		"access_token": newAccess,
	})
}

func GetUserInfoHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetUserInfoHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	UserInfo, res := logic.GetUserInfo(UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("GetUserInfoHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, UserInfo)
}

func GetUserCommentsHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetUserCommentsHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	UserName, ok := c.Get("username")
	if !ok || UserName == nil {
		zap.L().Error("GetUserCommentsHandle failed: UserName not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	Comments, res := logic.GetCommentsByUser(UserID.(int64), UserName.(string))
	if res != models.CodeSuccess {
		zap.L().Error("GetUserCommentsHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, Comments)
}

func GetUserRatingsHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetUserRatingsHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	Ratings, res := logic.GetRatingsByUser(UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("GetUserRatingsHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, Ratings)
}
