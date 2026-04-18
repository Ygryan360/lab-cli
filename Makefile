.PHONY: build install clean run

BINARY := lab
INSTALL_DIR := $(HOME)/.local/bin

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/lab/

install: build
	mkdir -p $(INSTALL_DIR)
	mv $(BINARY) $(INSTALL_DIR)/$(BINARY)
	chmod +x $(INSTALL_DIR)/$(BINARY)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY)"

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

# Cross-compile for Linux ARM (e.g. Raspberry Pi)
build-arm:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY)-arm64 ./cmd/lab/
