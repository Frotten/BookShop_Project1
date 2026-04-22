package controllers

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SeckillListHandle(c *gin.Context) {
	list, res := logic.GetActiveSeckillList()
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, list)
}

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

func AdminSeckillListHandle(c *gin.Context) {
	list, res := logic.GetActiveSeckillList()
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, list)
}
