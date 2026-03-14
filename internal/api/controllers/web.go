package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

type WebController struct {
	router      *gin.Engine
	userService UserService
}

func NewWebController(e *gin.Engine, us UserService) *WebController {
	return &WebController{router: e, userService: us}
}

func (ctrl *WebController) RegisterRoutes() {
	ctrl.router.GET("/", ctrl.Index)
	ctrl.router.GET("/wishlists", ctrl.Wishes)
	ctrl.router.GET("/users/:username", ctrl.PublicUserWishlists)
	ctrl.router.GET("/wishlist/:list_id", ctrl.Wishlist)
	ctrl.router.GET("/shared/:slug", ctrl.WishlistBySharedLink)
	ctrl.router.GET("/verify-email", ctrl.VerifyEmail)
	ctrl.router.GET("/reset-password", ctrl.ResetPassword)
	ctrl.router.NoRoute(ctrl.NotFound)
}

func (ctrl *WebController) Index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index", gin.H{})
}

func (ctrl *WebController) Wishes(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "wishes", gin.H{})
}

func (ctrl *WebController) PublicUserWishlists(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "wishes", gin.H{"viewed_username": ctx.Param("username")})
}

func (ctrl *WebController) Wishlist(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "list-detail", gin.H{"list_id": ctx.Param("list_id")})
}

func (ctrl *WebController) WishlistBySharedLink(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "list-detail", gin.H{"shared_slug": ctx.Param("slug")})
}

func (ctrl *WebController) VerifyEmail(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.Redirect(http.StatusFound, "/?error=invalid_token")
		return
	}

	if err := ctrl.userService.VerifyEmail(ctx, token); err != nil {
		ctx.Redirect(http.StatusFound, "/?error=verification_failed")
		return
	}

	ctx.Redirect(http.StatusFound, "/?verified=true")
}

func (ctrl *WebController) ResetPassword(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "reset-password", gin.H{})
}

func (ctrl *WebController) NotFound(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, viper.GetString(config.ApiBasePath)) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "route not found"})
		return
	}

	ctx.HTML(http.StatusNotFound, "404", gin.H{})
}
