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

// ScoreBookHandle 按评分分页获取书籍列表（指定页码）
// @Summary      按评分分页获取书籍（指定页码）
// @Description  按书籍评分降序分页返回书籍列表，指定页码（每页 8 条）
// @Tags         书籍
// @Produce      json
// @Param        page  path      int  true  "页码（从 1 开始）"
// @Success      200   {object}  models.ResponseData{data=models.Page}  "获取成功"
// @Failure      200   {object}  models.ResponseData  "参数错误"
// @Router       /api/getBooksJSON/{page} [get]
func ScoreBookHandle(c *gin.Context) {
	PageString := c.Param("page")
	PageInt, err := strconv.ParseInt(PageString, 10, 64)
	if err != nil || PageInt <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	Pages, err := logic.GetPageBooks("score", PageInt)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	HandleSuccess(c, Pages)
}

// GetBooksJSON 按评分获取第一页书籍列表
// @Summary      获取书籍列表（第1页，按评分排序）
// @Description  按书籍评分降序返回第 1 页书籍列表
// @Tags         书籍
// @Produce      json
// @Success      200  {object}  models.ResponseData{data=models.Page}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "参数错误"
// @Router       /api/getBooksJSON [get]
func GetBooksJSON(c *gin.Context) {
	PageInt := int64(1)
	Pages, err := logic.GetPageBooks("score", PageInt)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	HandleSuccess(c, Pages)
}

// GetBooksBySaleJSON 按销量获取第一页书籍列表
// @Summary      获取书籍列表（第1页，按销量排序）
// @Description  按书籍销量降序返回第 1 页书籍列表
// @Tags         书籍
// @Produce      json
// @Success      200  {object}  models.ResponseData{data=models.Page}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "参数错误"
// @Router       /api/getBooksBySaleJSON [get]
func GetBooksBySaleJSON(c *gin.Context) {
	PageInt := int64(1)
	Pages, err := logic.GetPageBooksBySale("sale", PageInt)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	HandleSuccess(c, Pages)
}

