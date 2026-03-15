BIN_NAME ?= wishlist
MAIN_PATH ?= ./cmd/main.go
CURL_TEST_PATH = ./tests/test_endpoints_with_curl.sh
E2E_BASE_URL ?= http://localhost:8080/api/v1
SWAG_ARGS ?= -o docs -g main.go -d ./cmd,./internal/api/controllers,./internal/models,./internal/api/errors --parseInternal
GOOSE_DSN ?= "postgres://username:password@localhost:5432/wishlist?sslmode=disable"

.PHONY: build run test test-curl test-unit test-integration test-e2e swagger swagger-clean docker-build docker-run compose-up compose-down migrate clean

build: ## Build application binary
	go build -o $(BIN_NAME) $(MAIN_PATH)

run: ## Run app
	go run $(MAIN_PATH)

test: test-unit test-integration test-e2e ## Run all tests

test-curl: ## Run interactive curl test script
	bash $(CURL_TEST_PATH)

test-unit: ## Run unit tests
	go test -race -count=1 ./...

test-integration: ## Run integration tests
	go test -count=1 -tags=integration ./...

test-e2e: ## Run end-to-end test
	E2E_BASE_URL="$(E2E_BASE_URL)" go test -v -count=1 -tags=e2e ./tests/e2e

swagger: ## Generate Swagger docs
	swag init $(SWAG_ARGS)

swagger-clean: ## Remove generated Swagger files
	rm -r ./docs/

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker image
	docker run --rm -p 8080:8080 -v "$(PWD)/config.yaml:/app/config.yaml:ro" $(DOCKER_IMAGE)

compose-up: ## Start full stack in Docker (application, postgres, redis, minio)
	docker compose up -d

compose-down: ## Start full stack in Docker (application, postgres, redis, minio)
	docker compose down

migrate: ## Apply database migrations
	goose postgres "$(GOOSE_DSN)" -dir migrations up

clean:
	rm -r -- $(BIN_NAME) app*.log logs/
