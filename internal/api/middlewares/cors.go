package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins ...string) gin.HandlerFunc {
	origins := "*"
	if len(allowedOrigins) > 0 {
		origins = strings.Join(allowedOrigins, ", ")
	}

	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", origins)
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Max-Age", "86400") // 1 day

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	}
}
