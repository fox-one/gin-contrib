package gin_helper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fox-one/gin-contrib/errors"
	"github.com/gin-gonic/gin"
)

var (
	badRequest = errors.New(1, "invalid operation", http.StatusBadRequest)
	serverErr  = errors.New(2, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
)

const responseJsonKeyTransformerContextKey = "_gin_helper_response_json_key_transformer"

func TransformResponseJsonKey(fn JsonKeyTransformer) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(responseJsonKeyTransformerContextKey, fn)
	}
}

func Response(c *gin.Context, code int, obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		log.Panic(err)
	}

	if v, ok := c.Get(responseJsonKeyTransformerContextKey); ok {
		if fn, ok := v.(JsonKeyTransformer); ok && fn != nil {
			data = TransformJsonKeys(data, fn)
		}
	}

	c.AbortWithStatusJSON(code, json.RawMessage(data))
}

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

	Response(c, http.StatusOK, resp)
}

func Data(c *gin.Context, args ...interface{}) {
	var data interface{}

	switch len(args) {
	case 1:
		data = args[0]
	default:
		m := make(gin.H, len(args)/2)
		for idx := 0; idx+1 < len(args); idx += 2 {
			k, v := args[idx].(string), args[idx+1]
			m[k] = v
		}

		data = m
	}

	Response(c, http.StatusOK, gin.H{"code": 0, "data": data})
}

func Fail(c *gin.Context, status, code int, msg string, data interface{}, hints ...interface{}) {
	resp := gin.H{
		"code": code,
		"msg":  msg,
	}

	if data != nil {
		resp["data"] = data
	}

	if len(hints) > 0 && IsDebug() {
		switch v := hints[0].(type) {
		case string:
			resp["hint"] = fmt.Sprintf(v, hints[1:]...)
		case error:
			resp["hint"] = v.Error()
		case fmt.Stringer:
			resp["hint"] = v.String()
		default:
			log.Panicln("unsupported hint", v)
		}
	}

	Response(c, status, resp)
}

func unpackErrWithDefault(err error, status, code int, msg string) (int, int, string) {
	if e, ok := err.(errors.Error); ok {
		code, msg = e.Code(), e.Message()

		if re, ok := err.(errors.RequestError); ok {
			status = re.StatusCode()
		}
	} else if err != nil {
		msg = msg + ": " + err.Error()
	}

	return status, code, msg
}

func FailError(c *gin.Context, err error, hints ...interface{}) {
	status, code, msg := 400, 1, "invalid operation"
	status, code, msg = unpackErrWithDefault(err, status, code, msg)
	Fail(c, status, code, msg, nil, hints...)
}

func FailServer(c *gin.Context, err error, hints ...interface{}) {
	status, code, msg := 500, 2, http.StatusText(http.StatusInternalServerError)
	status, code, msg = unpackErrWithDefault(err, status, code, msg)
	Fail(c, status, code, msg, nil, hints...)
}

func FailErrorWithData(c *gin.Context, err error, data interface{}) {
	status, code, msg := 400, 1, "invalid operation"
	status, code, msg = unpackErrWithDefault(err, status, code, msg)
	Fail(c, status, code, msg, data)
}

// pagination
func OkWithPagination(c *gin.Context, cursor string, args ...interface{}) {
	resp := make(gin.H, len(args)/2+1)
	for idx := 0; idx+1 < len(args); idx += 2 {
		k, v := args[idx].(string), args[idx+1]
		resp[k] = v
	}

	resp["pagination"] = map[string]interface{}{
		"next_cursor": cursor,
		"has_next":    len(cursor) > 0,
	}

	Response(c, http.StatusOK, resp)
}

type JSONString string

func (s JSONString) MarshalJSON() ([]byte, error) {
	return []byte(s), nil
}

func OKWithString(c *gin.Context, text string) {
	OK(c, JSONString(text))
}
