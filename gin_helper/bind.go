package gin_helper

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func BindJson(c *gin.Context, obj interface{}) error {
	return c.ShouldBindBodyWith(obj, binding.JSON)
}

func BindQuery(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
}
