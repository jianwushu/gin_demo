package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Body{Code: 0, Message: "success", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Body{Code: 0, Message: "success", Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Body{Code: status, Message: message})
}
