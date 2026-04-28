package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SeckillListHandle 获取进行中的秒杀活动列表
// @Summary      获取进行中的秒杀活动列表
// @Description  获取当前所有进行中的秒杀活动（状态为上架）
// @Tags         秒杀
// @Produce      json
// @Success      200  {object}  models.ResponseData{data=[]models.SeckillProductView}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "服务器繁忙"
// @Router       /api/seckill/list [get]
func SeckillListHandle(c *gin.Context) {
	list, res := logic.GetActiveSeckillList()
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, list)
}

// SeckillDetailHandle 获取秒杀活动详情
// @Summary      获取秒杀活动详情
// @Description  根据秒杀活动 ID 获取详细信息（包含书籍信息、价格、库存、时间等）
// @Tags         秒杀
// @Produce      json
// @Param        id   path      int  true  "秒杀活动 ID"
// @Success      200  {object}  models.ResponseData{data=models.SeckillProductView}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "参数错误 / 活动不存在"
// @Router       /api/seckill/{id} [get]
func SeckillDetailHandle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	view, res := logic.GetSeckillDetail(id)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, view)
}

// SeckillHandle 参与秒杀
// @Summary      参与秒杀
// @Description  登录用户参与指定秒杀活动抢购，成功后异步生成订单（需要登录）
// @Tags         秒杀
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.SeckillRequest  true  "秒杀参数（product_id 必填）"
// @Success      200   {object}  models.ResponseData{data=map[string]string}  "抢购成功"
// @Failure      200   {object}  models.ResponseData  "未登录 / 参数错误 / 活动已结束 / 库存不足 / 重复参与"
// @Router       /api/seckill/do [post]
func SeckillHandle(c *gin.Context) {
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		HandleResponse(c, models.CodeNeedLogin)
		return
	}
	var req models.SeckillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("SeckillHandle ShouldBindJSON failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.DoSeckill(UserID.(int64), req.ProductID)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, gin.H{"msg": "抢购成功，订单生成中，请前往个人中心查看"})
}

// AdminCreateSeckillHandle 管理员创建秒杀活动
// @Summary      创建秒杀活动（管理员）
// @Description  管理员为指定书籍创建秒杀活动，设置秒杀价格、库存和时间（需要管理员权限）
// @Tags         秒杀管理（管理员）
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.SeckillParam  true  "秒杀活动参数"
// @Success      200   {object}  models.ResponseData  "创建成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 无权限 / 服务器繁忙"
// @Router       /admin/seckill/create [post]
func AdminCreateSeckillHandle(c *gin.Context) {
	var param models.SeckillParam
	if err := c.ShouldBindJSON(&param); err != nil {
		zap.L().Error("AdminCreateSeckillHandle ShouldBindJSON failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.AdminCreateSeckill(&param)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

// AdminDownSeckillHandle 管理员下架秒杀活动
// @Summary      下架秒杀活动（管理员）
// @Description  管理员将指定秒杀活动下架（状态置为 0）（需要管理员权限）
// @Tags         秒杀管理（管理员）
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "秒杀活动 ID"
// @Success      200  {object}  models.ResponseData  "下架成功"
// @Failure      200  {object}  models.ResponseData  "参数错误 / 无权限 / 服务器繁忙"
// @Router       /admin/seckill/down/{id} [post]
func AdminDownSeckillHandle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.AdminDownSeckill(id)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

// AdminSeckillListHandle 管理员获取秒杀活动列表
// @Summary      获取秒杀活动列表（管理员）
// @Description  管理员获取所有进行中的秒杀活动列表（需要管理员权限）
// @Tags         秒杀管理（管理员）
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ResponseData{data=[]models.SeckillProductView}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "无权限 / 服务器繁忙"
// @Router       /admin/seckill/list [get]
func AdminSeckillListHandle(c *gin.Context) {
	list, res := logic.GetActiveSeckillList()
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, list)
}
