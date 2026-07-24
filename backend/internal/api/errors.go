package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ErrorJSON(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Code: statusCode, Message: message})
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "panic", r, "path", c.Request.URL.Path)
				ErrorJSON(c, 500, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
