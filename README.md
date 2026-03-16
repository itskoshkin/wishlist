# Wishlist

Wishlist is a service for storing gift ideas, organizing them into wishlists, and sharing them with others. The goal is to keep everything in one place instead of scattered notes, bookmarks and chat messages.

[wishlist.itskoshkin.ru](https://wishlist.itskoshkin.ru)

## Features

- Create and manage wishlists
- Share them by direct link or through a public profile
- Discover other users and view their wishes

<details>
<summary><h3>Technical features</h3></summary>

- User registration, email verification and password reset
- CRUD for `List` and `Wish` entities
- User avatars and wish images stored in S3
- Built-in web interface alongside a REST API

</details>

<details>
<summary><h2>Technical Details</h2></summary>

### Stack

Go, [Gin](https://github.com/gin-gonic/gin), [Viper](https://github.com/spf13/viper), [pgx](https://github.com/jackc/pgx) (PostgreSQL), [go-redis](https://github.com/redis/go-redis) (Redis), [minio-go](https://github.com/minio/minio-go) (MinIO), [kafka-go](https://github.com/segmentio/kafka-go) (Kafka), [amqp091-go](https://github.com/rabbitmq/amqp091-go) (RabbitMQ), [swaggo/swag](https://github.com/swaggo/swag), and [Goose](https://github.com/pressly/goose).

### Architecture

The project has two runtimes:

- `api` serves the REST API and web UI and works with PostgreSQL, Redis, and MinIO
- `email-sender` consumes email events from Kafka or RabbitMQ and sends SMTP emails

If `app.broker.type` is set to `none`, the API can send emails directly without `email-sender`.

<details>
<summary><h3>Project structure</h3></summary>

```text
.
├── cmd
│   ├── api/main.go             # API and web application entrypoint
│   └── email-sender/main.go    # Background runtime for sending emails from broker events
├── docs/                       # Generated Swagger/OpenAPI documentation
├── internal/
│   ├── api/                    # HTTP handlers, middleware, and API-specific errors
│   ├── app/                    # Application bootstrap and dependency wiring
│   ├── broker/                 # Kafka and RabbitMQ adapters
│   ├── config/                 # Configuration loading and validation
│   ├── emailsender/            # Email sender runtime logic
│   ├── events/                 # Event types, payloads, topics, and event publishing
│   ├── logger/                 # Logging setup and helpers
│   ├── models/                 # Shared domain and transport models
│   ├── services/               # Core business logic
│   ├── storage/                # Database, cache, and object storage operations
│   └── utils/                  # Shared utility packages
├── migrations/                 # Goose database migrations
├── pkg/                        # Infrastructure clients
├── static/                     # Frontend assets and templates
└── tests/                      # End-to-end tests
```

</details>

</details>

<details>
<summary><h2>Build & Run</h2></summary>

<details>
<summary><h3>Quick start</h3></summary>

#### Local

1. Copy config template and fill your parameters:
   ```bash
   cp example_config.yaml config.yaml
   ```
2. Start infrastructure services if not already present (see [Infrastructure containers](#infrastructure-containers))
3. Create database and edit `GOOSE_DSN` in `Makefile`
4. Run locally:
   ```bash
   make migrate build run
   ```

#### Docker Compose

1. Copy config template and fill your parameters:
   ```bash
   cp example_config.yaml config.yaml
   ```
2. Start infrastructure services if not already present (see [Infrastructure containers](#infrastructure-containers))
3. Run the full stack:
   ```bash
   CONFIG_PATH=./config.yaml docker compose up -d --build
   ```
4. Or run only API and email sender with an external broker:
   ```bash
   CONFIG_PATH=./config.yaml \
   APP_BROKER_TYPE=kafka \
   APP_BROKER_KAFKA_BROKERS=10.0.0.30:9092 \
   docker compose up -d --build api email-sender
   ```

</details>

<details>
<summary><h3>Requirements</h3></summary>

- Go 1.26 or newer
  ```bash
  brew install go # on macOS
  ```
  ```bash
  add-apt-repository ppa:longsleep/golang-backports -y && apt update && apt install golang -y # on Debian-based Linux
  ```
- [PostgreSQL](https://www.postgresql.org/download/), [Redis](https://redis.io/docs/latest/operate/oss_and_stack/install/archive/install-redis/) and [MinIO](https://github.com/minio/minio?tab=readme-ov-file)
- [Kafka](https://kafka.apache.org/quickstart/) or [RabbitMQ](https://www.rabbitmq.com/docs/download) if broker-based email delivery is enabled
- Goose
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- Swag
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```
- Docker and Docker Compose if you want to run the project in containers
  ```bash
  /bin/bash -c "$(curl -fsSL https://get.docker.com)" # one-liner for Debian-based Linux
  ```

> 💡 If you want to run infrastructure services with Docker instead of installing them manually, see [Infrastructure containers](#infrastructure-containers).

</details>

<details>
<summary><h3>Build</h3></summary>

Build API:
```bash
go build -o bin/api ./cmd/api
```

Build email sender:
```bash
go build -o bin/email-sender ./cmd/email-sender
```

Build both:
```bash
go build ./cmd/api ./cmd/email-sender ./...
```

</details>

<details>
<summary><h3>Config</h3></summary>

Copy the template and fill it manually:
```bash
cp example_config.yaml config.yaml
```

Important config sections:

- `app.api`
- `app.api.auth`
- `app.database`
- `app.redis`
- `app.minio`
- `app.email`
- `app.webapp`
- `app.broker`

Broker modes:

- `none` — the API sends emails directly via SMTP
- `kafka` — the API publishes email events to Kafka, `email-sender` consumes them
- `rabbitmq` — the API publishes email events to RabbitMQ, `email-sender` consumes them

For Docker, keep `config.yaml` on the host and mount it into the container:
```bash
-v "$(pwd)/config.yaml:/app/config.yaml:ro"
```

</details>

<details>
<summary><h3>Migrations</h3></summary>

1. Create database first
2. Run migrations:
    ```bash
    make migrate
    ```
   or
    ```bash
    goose postgres "postgres://username:password@localhost:5432/dbname?sslmode=disable" -dir migrations up
    ```
3. Check status:
    ```bash
    goose postgres "postgres://username:password@localhost:5432/dbname?sslmode=disable" -dir migrations status
    ```

</details>

<details>
<summary><h3>Run</h3></summary>

Make sure infrastructure services are already running. Example commands for running them in Docker containers are listed in [Infrastructure containers](#infrastructure-containers).

Run API:
```bash
go run ./cmd/api
```

Run email sender:
```bash
go run ./cmd/email-sender
```

</details>

</details>

<details>
<summary><h2>Docker</h2></summary>

<details>
<summary><h3>Containers</h3></summary>

Run API and email sender as separate containers.

#### API

```bash
docker build \
  --build-arg MAIN_PATH=./cmd/api \
  --build-arg BIN_NAME=wishlist-api \
  -t wishlist-api:latest .
```

```bash
docker run -d \
  --name wishlist-api \
  --restart unless-stopped \
  -p 8080:8080 \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  wishlist-api:latest
```

#### Sender

```bash
docker build \
  --build-arg MAIN_PATH=./cmd/email-sender \
  --build-arg BIN_NAME=wishlist-email-sender \
  -t wishlist-email-sender:latest .
```

```bash
docker run -d \
  --name wishlist-email-sender \
  --restart unless-stopped \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  wishlist-email-sender:latest
```

### Infrastructure containers

#### Postgres

```bash
docker volume create postgres

docker run -dit \
  --name postgres \
  -p 5432:5432 \
  -e POSTGRES_USER="username" \
  -e POSTGRES_PASSWORD="********" \
  -e POSTGRES_DB="default" \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -v postgres:/var/lib/postgresql/data \
  postgres
```

#### Redis

```bash
docker volume create redis

docker run -dit \
  --name redis \
  -p 6379:6379 \
  -p 6380:8001 \
  -e REDIS_ARGS="--requirepass ********" \
  -v redis:/data \
  redis/redis-stack:latest
```

#### MinIO

```bash
docker run -d \
  --name minio \
  -p 9003:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER="username" \
  -e MINIO_ROOT_PASSWORD="********" \
  -v /HSE/MinIO/Data:/data \
  --restart always \
  quay.io/minio/minio \
  server /data --console-address ":9001"
```

#### Kafka

Generate a cluster ID:
```bash
uuidgen
```

Prepare the JAAS config:
```bash
mkdir -p kafka
cd kafka
```

```bash
cat > jaas.conf << 'EOF'
KafkaServer {
    org.apache.kafka.common.security.plain.PlainLoginModule required
    username="admin"
    password="admin-secret"
    user_wishlist="wishlist-secret";
};
EOF
```

Run Kafka:
```bash
docker volume create kafka

docker run -d \
  --name kafka \
  --restart unless-stopped \
  -p 19092:19092 \
  -p 29092:29092 \
  -v kafka:/var/lib/kafka/data \
  -v "$(pwd):/etc/kafka/jaas:ro" \
  -e KAFKA_NODE_ID=1 \
  -e KAFKA_PROCESS_ROLES=broker,controller \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:19093 \
  -e KAFKA_LISTENERS=INTERNAL://0.0.0.0:19092,EXTERNAL://0.0.0.0:29092,CONTROLLER://0.0.0.0:19093 \
  -e KAFKA_ADVERTISED_LISTENERS=INTERNAL://127.0.0.1:19092,EXTERNAL://your-domain.com:29092 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,INTERNAL:SASL_PLAINTEXT,EXTERNAL:SASL_PLAINTEXT \
  -e KAFKA_INTER_BROKER_LISTENER_NAME=INTERNAL \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  -e KAFKA_SASL_ENABLED_MECHANISMS=PLAIN \
  -e KAFKA_SASL_MECHANISM_INTER_BROKER_PROTOCOL=PLAIN \
  -e KAFKA_OPTS=-Djava.security.auth.login.config=/etc/kafka/jaas/jaas.conf \
  -e CLUSTER_ID=<uuid> \
  apache/kafka:latest
```

#### RabbitMQ

```bash
docker volume create rabbitmq

docker run -d \
  --name rabbitmq \
  --restart unless-stopped \
  -p 5672:5672 \
  -p 15672:15672 \
  -v rabbitmq:/var/lib/rabbitmq \
  -e RABBITMQ_DEFAULT_USER=username \
  -e RABBITMQ_DEFAULT_PASS=******** \
  rabbitmq:4-management
```

</details>

<details>
<summary><h3>Compose</h3></summary>

Run full stack:
```bash
CONFIG_PATH=./config.yaml docker compose up -d --build
```

Run only API and email sender with external Kafka:
```bash
CONFIG_PATH=./config.yaml \
APP_BROKER_TYPE=kafka \
APP_BROKER_KAFKA_BROKERS=10.0.0.30:9092 \
docker compose up -d --build api email-sender
```

Run only API and email sender with external RabbitMQ:
```bash
CONFIG_PATH=./config.yaml \
APP_BROKER_TYPE=rabbitmq \
APP_BROKER_RABBITMQ_URL=amqp://user:pass@10.0.0.31:5672/ \
docker compose up -d --build api email-sender
```

</details>

</details>

<details>
<summary><h2>Test</h2></summary>

Run all tests:
```bash
go test ./...
```

Run unit tests only:
```bash
go test ./internal/...
```

Run integration tests:
```bash
go test -tags=integration ./...
```

Run end-to-end tests (API must be already running):
```bash
go test -tags=e2e ./tests/e2e
```

Test a specific package:
```bash
go test ./internal/services
```

Run a specific test:
```bash
go test -run TestUserService_Register ./internal/services
```

</details>

<details>
<summary><h2>Misc</h2></summary>

<details>
<summary><h3>Swagger</h3></summary>

Generate Swagger docs:
```bash
swag init \
  -g main.go \
  -d ./cmd/api,./internal/api/controllers,./internal/models,./internal/api/errors \
  -o docs/swagger \
  --parseInternal
```

</details>

</details>

<details>
<summary><h2>TODO</h2></summary>

- [x] Add Redis caching for JWT tokens
- [x] Add email sending service for password reset and account verification
- [x] Add S3 for user avatars and other media
- [x] Add Wish and Wishlist entities to API
- [x] Add goose migrations
- [x] Add build instructions
- [x] Add Swagger documentation for API
- [x] Add unit, integration and end-to-end tests
- [x] Dockerize
- [x] Ask Claude to conjure some UI
- [x] Add message broker for emails
- [~] Set up CI/CD
- [x] Deploy
- [ ] Future
  - [ ] Add pagination for lists
  - [ ] Rate-limiting
  - [ ] Migrate from JWT in localStorage to server-side auth with httpCookie
  - [ ] Add support for HEIC
  - [ ] Allow access to lists without registration?
  - [ ] Separate config for monolith and email sender service

<details>
<summary><h3>Post-review</h3></summary>

- [ ] Findings after review
    - [ ] Fix CORS
    - [ ] Remove `password` from SELECT's and Scan's
    - [ ] Deduplicate auth params in Wish controller
    - [ ] Replace placeholders in README
    - [ ] Make hardcoded config path be read from env
    - [ ] Transactions for email verification flow
    - [ ] Fix port check race condition
    - [ ] Register returns 200 instead of 201
    - [ ] Avatar size misalign

</details>

</details>
