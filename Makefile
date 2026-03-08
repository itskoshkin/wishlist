BIN_NAME ?= wishlist
MAIN_PATH ?= ./cmd/main.go
CURL_TEST_PATH = ./tests/test_endpoints_with_curl.sh
SWAG_ARGS ?= -o docs -g main.go -d ./cmd,./internal/api/controllers,./internal/models,./internal/api/errors --parseInternal

.PHONY: build run test-curl swagger swagger-clean migrate clean

build: ## Build application binary
	go build -o $(BIN_NAME) $(MAIN_PATH)

run: ## Run app
	go run $(MAIN_PATH)

test-curl: ## Run interactive curl test script
	bash $(CURL_TEST_PATH)

swagger: ## Generate Swagger docs
	swag init $(SWAG_ARGS)

swagger-clean: ## Remove generated Swagger files
	rm -r ./docs/

migrate: ## Apply database migrations
	goose postgres "$(GOOSE_DSN)" -dir migrations up

clean:
	rm -r -- $(BIN_NAME) app*.log logs/
