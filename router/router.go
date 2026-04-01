package router

import (
	"Project1_Shop/controllers"
	"Project1_Shop/logger"
	"Project1_Shop/pkg/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetUp() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.Static("/static", "./views/static")
	r.LoadHTMLGlob("./views/pages/*.html")
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Shop")
	})
	r.POST("/refreshtoken", controllers.RefreshHandler) //前端在中间件校验Token未通过时触发该函数以刷新token
	v1 := r.Group("/api")
	{
		v1.POST("/register", controllers.SignUpHandler)
		v1.POST("/login", controllers.LoginHandler)
		v1.POST("/AdminRegister", controllers.AdminRegisterHandler)
		v1.POST("/AdminLogin", controllers.AdminLoginHandler)
		v1.GET("/getBooksJSON", controllers.GetBooksJSON)
		v1.GET("/getBooksJSON/:page", controllers.ScoreBookHandle)
		v1.GET("/getBookDetail/:book_id", controllers.GetBookParamHandle)
		//v1.GET("/topSale", controllers.TopSaleHandle)
		v1.GET("/topScore", controllers.TopScoreHandle)
		v1.GET("/comments", controllers.CommentsHandle)
		Login := v1.Use(middlewares.JWTAuthMiddleware())
		{
			Login.GET("/userInfo", controllers.GetUserInfoHandle)
			Login.POST("/rateBook", controllers.RateBookHandle)
			Login.POST("/comment", controllers.CommentHandle)
			Login.POST("/comment/like", controllers.CommentLikeHandle)
			Login.POST("/cart", controllers.AddBookToCartHandle)
			Login.GET("/cart", controllers.GetCartListHandle)
			Login.PUT("/cart", controllers.UpdateCartItemHandle)
			Login.DELETE("/cart/:book_id", controllers.DeleteCartItemHandle)
			Login.DELETE("/cart", controllers.ClearCartHandle)
		}
	}
	v2 := r.Group("/page")
	{
		v2.GET("/HomePage", controllers.HomePageHandle)
		v2.GET("/LoginPage", controllers.LoginPageHandle)
		v2.GET("/RegisterPage", controllers.RegisterPageHandle)
		v2.GET("/AdminLoginPage", controllers.AdminLoginPageHandle)
		v2.GET("/AdminRegisterPage", controllers.AdminRegisterPageHandle)
		v2.GET("/BooksPage", controllers.BookDetailPageHandle)
		User := v2.Use(middlewares.CookieAuthMiddleware())
		{
			User.GET("/ProfilePage", controllers.ProfilePageHandle)
		}
		Admin := v2.Use(middlewares.CookieAuthMiddleware(), middlewares.AdminOnlyMiddleware())
		{
			Admin.GET("/AdminHomePage", controllers.AdminHomePageHandle)
			Admin.GET("/AddBookPage", controllers.AddBookPageHandle)
			Admin.GET("/DeleteBookPage", controllers.DeleteBookPageHandle)
			Admin.GET("/UpdateBookPage", controllers.UpdateBookPageHandle)
		}
	}
	v3 := r.Group("/admin")
	v3.Use(middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware())
	{
		//v3.GET("/status", controllers.AdminStatusHandle) //运营统计
		book := v3.Group("/book") //管理员对书籍的增删改查
		{
			book.POST("/add", controllers.AdminAddBookHandle)
			book.GET("/getbook/:book_id", controllers.GetBookParamHandle)
			book.DELETE("/delete/:book_id", controllers.AdminDeleteBookHandle)
			book.POST("/update", controllers.AdminUpdateBookHandle)
		}
		//tags := v3.Group("/tags")
		//{
		//	tags.POST("/add", controllers.AdminAddTagHandle)
		//	tags.DELETE("/delete/:tag_id", controllers.AdminDeleteTagHandle)
		//}
		//user := v3.Group("/user")
		//{
		//	user.GET("/list", controllers.AdminListUserHandle)
		//	user.GET("/manage", controllers.AdminManageUserHandle)
		//}
		//order := v3.Group("/order")
		//{
		//	order.GET("/list", controllers.AdminListOrderHandle)
		//	order.GET("/manage", controllers.AdminManageOrderHandle)
		//}
	}
	zap.L().Info("SetUp Server ...")
	return r
}
