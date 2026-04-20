package controllers

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ScoreBookHandle(c *gin.Context) {
	PageString := c.Param("page")
	PageInt, err := strconv.ParseInt(PageString, 10, 64)
	if err != nil || PageInt <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	Pages, err := logic.GetPageBooks(PageInt)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	HandleSuccess(c, Pages)
}

func GetBooksJSON(c *gin.Context) {
	PageInt := int64(1)
	Pages, err := logic.GetPageBooks(PageInt)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
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

func GetBookByTitleHandle(c *gin.Context) {
	Title := c.Param("title")
	Books, err := logic.GetBooksByTitle(Title)
	if err != nil {
		zap.L().Error("GetBooksByTitleHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeBookNotExist)
		return
	}
	HandleSuccess(c, Books)
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

func TopSaleHandle(c *gin.Context) {
	Books, res := logic.GetTopSaleList()
	if res != models.CodeSuccess {
		zap.L().Error("GetTopSaleList failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, Books)
}

func CommentHandle(c *gin.Context) {
	var CP models.CommentParam
	if err := c.ShouldBindJSON(&CP); err != nil {
		zap.L().Error("CommentHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	UserID, ok := c.Get("userID")
	if !ok || UserID == nil {
		zap.L().Error("RateBookHandle failed: UserID not found in context")
		HandleResponse(c, models.CodeServerBusy)
		return
	}
	Comment := &models.CommentBook{
		UserID:      UserID.(int64),
		BookID:      CP.BookID,
		Comment:     CP.Comment,
		ParentID:    CP.ParentID,
		RootID:      CP.RootID,
		LikeCount:   0,
		CommentTime: time.Now().Format(models.TimeParseLayout),
	}
	res := logic.CommentBook(Comment)
	if res != models.CodeSuccess {
		zap.L().Error("CommentHandle failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}

func CommentsHandle(c *gin.Context) {
	bookIDStr := c.Query("book_id")
	bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
	if err != nil || bookID <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	list, res := logic.GetCommentsByBookID(bookID)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, list)
}

func CommentLikeHandle(c *gin.Context) {
	var p models.CommentLikeParam
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("CommentLikeHandle failed", zap.Error(err))
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	res := logic.LikeComment(p.CommentID)
	if res != models.CodeSuccess {
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, nil)
}
