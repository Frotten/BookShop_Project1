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
		v1.GET("/topSale", controllers.TopSaleHandle)
		v1.GET("/topScore", controllers.TopScoreHandle)
		v1.GET("/comments", controllers.CommentsHandle)
		v1.GET("/getBookTitle/:title", controllers.GetBookByTitleHandle)
		v1.GET("/getBooksBySaleJSON", controllers.GetBooksBySaleJSON)
		v1.GET("/getBooksBySaleJSON/:page", controllers.SaleBookHandle)
		v1.GET("/seckill/list", controllers.SeckillListHandle)
		v1.GET("/seckill/:id", controllers.SeckillDetailHandle)
		Login := v1.Use(middlewares.JWTAuthMiddleware(), middlewares.CheckLoginOnlyMiddleware())
		{
			Login.GET("/userInfo", controllers.GetUserInfoHandle)
			Login.GET("/userComments", controllers.GetUserCommentsHandle)
			Login.POST("/rateBook", controllers.RateBookHandle)
			Login.POST("/comment", controllers.CommentHandle)
			Login.POST("/comment/like", controllers.CommentLikeHandle)
			Login.GET("/userRatings", controllers.GetUserRatingsHandle)
			Login.POST("/cart", controllers.AddBookToCartHandle)
			Login.GET("/cart", controllers.GetCartListHandle)
			Login.PUT("/cart", controllers.UpdateCartItemHandle)
			Login.DELETE("/cart/:book_id", controllers.DeleteCartItemHandle)
			Login.DELETE("/cart", controllers.ClearCartHandle)
			Login.POST("/orderCreate", controllers.CreateOrderHandle)
			Login.GET("/userOrders", controllers.GetUserOrderHandle)
			Login.GET("/orderDetail/:order_id", controllers.GetOrderDetailHandle)
			Login.POST("/orderPay", controllers.OrderPayHandle)
			Login.POST("/orderCancel", controllers.OrderCancelHandle)
			Login.POST("/orderConfirm", controllers.OrderConfirmHandle)
			Login.POST("/seckill/do", controllers.SeckillHandle)
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
		v2.GET("/SeckillPage", controllers.SeckillPageHandle)
		User := v2.Use(middlewares.CookieAuthMiddleware())
		{
			User.GET("/ProfilePage", controllers.ProfilePageHandle)
			User.GET("/OrderDetailPage", controllers.OrderDetailPageHandle)
		}
		Admin := v2.Use(middlewares.CookieAuthMiddleware(), middlewares.AdminOnlyMiddleware())
		{
			Admin.GET("/AdminHomePage", controllers.AdminHomePageHandle)
			Admin.GET("/AddBookPage", controllers.AddBookPageHandle)
			Admin.GET("/DeleteBookPage", controllers.DeleteBookPageHandle)
			Admin.GET("/UpdateBookPage", controllers.UpdateBookPageHandle)
			Admin.GET("/OrderShipmentPage", controllers.OrderShipmentPageHandle)
			Admin.GET("/SeckillManagePage", controllers.SeckillManagePageHandle)
		}
	}
	v3 := r.Group("/admin")
	v3.Use(middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware())
	{
		book := v3.Group("/book")
		{
			book.POST("/add", controllers.AdminAddBookHandle)
			book.GET("/getbook/:book_id", controllers.GetBookParamHandle)
			book.DELETE("/delete/:book_id", controllers.AdminDeleteBookHandle)
			book.POST("/update", controllers.AdminUpdateBookHandle)
			book.GET("/getBookTitle/:title", controllers.GetBookByTitleHandle)
		}
		order := v3.Group("/order")
		{
			order.GET("/list", controllers.GetShipOrderHandle)
			order.POST("/orderShip", controllers.OrderShipHandle)
		}
		seckill := v3.Group("/seckill")
		{
			seckill.GET("/list", controllers.AdminSeckillListHandle)
			seckill.POST("/create", controllers.AdminCreateSeckillHandle)
			seckill.POST("/down/:id", controllers.AdminDownSeckillHandle)
		}
	}
	zap.L().Info("SetUp Server ...")
	return r
}
