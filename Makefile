.PHONY: proto
proto:
	mkdir -p internal/delivery/grpc/pb
	protoc --proto_path=proto \
		--go_out=internal/delivery/grpc/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/delivery/grpc/pb --go-grpc_opt=paths=source_relative \
		auth.proto

.PHONY: build
build:
	go build -o bin/auth cmd/auth/main.go

.PHONY: run
run:
	go run cmd/auth/main.go