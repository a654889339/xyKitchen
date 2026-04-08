package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func JSON(c *gin.Context, code int, payload gin.H) {
	if payload == nil {
		payload = gin.H{}
	}
	payload["code"] = code
	c.JSON(http.StatusOK, payload)
}

func Err(c *gin.Context, status int, code int, message string) {
	c.JSON(status, gin.H{"code": code, "message": message})
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data})
}

func OKMsg(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": message})
}

func OKDataMsg(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "message": message})
}
