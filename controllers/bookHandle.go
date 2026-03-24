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
	if err != nil || Page <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	Books, total, err := mysql.GetBooksPageByScore(Page)
	Pages := &models.Page{
		Page:  Page,
		Total: total,
		Data:  Books,
	}
	if err != nil {
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, Pages)
}

func GetBooksJSON(c *gin.Context) {
	Page := int64(1)
	Books, total, err := mysql.GetBooksPageByScore(Page)
	if err != nil {
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	Pages := &models.Page{
		Page:  Page,
		Total: total,
		Data:  Books,
	}
	HandleSuccess(c, Pages)
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

func GetBookParamHandle(c *gin.Context) {
	IDString := c.Param("book_id")
	ID, err := strconv.ParseInt(IDString, 10, 64)
	if err != nil {
		zap.L().Error("GetBookParamHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	Book, err := logic.GetBookByID(ID)
	if err != nil {
		zap.L().Error("GetBookParamHandle failed", zap.Error(err))
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

func AdminUpdateBookHandle(c *gin.Context) {
	var B models.UpdateBookParam
	if err := c.ShouldBindJSON(&B); err != nil {
		zap.L().Error("AdminUpdateBookHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	exist := mysql.ExistBook(B.BookID)
	if !exist {
		HandleResponse(c, models.CodeBookNotExist)
		return
	}
	res := logic.UpdateBook(&B)
	if res != models.CodeSuccess {
		zap.L().Error("AdminUpdateBookHandle failed")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	HandleSuccess(c, nil)
}

func RateBookHandle(c *gin.Context) {
	var R models.UserRateBook
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("RateBookHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	if err := c.ShouldBindJSON(&R); err != nil {
		zap.L().Error("RateBookHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	R.UserID = UserID.(int64)
	res := logic.RateBook(&R)
	if res != models.CodeSuccess {
		zap.L().Error("RateBookHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

func TopScoreHandle(c *gin.Context) {
	ListBook, res := logic.GetTopScoreList()
	if res != models.CodeSuccess {
		zap.L().Error("GetTopScoreList failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, ListBook)
}