// SaleBookHandle 按销量分页获取书籍列表（指定页码）
// @Summary      按销量分页获取书籍（指定页码）
// @Description  按书籍销量降序分页返回书籍列表，指定页码（每页 8 条）
// @Tags         书籍
// @Produce      json
// @Param        page  path      int  true  "页码（从 1 开始）"
// @Success      200   {object}  models.ResponseData{data=models.Page}  "获取成功"
// @Failure      200   {object}  models.ResponseData  "参数错误"
// @Router       /api/getBooksBySaleJSON/{page} [get]
func SaleBookHandle(c *gin.Context) {
	PageString := c.Param("page")
	PageInt, err := strconv.ParseInt(PageString, 10, 64)
	if err != nil || PageInt <= 0 {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	Pages, err := logic.GetPageBooksBySale("sale", PageInt)
	if err != nil {
		HandleResponse(c, models.CodeInvalidParam)
		return
	}
	HandleSuccess(c, Pages)
}

// AdminAddBookHandle 管理员添加书籍
// @Summary      添加书籍（管理员）
// @Description  管理员添加新书籍（需要管理员权限）
// @Tags         书籍管理（管理员）
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.AddBookParam  true  "书籍信息"
// @Success      200   {object}  models.ResponseData  "添加成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 无权限 / 服务器繁忙"
// @Router       /admin/book/add [post]
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

// GetBookParamHandle 根据 ID 获取书籍详情
// @Summary      根据书籍 ID 获取详情
// @Description  通过书籍 ID 获取书籍详细信息（管理员和普通用户均可访问）
// @Tags         书籍
// @Produce      json
// @Param        book_id  path      int  true  "书籍 ID"
// @Success      200      {object}  models.ResponseData{data=models.Book}  "获取成功"
// @Failure      200      {object}  models.ResponseData  "参数错误 / 书籍不存在 / 服务器繁忙"
// @Router       /api/getBookDetail/{book_id} [get]
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

// GetBookByTitleHandle 根据书名模糊搜索书籍
// @Summary      根据书名搜索书籍
// @Description  通过书名关键字模糊搜索书籍列表
// @Tags         书籍
// @Produce      json
// @Param        title  path      string  true  "书名关键字"
// @Success      200    {object}  models.ResponseData{data=[]models.Book}  "获取成功"
// @Failure      200    {object}  models.ResponseData  "书籍不存在"
// @Router       /api/getBookTitle/{title} [get]
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

// AdminDeleteBookHandle 管理员删除书籍
// @Summary      删除书籍（管理员）
// @Description  管理员根据书籍 ID 删除书籍（需要管理员权限）
// @Tags         书籍管理（管理员）
// @Produce      json
// @Security     BearerAuth
// @Param        book_id  path      int  true  "书籍 ID"
// @Success      200      {object}  models.ResponseData  "删除成功"
// @Failure      200      {object}  models.ResponseData  "参数错误 / 书籍不存在 / 无权限 / 服务器繁忙"
// @Router       /admin/book/delete/{book_id} [delete]
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

// AdminUpdateBookHandle 管理员更新书籍信息
// @Summary      更新书籍信息（管理员）
// @Description  管理员更新书籍信息，book_id 为必填（需要管理员权限）
// @Tags         书籍管理（管理员）
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.UpdateBookParam  true  "更新书籍参数"
// @Success      200   {object}  models.ResponseData  "更新成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 书籍不存在 / 无权限 / 服务器繁忙"
// @Router       /admin/book/update [post]
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

// RateBookHandle 用户对书籍评分
// @Summary      对书籍评分
// @Description  登录用户对书籍进行评分（需要登录）
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.UserRateBook  true  "评分参数"
// @Success      200   {object}  models.ResponseData  "评分成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 未登录 / 服务器繁忙"
// @Router       /api/rateBook [post]
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

// TopScoreHandle 获取评分最高的书籍列表
// @Summary      获取评分 Top 书籍
// @Description  获取综合评分最高的 Top 书籍列表
// @Tags         书籍
// @Produce      json
// @Success      200  {object}  models.ResponseData{data=[]models.Book}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "服务器繁忙"
// @Router       /api/topScore [get]
func TopScoreHandle(c *gin.Context) {
	ListBook, res := logic.GetTopScoreList()
	if res != models.CodeSuccess {
		zap.L().Error("GetTopScoreList failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, ListBook)
}

// TopSaleHandle 获取销量最高的书籍列表
// @Summary      获取销量 Top 书籍
// @Description  获取销量最高的 Top 书籍列表
// @Tags         书籍
// @Produce      json
// @Success      200  {object}  models.ResponseData{data=[]models.Book}  "获取成功"
// @Failure      200  {object}  models.ResponseData  "服务器繁忙"
// @Router       /api/topSale [get]
func TopSaleHandle(c *gin.Context) {
	Books, res := logic.GetTopSaleList()
	if res != models.CodeSuccess {
		zap.L().Error("GetTopSaleList failed")
		HandleResponse(c, res)
		return
	}
	HandleSuccess(c, Books)
}

// CommentHandle 发表评论
// @Summary      发表书籍评论
// @Description  登录用户对书籍发表评论，支持嵌套回复（需要登录）
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CommentParam  true  "评论参数"
// @Success      200   {object}  models.ResponseData  "评论成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 未登录 / 服务器繁忙"
// @Router       /api/comment [post]
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

// CommentsHandle 获取书籍评论列表
// @Summary      获取书籍评论列表
// @Description  根据书籍 ID 获取该书籍的所有评论（树形结构）
// @Tags         书籍
// @Produce      json
// @Param        book_id  query     int  true  "书籍 ID"
// @Success      200      {object}  models.ResponseData{data=[]models.CommentBook}  "获取成功"
// @Failure      200      {object}  models.ResponseData  "参数错误 / 服务器繁忙"
// @Router       /api/comments [get]
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

// CommentLikeHandle 点赞评论
// @Summary      点赞评论
// @Description  登录用户对指定评论进行点赞（需要登录）
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CommentLikeParam  true  "评论点赞参数"
// @Success      200   {object}  models.ResponseData  "点赞成功"
// @Failure      200   {object}  models.ResponseData  "参数错误 / 服务器繁忙"
// @Router       /api/comment/like [post]
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
