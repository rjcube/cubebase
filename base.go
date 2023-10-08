package cubebase

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一返回结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Page struct {
	Data  []any `json:"data"`
	Draw  any   `json:"draw"`
	Total int64 `json:"total"`
}

// CubeResponse 统一返回结构
func CubeResponse(c *gin.Context, code int, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func CubeResponseSuccess(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data:    obj,
	})
}

func CubeResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func CubeResponseErrorForServerErr(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: message,
		Data:    nil,
	})
}
