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
	r.GET("/HomePage", controllers.HomePageHandle)
	r.GET("/LoginPage", controllers.LoginPageHandle)
	r.GET("/RegisterPage", controllers.RegisterPageHandle)
	r.GET("/ProfilePage", controllers.ProfilePageHandle)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", controllers.SignUpHandler)
		v1.POST("/login", controllers.LoginHandler)
		v1.Use(middlewares.JWTAuthMiddleware())
		{

		}
	}
	zap.L().Info("SetUp Server ...")
	return r
}
