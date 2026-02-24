package apiModels

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Message string `json:"message"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func Error(ctx *gin.Context, code int, message string) {
	ctx.AbortWithStatusJSON(code, APIError{
		Code:    code,
		Message: message,
	})
}

func ErrorWithDetails(ctx *gin.Context, code int, message, details string) {
	ctx.AbortWithStatusJSON(code, APIError{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func InternalError(ctx *gin.Context, details string) {
	log.Printf("[%s] Internal error: %s", ctx.GetString("request_id"), details)
	ErrorWithDetails(ctx, 500, "Internal server error", fmt.Sprintf("Your request ID is %s", ctx.GetString("request_id")))
}
