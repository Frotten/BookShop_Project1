package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Login.html", nil)
}

func RegisterPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Register.html", nil)
}

func AdminLoginPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "AdminLogin.html", nil)
}

func AdminRegisterPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "AdminRegister.html", nil)
}

func AdminHomePageHandle(c *gin.Context) {
	AdminHomePageHandleWithInfo(c, nil)
}

func AdminHomePageHandleWithInfo(c *gin.Context, Info interface{}) {
	c.HTML(http.StatusOK, "AdminHome.html", Info)
}

func ProfilePageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Profile.html", nil)
}

func HomePageHandleWithInfo(c *gin.Context, Info interface{}) {
	c.HTML(http.StatusOK, "Home.html", Info)
}
