package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateOrderHandle 创建订单
// @Summary      创建订单
// @Description  登录用户根据购物车商品创建订单，成功后自动清空购物车（需要登录）
// @Tags         订单
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.OrderRequest  true  "订单创建参数（包含商品列表）"
// @Success      200   {object}  models.ResponseData  "创建成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 库存不足 / 未登录 / 服务器繁忙"
// @Router       /api/orderCreate [post]
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

// GetUserOrderHandle 获取当前用户订单列表
// @Summary      获取当前用户订单列表
// @Description  获取当前登录用户的所有订单及订单详情（需要登录）
// @Tags         订单
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=[]models.OrderView}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "未登录 / 服务器繁忙"
// @Router       /api/userOrders [get]
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

// GetOrderDetailHandle 获取订单详情
// @Summary      获取订单详情
// @Description  根据订单 ID 获取订单详细信息（需要登录）
// @Tags         订单
// @Produce      json
// @Security     BearerAuth
// @Param        order_id  path      int  true  "订单 ID"
// @Success      200       {object}  models.ResponseData{data=models.OrderView}  "获取成功"
// @Failure      200       {object}  models.ResponseData  "参数错误 / 订单不存在 / 未登录"
// @Router       /api/orderDetail/{order_id} [get]
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

// OrderPayHandle 支付订单
// @Summary      支付订单
// @Description  用户支付指定订单（需要登录）
// @Tags         订单
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.OrderConfirmParam  true  "订单 ID"
// @Success      200   {object}  models.ResponseData  "支付成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 订单不存在 / 未登录 / 服务器繁忙"
// @Router       /api/orderPay [post]
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

// OrderCancelHandle 取消订单
// @Summary      取消订单
// @Description  用户取消指定未支付订单（需要登录）
// @Tags         订单
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.OrderConfirmParam  true  "订单 ID"
// @Success      200   {object}  models.ResponseData  "取消成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 订单不存在 / 未登录 / 服务器繁忙"
// @Router       /api/orderCancel [post]
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

// GetShipOrderHandle 获取待发货订单列表（管理员）
// @Summary      获取待发货订单列表（管理员）
// @Description  管理员获取所有待发货订单列表（需要管理员权限）
// @Tags         订单管理（管理员）
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=[]models.OrderView}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "无权限 / 服务器繁忙"
// @Router       /admin/order/list [get]
func GetShipOrderHandle(c *gin.Context) {
	OV, res := logic.GetShipOrder()
	if res != models.CodeSuccess {
		zap.L().Error("GetShipOrderHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, OV)
}

// OrderShipHandle 订单发货（管理员）
// @Summary      订单发货（管理员）
// @Description  管理员对指定订单执行发货操作（需要管理员权限）
// @Tags         订单管理（管理员）
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.OrderConfirmParam  true  "订单 ID"
// @Success      200   {object}  models.ResponseData  "发货成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 订单不存在 / 无权限 / 服务器繁忙"
// @Router       /admin/order/orderShip [post]
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

// OrderConfirmHandle 确认收货
// @Summary      确认收货
// @Description  用户确认收货，订单状态变更为已完成（需要登录）
// @Tags         订单
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.OrderConfirmParam  true  "订单 ID"
// @Success      200   {object}  models.ResponseData  "确认成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 订单不存在 / 重复操作 / 未登录 / 服务器繁忙"
// @Router       /api/orderConfirm [post]
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
