package controllers

import (
	"Project1_Shop/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleResponse(c *gin.Context, statusCode models.ResCode) {
	Rd := &models.ResponseData{
		Code: statusCode,
		Msg:  statusCode.Msg(),
		Data: nil,
	}
	c.JSON(http.StatusOK, Rd)
}

func HandleResponseWithData(c *gin.Context, statusCode models.ResCode, data interface{}) {
	Rd := &models.ResponseData{
		Code: statusCode,
		Msg:  statusCode.Msg(),
		Data: data,
	}
	c.JSON(http.StatusOK, Rd)
}

func HandleSuccess(c *gin.Context, data interface{}) {
	HandleResponseWithData(c, models.CodeSuccess, data)
}
