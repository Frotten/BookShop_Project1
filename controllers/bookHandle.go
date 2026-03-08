package controllers

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/models"
	"strconv"

	"github.com/gin-gonic/gin"
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
