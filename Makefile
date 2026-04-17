.PHONY: build build-encrypt test clean encrypt

APP_NAME := bootstrap
BUILD_DIR := build
ZIP_NAME := audit-agent.zip

build:
	@echo "Building Lambda binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/lambda/
	@cd $(BUILD_DIR) && zip $(ZIP_NAME) $(APP_NAME)
	@echo "Build complete: $(BUILD_DIR)/$(ZIP_NAME)"

build-encrypt:
	@echo "Building encrypt tool..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/encrypt ./cmd/encrypt/
	@echo "Build complete: $(BUILD_DIR)/encrypt"

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)

encrypt:
	@go run ./cmd/encrypt/ $(ARGS)
