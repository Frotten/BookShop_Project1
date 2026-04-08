package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
