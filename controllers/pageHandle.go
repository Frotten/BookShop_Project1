package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Login.html", nil)
}

func ProfilePageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Profile.html", nil)
}

func RegisterPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Register.html", nil)
}

func HomePageHandleWithInfo(c *gin.Context, Info interface{}) {
	c.HTML(http.StatusOK, "Home.html", Info)
}
