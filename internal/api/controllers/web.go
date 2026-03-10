package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

type WebController struct{}

func NewWebController() *WebController {
	return &WebController{}
}

func (h *WebController) Index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index", gin.H{})
}

func (h *WebController) Wishes(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "wishes", gin.H{})
}

func (h *WebController) WishlistByID(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "list-detail", gin.H{"list_id": ctx.Param("list_id")})
}

func (h *WebController) WishlistBySharedSlug(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "list-detail", gin.H{"shared_slug": ctx.Param("slug")})
}

func (h *WebController) NotFound(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, viper.GetString(config.ApiBasePath)) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "route not found"})
		return
	}

	ctx.HTML(http.StatusNotFound, "404", gin.H{})
}
