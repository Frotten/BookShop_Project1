package controllers

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"database/sql"
	"errors"
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

func DeleteBookParamHandle(c *gin.Context) {
	IDString := c.Param("book_id")
	ID, err := strconv.ParseInt(IDString, 10, 64)
	if err != nil {
		zap.L().Error("DeleteBookParamHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	Book, err := logic.GetBookByID(ID)
	if err != nil {
		zap.L().Error("DeleteBookParamHandle failed", zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			HandleResponse(c, models.CodeBookNotExist)
			return
		}
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, Book)
}

func AdminDeleteBookHandle(c *gin.Context) {
	IDString := c.Param("book_id")
	ID, err := strconv.ParseInt(IDString, 10, 64)
	if err != nil {
		zap.L().Error("AdminDeleteBookHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.DeleteBook(ID)
	if res != models.CodeSuccess {
		zap.L().Error("AdminDeleteBookHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

func AdminListBookHandle(c *gin.Context) {

}
