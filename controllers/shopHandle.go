package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomePageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Home.html", nil)
}

func LoginPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Login.html", nil)
}

func ProfilePageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Profile.html", nil)
}

func RegisterPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Register.html", nil)
}
