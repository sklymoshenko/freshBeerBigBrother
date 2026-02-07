BIN_DIR := bin
BIN_NAME := bigbrother

.PHONY: build run clean test bench

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/bigbrother

run: build
	./$(BIN_DIR)/$(BIN_NAME)

test:
	go test ./...

bench:
	go test -bench . -benchmem ./internal/processor

clean:
	rm -rf $(BIN_DIR)
