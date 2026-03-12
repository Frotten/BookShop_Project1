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
		//Login := v1.Use(middlewares.JWTAuthMiddleware())
		//{
		//
		//}
	}
	v2 := r.Group("/page")
	{
		v2.GET("/HomePage", controllers.SBHbasic)
		v2.GET("/HomePage/:page", controllers.ScoreBookHandle)
		v2.GET("/LoginPage", controllers.LoginPageHandle)
		v2.GET("/RegisterPage", controllers.RegisterPageHandle)
		v2.GET("/AdminLoginPage", controllers.AdminLoginPageHandle)
		v2.GET("/AdminRegisterPage", controllers.AdminRegisterPageHandle)
		User := v2.Use(middlewares.CookieAuthMiddleware())
		{
			User.GET("/ProfilePage", controllers.ProfilePageHandle)
		}
		Admin := v2.Use(middlewares.CookieAuthMiddleware(), middlewares.AdminOnlyMiddleware())
		{
			Admin.GET("/AdminHomePage", controllers.AdminHomePageHandle)
			Admin.GET("/AddBookPage", controllers.AddBookPageHandle)
		}
	}
	v3 := r.Group("/admin")
	v3.Use(middlewares.JWTAuthMiddleware(), middlewares.AdminOnlyMiddleware())
	{
		//v3.GET("/status", controllers.AdminStatusHandle) //运营统计
		book := v3.Group("/book") //管理员对书籍的增删改查
		{
			book.POST("/add", controllers.AdminAddBookHandle)
			//book.GET("/list", controllers.AdminListBookHandle)
			//book.DELETE("/delete", controllers.AdminDeleteBookHandle)
		}
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
