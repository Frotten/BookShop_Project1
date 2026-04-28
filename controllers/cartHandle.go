package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AddBookToCartHandle 添加书籍到购物车
// @Summary      添加书籍到购物车
// @Description  登录用户将书籍加入购物车，若未指定数量则默认为 1（需要登录）
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CartParam  true  "购物车参数（book_id 必填，quantity 可选）"
// @Success      200   {object}  models.ResponseData  "添加成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 未登录 / 服务器繁忙"
// @Router       /api/cart [post]
func AddBookToCartHandle(c *gin.Context) {
	var CartParam models.CartParam
	if err := c.ShouldBindJSON(&CartParam); err != nil {
		zap.L().Error("AddBookToCartHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	if CartParam.Quantity <= 0 {
		CartParam.Quantity = 1
	}
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("AddBookToCartHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	CartParam.UserID = UserID.(int64)
	res := logic.AddBookToCart(&CartParam)
	if res != models.CodeSuccess {
		zap.L().Error("AddBookToCartHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

// GetCartListHandle 获取购物车列表
// @Summary      获取购物车列表
// @Description  获取当前登录用户的购物车中所有商品（需要登录）
// @Tags         购物车
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=[]models.Cart}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "未登录 / 服务器繁忙"
// @Router       /api/cart [get]
func GetCartListHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetCartListHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	cartList, res := logic.GetCartList(UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("GetCartListHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, cartList)
}

// UpdateCartItemHandle 更新购物车商品数量
// @Summary      更新购物车中书籍数量
// @Description  更新当前登录用户购物车中指定书籍的数量（需要登录）
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CartParam  true  "更新参数（book_id 和 quantity 必填）"
// @Success      200   {object}  models.ResponseData  "更新成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 未登录 / 服务器繁忙"
// @Router       /api/cart [put]
func UpdateCartItemHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetCartListHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	var CartParam models.CartParam
	if err := c.ShouldBindJSON(&CartParam); err != nil {
		zap.L().Error("UpdateCartItemHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	CartParam.UserID = UserID.(int64)
	if CartParam.Quantity <= 0 {
		CartParam.Quantity = 1
	}
	res := logic.UpdateCartItem(&CartParam)
	if res != models.CodeSuccess {
		zap.L().Error("UpdateCartItemHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

// DeleteCartItemHandle 删除购物车中的某件商品
// @Summary      删除购物车中指定书籍
// @Description  从当前登录用户的购物车中移除指定书籍（需要登录）
// @Tags         购物车
// @Produce      json
// @Security     BearerAuth
// @Param        book_id  path      int  true  "要删除的书籍 ID"
// @Success      200      {object}  models.ResponseData  "删除成功"
// @Failure      200      {object}  models.ResponseData  "参数错误 / 未登录 / 服务器繁忙"
// @Router       /api/cart/{book_id} [delete]
func DeleteCartItemHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("DeleteCartItemHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	BookIDStr := c.Param("book_id")
	if BookIDStr == "" {
		zap.L().Error("DeleteCartItemHandle failed: book_id is empty")
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	BookID, err := strconv.ParseInt(BookIDStr, 10, 64)
	if err != nil {
		zap.L().Error("DeleteCartItemHandle failed: invalid book_id", zap.String("book_id", BookIDStr), zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.DeleteCartItem(UserID.(int64), BookID)
	if res != models.CodeSuccess {
		zap.L().Error("DeleteCartItemHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

// ClearCartHandle 清空购物车
// @Summary      清空购物车
// @Description  清空当前登录用户购物车中的所有商品（需要登录）
// @Tags         购物车
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData  "清空成功"
// @Failure      200  {object}  models.ResponseData  "未登录 / 服务器繁忙"
// @Router       /api/cart [delete]
func ClearCartHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("ClearCartHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	res := logic.ClearCart(UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("ClearCartHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}
