package gin_helper

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func ExtractBody(c *gin.Context) (body []byte, err error) {
	if cb, ok := c.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}

	if body == nil {
		body, err = ioutil.ReadAll(c.Request.Body)
		if err == nil {
			c.Set(gin.BodyBytesKey, body)
		}
	}

	return
}

func BindJson(c *gin.Context, obj interface{}) error {
	return c.ShouldBindBodyWith(obj, binding.JSON)
}

func BindQuery(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
}

func BindUri(c *gin.Context, obj interface{}) error {
	return c.ShouldBindUri(obj)
}
