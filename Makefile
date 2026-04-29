.PHONY: proto setup build run

proto:
	protoc --proto_path=proto --go_out=backend-core --go-grpc_out=backend-core proto/*.proto
	python3 -m grpc_tools.protoc -Iproto --python_out=backend-ai --grpc_python_out=backend-ai proto/*.proto

setup:
	cd backend-core && go mod tidy
	cd backend-ai && pip install -r requirements.txt
	cd frontend && npm install

build:
	docker-compose build

run:
	docker-compose up
