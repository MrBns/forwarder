set positional-arguments := true
set dotenv-load := true

BIN_NAME := "forwarder.bin"
BIN_DIR := "."

# Run with live reload (requires `air` installed).
run:
    air

# Backward-compatible alias for local development.
dev:
    just run

# Build an optimized production binary.
build:
    mkdir -p {{ BIN_DIR }}
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o {{ BIN_DIR }}/{{ BIN_NAME }} .

# start the server from build binary
start:
    {{ BIN_DIR }}/{{ BIN_NAME }}
