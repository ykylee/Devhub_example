MIGRATIONS_DIR ?= backend-core/migrations
MIGRATE_DB_URL ?= postgres://user:pass@localhost:5432/devhub?sslmode=disable

# Frontend build-time env (NEXT_PUBLIC_* 변수는 client bundle 에 inline — build 시점 결정)
NEXT_PUBLIC_BASE_PATH ?= devhub
BACKEND_API_URL ?= http://localhost:8080

# Docker compose file for image build (override with -f flag for other files)
COMPOSE_FILE ?= docker-compose.yml

.PHONY: init proto-tools proto setup migrate-tools migrate-create migrate-up migrate-down migrate-version build build-backend build-frontend build-docker run test test-race test-coverage test-frontend e2e lint-migrations

init: setup proto-tools migrate-tools proto

proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

migrate-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

proto:
	protoc --proto_path=proto --go_out=backend-core --go-grpc_out=backend-core proto/*.proto
	python3 -m grpc_tools.protoc -Iproto --python_out=backend-ai --grpc_python_out=backend-ai proto/*.proto

setup:
	cd backend-core && go mod tidy
	cd backend-ai && python3 -m pip install -r requirements.txt
	cd frontend && npm install

migrate-create:
	@test -n "$(NAME)" || (echo "usage: make migrate-create NAME=create_webhook_events" && exit 1)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DB_URL)" down 1

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DB_URL)" version

# ----------------------------------------------------------------------------
# Build targets
# - `build`             : 정적 binary + frontend standalone + docker image 전체 build
# - `build-backend`     : backend Go binary 만 (정적 link, CGO_ENABLED=0)
# - `build-frontend`    : frontend Next.js standalone build 만
# - `build-docker`      : docker image 만 (prebuilt bin/main + .next/standalone 사용)
# - 주의: Dockerfile 은 `COPY bin/main ./main` 이므로 backend build 가 선행되어야 함.
#   docker build 가 orphan bin/main 을 stale state 로 사용하지 않도록 build 가 build-docker 에 의존.
# ----------------------------------------------------------------------------

build: build-backend build-frontend build-docker

build-backend:
	@echo "==> building backend-core (CGO_ENABLED=0, static, alpine-compatible)"
	cd backend-core && CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o bin/main .
	@echo "==> bin/main built (statically linked, size: $$(stat -c%s backend-core/bin/main 2>/dev/null || stat -f%z backend-core/bin/main) bytes)"
	@file backend-core/bin/main | head -1

build-frontend:
	@echo "==> building frontend (NEXT_OUTPUT=standalone, NEXT_PUBLIC_BASE_PATH=$(NEXT_PUBLIC_BASE_PATH))"
	cd frontend && NEXT_OUTPUT=standalone NEXT_PUBLIC_BASE_PATH=$(NEXT_PUBLIC_BASE_PATH) BACKEND_API_URL=$(BACKEND_API_URL) npm run build
	@echo "==> frontend standalone build done (.next/standalone/ + .next/static/)"

build-docker:
	@echo "==> building docker images (using prebuilt bin/main + .next/standalone)"
	@if [ ! -f backend-core/bin/main ]; then \
		echo "ERROR: backend-core/bin/main not found. Run 'make build-backend' first."; \
		exit 1; \
	fi
	@if [ ! -d frontend/.next/standalone ]; then \
		echo "ERROR: frontend/.next/standalone not found. Run 'make build-frontend' first."; \
		exit 1; \
	fi
	docker compose -f $(COMPOSE_FILE) build
	@echo "==> docker images built (use 'docker compose -f $(COMPOSE_FILE) up -d' to start)"

run:
	@echo "Run is environment-specific. See docs/setup/environment-setup.md."
	@echo "  Native:  see section 2 of the guide (go run ./backend-core, python backend-ai/main.py, npm run dev)"
	@echo "  Docker:  docker-compose up      (requires local, untracked docker-compose.yml)"

# ----------------------------------------------------------------------------
# Test targets
# - `test`         : backend Go `go test ./...`
# - `test-race`    : same with -race + -count=1 to expose data races
# - `test-coverage`: writes coverage.out + prints package-level summary
# - `test-frontend`: frontend Vitest unit suite
# - `e2e`          : frontend Playwright e2e suite
# ----------------------------------------------------------------------------

test:
	cd backend-core && go test ./...

test-race:
	cd backend-core && go test -race -count=1 ./...

test-coverage:
	cd backend-core && go test -coverprofile=coverage.out ./...
	cd backend-core && go tool cover -func=coverage.out | tail -20

test-frontend:
	cd frontend && npm run test

e2e:
	cd frontend && npm run e2e

lint-migrations:
	@bash ./scripts/check-migration-uniqueness.sh
