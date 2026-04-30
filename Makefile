MIGRATIONS_DIR ?= backend-core/migrations
MIGRATE_DB_URL ?= postgres://user:pass@localhost:5432/devhub?sslmode=disable

.PHONY: init proto-tools proto setup migrate-tools migrate-create migrate-up migrate-down migrate-version build run

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

build:
	docker-compose build

run:
	docker-compose up
