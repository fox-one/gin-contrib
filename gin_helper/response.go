package gin_helper

import (
	"fmt"
	"net/http"

	"github.com/fox-one/gin-contrib/errors"
	"github.com/gin-gonic/gin"
)

var (
	badRequest = errors.New(1, "invalid operation", http.StatusBadRequest)
	serverErr  = errors.New(2, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
)

// OK() 将 args 构造成 map 然后转成 json
// 如果 args length 为 1，直接解析成 json
func OK(c *gin.Context, args ...interface{}) {
	if len(args) == 1 {
		c.JSON(http.StatusOK, args[0])
		return
	}

	resp := make(gin.H, len(args)/2)
	for idx := 0; idx+1 < len(args); idx += 2 {
		k, v := args[idx].(string), args[idx+1]
		resp[k] = v
	}

	c.JSON(http.StatusOK, resp)
}

func Fail(c *gin.Context, status, code int, msg string, hints ...interface{}) {
	resp := gin.H{
		"code": code,
		"msg":  msg,
	}

	if len(hints) > 0 && IsDebug() {
		resp["hint"] = fmt.Sprintf(hints[0].(string), hints[1:]...)
	}

	c.AbortWithStatusJSON(status, resp)
}

func FailError(c *gin.Context, err error, hints ...interface{}) {
	status, code, msg := 400, 1, "invalid operation"

	if e, ok := err.(errors.Error); ok {
		code, msg = e.Code(), e.Message()

		if re, ok := err.(errors.RequestError); ok {
			status = re.StatusCode()
		}
	} else if err != nil {
		msg = msg + ": " + err.Error()
	}

	Fail(c, status, code, msg, hints...)
}

func FailServer(c *gin.Context, err error, hints ...interface{}) {
	status, code, msg := 500, 2, http.StatusText(http.StatusInternalServerError)

	if e, ok := err.(errors.Error); ok {
		code, msg = e.Code(), e.Message()

		if re, ok := err.(errors.RequestError); ok {
			status = re.StatusCode()
		}
	} else if err != nil {
		msg = msg + ": " + err.Error()
	}

	Fail(c, status, code, msg, hints...)
}
