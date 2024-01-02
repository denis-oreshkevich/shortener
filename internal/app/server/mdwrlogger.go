package server

import (
	"time"

	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging func to log the request details and it's execution time.
func Logging(c *gin.Context) {
	r := c.Request
	start := time.Now()

	c.Next()

	duration := time.Since(start)

	logger.Log.Info("request", zap.String("uri", r.RequestURI),
		zap.String("method", r.Method),
		zap.Int("status", c.Writer.Status()),
		zap.Duration("duration", duration),
		zap.Int("size", c.Writer.Size()))
}
