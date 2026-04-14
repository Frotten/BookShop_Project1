package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateOrderHandle(c *gin.Context) { //还需要将生成的订单发送到MQ中等待处理或超时过期
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
	res := logic.ReduceStock(orderParam)
	if res != models.CodeSuccess {
		zap.L().Error("ReduceStockHandle failed")
		HandleResponse(c, res)
		return
	}
	res = logic.CreateOrder(orderParam, UserID.(int64))
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
	res = logic.GetOrderItems(UserID.(int64), orderViews)
	if res != models.CodeSuccess {
		zap.L().Error("GetOrderItemsHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, orderViews)
}

func GetOrderDetailHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("GetOrderDetailHandle: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	oidStr := c.Param("order_id")
	orderID, err := strconv.ParseInt(oidStr, 10, 64)
	if err != nil || orderID <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	ov, res := logic.GetOrderDetailSecure(UserID.(int64), orderID)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, ov)
}

func OrderPayHandle(c *gin.Context) {
	var p models.OrderConfirmParam
	if err := c.ShouldBindJSON(&p); err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("OrderPayHandle: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	res := logic.OrderPay(UserID.(int64), p.OrderID)
	if res != models.CodeSuccess {
		zap.L().Error("OrderPay failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

func OrderCancelHandle(c *gin.Context) {
	var p models.OrderConfirmParam
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("OrderCancelHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.OrderCancel(p.OrderID)
	if res != models.CodeSuccess {
		zap.L().Error("OrderCancel failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}
