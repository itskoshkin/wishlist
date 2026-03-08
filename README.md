# WIP

### TODO
- [x] Add Redis caching for JWT tokens
- [x] Add email sending service for password reset and account verification
- [x] Add S3 for user avatars and other media
- [x] Add Wish and Wishlist entities to API
  - [ ] Add pagination for lists
- [x] Add goose migrations
- [x] Add build instructions
- [ ] Add Swagger documentation for API
- [ ] Add unit and integration tests
- [ ] Dockerize, set up CI/CD and deploy
- [ ] Add message broker for... Emails? Price change notifications?

---

# Temp

## How to build

1. Install dependencies
    ```bash
    go install github.com/pressly/goose/v3/cmd/goose@latest
    ```
2. Run `make` 
    ```bash
    export GOOSE_DSN="postgres://username:password@localhost:5432/wishlist?sslmode=disable"
    make migrate build run
    ```

