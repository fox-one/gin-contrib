package gin_helper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Log(in *time.Location) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().In(in)
		c.Next()

		end := time.Now()
		status := c.Writer.Status()
		method := c.Request.Method
		uri := c.Request.URL.String()

		content := fmt.Sprintf("[%d] %-4s %s", status, method, uri)

		entry := log.WithFields(log.Fields{
			"ts": start.Format(time.RFC3339),
			"lt": end.Sub(start),
			"ip": c.ClientIP(),
			"ua": c.Request.UserAgent(),
		})

		if status >= http.StatusOK && status < 300 {
			entry.Info(content)
		} else if status == http.StatusNotFound {
			entry.Warn(content)
		} else {
			entry.Error(content)
		}
	}
}
