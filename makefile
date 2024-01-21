BINARY_NAME := ntop
VERSION := 0.2.0

.PHONY: all windows linux mac

all: windows linux mac

windows:
	@echo "Building for Windows (x64)..."
	GOOS=windows GOARCH=amd64 go build -o release/${BINARY_NAME}-${VERSION}-windows-amd64.exe

linux:
	@echo "Building for Linux (x64)..."
	GOOS=linux GOARCH=amd64 go build -o release/${BINARY_NAME}-${VERSION}-linux-amd64

mac:
	@echo "Building for MacOS (Intel)..."
	GOOS=darwin GOARCH=amd64 go build -o release/${BINARY_NAME}-${VERSION}-darwin-amd64
	@echo "Building for MacOS (ARM64)..."
	GOOS=darwin GOARCH=arm64 go build -o release/${BINARY_NAME}-${VERSION}-darwin-arm64
