BINARY_NAME=gohl
BUILD_DIR=dist
MAIN_PATH=cmd/gohl/main.go

all: build

build:
	@echo "Building for local OS..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH) scan

release:
	@echo "Compiling for Linux (AMD64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	@echo "Compiling for Raspberry Pi (ARM64)..."
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	@echo "Compiling for Windows..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_PATH)
	
	@echo "Done! Binaries are in $(BUILD_DIR)/"

clean:
	rm -rf $(BUILD_DIR)
	rm -rf ~/.gohl/history/*.json # Om du vill rensa historik vid clean