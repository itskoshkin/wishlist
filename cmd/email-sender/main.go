package main

import "wishlist/internal/emailsender"

func main() {
	emailsender.Load().Run()
}
