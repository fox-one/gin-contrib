package limiter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const defaultKey = "limiter_context_key"

func (limiter *Limiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(defaultKey, limiter)
	}
}

type WeightFunc func(c *gin.Context) int

func Weight(w int) WeightFunc {
	return func(c *gin.Context) int {
		return w
	}
}

func Available(group string, f WeightFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := c.MustGet(defaultKey).(*Limiter)
		w, key := f(c), c.ClientIP()
		remain, err := limiter.Available(key, group, w)
		if err != nil {
			log.Errorf("check rate limit failed: %s", err)
		}

		if remain < 0 {
			c.AbortWithStatus(http.StatusTooManyRequests)
		} else {
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remain))
		}
	}
}
