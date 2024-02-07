.PHONY: build run clean

# Name of the binary
BINARY = myprogram

# Source directory
SRC_DIR = .

# Output directory
BIN_DIR = ./bin

# Database directory
DATABASE_DIR = ./database

# Compiler and linker options
GO = go
GOFLAGS = -v
LDFLAGS =

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY) $(SRC_DIR)

run: build
	@echo "Running $(BINARY)..."
	./$(BIN_DIR)/$(BINARY)

clean:
	@echo "Cleaning..."
	@rm -f $(BIN_DIR)/$(BINARY)
	@rm -rf $(DATABASE_DIR)/users
