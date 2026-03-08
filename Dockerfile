FROM golang:1.26-alpine AS builder

ARG BIN_NAME="wishlist"
ARG MAIN_PATH="cmd/main.go"
ARG GO_BUILD_FLAGS="-s -w"
# ARG UPX_FLAGS="--best --lzma"

WORKDIR /src

RUN apk add --no-cache git
# RUN apk add --no-cache upx

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init -o docs -d cmd,internal/models,internal/api

RUN CGO_ENABLED=0 go build -ldflags="$GO_BUILD_FLAGS" -o $BIN_NAME $MAIN_PATH
# RUN upx $UPX_FLAGS $BIN_NAME

FROM alpine:latest

ARG BIN_NAME="wishlist"

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /src/$BIN_NAME .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /src/migrations ./migrations
COPY --from=builder /src/example_config.yaml ./config.yaml
COPY --from=builder /src/docs ./docs

EXPOSE 8080

CMD ["./wishlist"]
