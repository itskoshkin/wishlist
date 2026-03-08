BIN_NAME ?= wishlist
MAIN_PATH ?= ./cmd/main.go
CURL_TEST_PATH = ./tests/test_endpoints_with_curl.sh

.PHONY: build run test-curl migrate clean

build: ## Build application binary
	go build -o $(BIN_NAME) $(MAIN_PATH)

run: ## Run app
	go run $(MAIN_PATH)

test-curl: ## Run interactive curl test script
	bash $(CURL_TEST_PATH)

migrate: ## Apply database migrations
	goose postgres "$(GOOSE_DSN)" -dir migrations up

clean:
	rm -r -- $(BIN_NAME) app*.log logs/
