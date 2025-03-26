
# Default target
all: run

# cleaning
clean:
	@echo "Cleaning protogen files..."
	rm -fR ./protogen 
	mkdir -p ./protogen

# Build
build:
	@echo "Build proto..."
	@protoc ./proto/token/health/*.proto --go_out=. --go-grpc_out=.
	@protoc ./proto/token/payment/*.proto --go_out=. --go-grpc_out=.
	@protoc ./proto/token/pod/*.proto --go_out=. --go-grpc_out=.
	@protoc ./proto/token/card/*.proto --go_out=. --go-grpc_out=.
	@protoc ./proto/token/*.proto --go_out=. --go-grpc_out=.

# run
run:
	@echo "Run..."
	@go run ./cmd/main.go

.PHONY: all build run