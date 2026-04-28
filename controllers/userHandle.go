package controllers

import (
	"Project1_Shop/dao/redis"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"Project1_Shop/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SignUpHandler 用户注册
// @Summary      用户注册
// @Description  注册新用户，需要提供用户名、密码、确认密码、邮箱和性别
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      models.ParamSignUp  true  "注册参数"
// @Success      200   {object}  models.ResponseData{data=interface{}}  "注册成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 用户已存在 / 服务器繁忙"
// @Router       /api/register [post]
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

// LoginHandler 用户登录
// @Summary      用户登录
// @Description  用户登录，成功后通过 Cookie 下发 access_token 和 refresh_token，同时在响应体中返回 access_token
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      models.ParamLogin  true  "登录参数"
// @Success      200   {object}  models.ResponseData{data=map[string]string}  "登录成功，返回 access_token"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 用户名或密码错误 / 服务器繁忙"
// @Router       /api/login [post]
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

// RefreshHandler 刷新 Token
// @Summary      刷新 Access Token
// @Description  前端在 Access Token 过期时调用该接口，通过 Cookie 中的 refresh_token 换取新的 access_token
// @Tags         用户
// @Produce      json
// @Success      200  {object}  models.ResponseData{data=map[string]string}  "刷新成功，返回新的 access_token"
// @Failure      200  {object}  models.ResponseData  "无效的 Token"
// @Router       /refreshtoken [post]
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

// GetUserInfoHandle 获取当前用户信息
// @Summary      获取当前登录用户信息
// @Description  获取当前登录用户的基本信息（需要登录）
// @Tags         用户
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=models.UserView}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "未登录 / 服务器繁忙"
// @Router       /api/userInfo [get]
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

// GetUserCommentsHandle 获取当前用户的评论列表
// @Summary      获取当前用户的评论列表
// @Description  获取当前登录用户发表的所有评论（需要登录）
// @Tags         用户
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=[]models.CommentBook}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "未登录 / 服务器繁忙"
// @Router       /api/userComments [get]
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

// GetUserRatingsHandle 获取当前用户的评分记录
// @Summary      获取当前用户的评分记录
// @Description  获取当前登录用户对书籍的所有评分记录（需要登录）
// @Tags         用户
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=[]models.UserRating}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "未登录 / 服务器繁忙"
// @Router       /api/userRatings [get]
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
