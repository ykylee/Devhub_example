.PHONY: init proto-tools proto setup build run

init: setup proto-tools proto

proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

proto:
	protoc --proto_path=proto --go_out=backend-core --go-grpc_out=backend-core proto/*.proto
	python3 -m grpc_tools.protoc -Iproto --python_out=backend-ai --grpc_python_out=backend-ai proto/*.proto

setup:
	cd backend-core && go mod tidy
	cd backend-ai && python3 -m pip install -r requirements.txt
	cd frontend && npm install

build:
	docker-compose build

run:
	docker-compose up
