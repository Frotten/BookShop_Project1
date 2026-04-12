package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateOrderHandle(c *gin.Context) {
	var orderParam models.OrderRequest
	if err := c.ShouldBindJSON(&orderParam); err != nil {
		zap.L().Error("CreateOrderHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("CreateOrderHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	res := logic.CreateOrder(orderParam, UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("CreateOrderHandle failed")
		HandleResponse(c, res)
		return
	}
	res = logic.ClearCart(UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("ClearCart failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

func GetUserOrderHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetUserOrderHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	orderViews, res := logic.GetUserOrder(UserID.(int64))
	if res != models.CodeSuccess {
		zap.L().Error("GetUserOrderHandle failed")
		HandleResponse(c, res)
		return
	}
	res = logic.GetOrderItems(orderViews)
	if res != models.CodeSuccess {
		zap.L().Error("GetOrderItemsHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, orderViews)
}
