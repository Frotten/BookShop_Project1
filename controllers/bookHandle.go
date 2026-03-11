package controllers

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ScoreBookHandle(c *gin.Context) {
	PageString := c.Param("page")
	Page, err := strconv.ParseInt(PageString, 10, 64)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
	}
	Books, total, err := mysql.GetBooksPageByScore(Page)
	Pages := &models.Page{
		Page:  Page,
		Total: total,
		Data:  Books,
	}
	if err != nil {
		HandleResponse(c, models.CodeServerBusy)
	}
	HomePageHandleWithInfo(c, Pages)
}

func SBHbasic(c *gin.Context) {
	Page := int64(1)
	Books, total, err := mysql.GetBooksPageByScore(Page)
	Pages := &models.Page{
		Page:  Page,
		Total: total,
		Data:  Books,
	}
	if err != nil {
		HandleResponse(c, models.CodeServerBusy)
	}
	HomePageHandleWithInfo(c, Pages)
}

func AdminAddBookHandle(c *gin.Context) {
	var B models.AddBookParam
	if err := c.ShouldBindJSON(&B); err != nil {
		zap.L().Error("AdminAddBookHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.AddBook(&B)
	if res != models.CodeSuccess {
		zap.L().Error("AdminAddBookHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}
