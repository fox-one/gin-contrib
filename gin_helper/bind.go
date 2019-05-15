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

const (
	requestJsonKeyTransformerContextKey = "_gin_helper_request_json_key_transformer"
	requestTransformedBodyContextKey    = "_gin_helper_request_transformed_body"
)

func TransformRequestJsonKey(fn JsonKeyTransformer) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(requestJsonKeyTransformerContextKey, fn)
	}
}

func BindJson(c *gin.Context, obj interface{}) error {
	binding := binding.JSON

	if v, ok := c.Get(requestTransformedBodyContextKey); ok {
		return binding.BindBody(v.([]byte), obj)
	}

	body, err := ExtractBody(c)
	if err != nil {
		return err
	}

	if v, ok := c.Get(requestJsonKeyTransformerContextKey); ok {
		if fn, _ := v.(JsonKeyTransformer); fn != nil {
			body = TransformJsonKeys(body, fn)
			c.Set(requestTransformedBodyContextKey, body)
		}
	}

	return binding.BindBody(body, obj)
}

func BindQuery(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
}

func BindUri(c *gin.Context, obj interface{}) error {
	return c.ShouldBindUri(obj)
}
