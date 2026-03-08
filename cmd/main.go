package main

import "wishlist/internal/app"

// @title Wishlist API
// @description API for managing wishlists
// @version 1.0
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter token as "Bearer <access_token>"
func main() {
	app.Load().Run()
}
