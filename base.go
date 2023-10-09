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

type PageVO struct {
	List  []interface{} `json:"data"`
	Draw  interface{}   `json:"draw"`
	Total int64         `json:"total"`
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

type PageForm struct {
	PageIndex int64 `json:"pageIndex" binding:"required" label:"分页页码"`
	PageSize  int64 `json:"pageSize" binding:"required" label:"每页大小"`
	//StartRow  int64 `json:"startRow"`
}

func (p *PageForm) GetPageIndex() int64 {
	if 1 > p.PageIndex {
		return 1
	}
	return p.PageIndex
}

func (p *PageForm) GetPageSize() int64 {
	if 1 > p.PageSize {
		return 1
	} else if 2000 < p.PageSize {
		return 2000
	}
	return p.PageSize
}

func (p *PageForm) GetStartRow() int64 {
	pi := p.GetPageIndex()
	ps := p.PageSize
	return (pi - 1) * ps
}
