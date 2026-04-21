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
	oidStr := c.Param("order_id")
	orderID, err := strconv.ParseInt(oidStr, 10, 64)
	if err != nil || orderID <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	ov, res := logic.GetOrderDetailSecure(orderID)
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
	res := logic.OrderPay(p.OrderID)
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

func GetShipOrderHandle(c *gin.Context) {
	OV, res := logic.GetShipOrder()
	if res != models.CodeSuccess {
		zap.L().Error("GetShipOrderHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, OV)
}

func OrderShipHandle(c *gin.Context) {
	var p models.OrderConfirmParam
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("OrderShipHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.OrderShip(p.OrderID)
	if res != models.CodeSuccess {
		zap.L().Error("OrderShip failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

func OrderConfirmHandle(c *gin.Context) {
	var p models.OrderConfirmParam
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("OrderConfirmHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.OrderConfirm(p.OrderID)
	if res != models.CodeSuccess {
		zap.L().Error("OrderConfirm failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

//func Seckill(c *gin.Context) {
//	UserID, ok := c.Get("userID")
//	if !ok || UserID == nil {
//		zap.L().Error("GetUserOrderHandle failed: UserID not found in context")
//		HandleResponse(c, models.CodeServerBusy)
//		return
//	}
//	productIDStr := c.Param("id")
//	productID, err := strconv.ParseInt(productIDStr, 10, 64)
//	if err != nil || productID <= 0 {
//		HandleResponse(c, models.CodeInvalidParam)
//		return
//	}
//	res := logic.Seckill(UserID.(int64), productID)
//	if res != models.CodeSuccess {
//		zap.L().Error("Seckill failed")
//		HandleResponse(c, res)
//		return
//	}
//	HandleSuccess(c, nil)
//}
