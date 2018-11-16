package gin_helper

import (
	"github.com/gin-gonic/gin"
)

func IsDebug() bool {
	return gin.Mode() == gin.DebugMode
}
