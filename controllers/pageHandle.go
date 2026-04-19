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

func HomePageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "Home.html", nil)
}

func AddBookPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "AddBook.html", nil)
}

func DeleteBookPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "DeleteBook.html", nil)
}

func UpdateBookPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "UpdateBook.html", nil)
}

func BookDetailPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "BookDetail.html", nil)
}

func OrderDetailPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "OrderDetail.html", nil)
}

func OrderShipmentPageHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "OrderShipment.html", nil)
}
