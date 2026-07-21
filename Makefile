# Makefile for gostaticstructdiff project

# Variables
BINARY_NAME := gostaticstructdiff
EXAMPLES_DIR := examples
MODELS_DIR := $(EXAMPLES_DIR)/models
GENERATED_EXAMPLE_DIFFS := \
	$(EXAMPLES_DIR)/complex_diff.go \
	$(MODELS_DIR)/user_diff.go \
	$(MODELS_DIR)/metadata_diff.go

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

# Clean generated files and binary
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(GENERATED_EXAMPLE_DIFFS)

# Generate diff files for examples
generate_example: build
	@echo "Generating diff files..."
	./$(BINARY_NAME) -input $(EXAMPLES_DIR)/complex.go -verbose
	./$(BINARY_NAME) -input $(MODELS_DIR)/user.go -verbose
	./$(BINARY_NAME) -input $(MODELS_DIR)/metadata.go -verbose
	./$(BINARY_NAME) -input $(MODELS_DIR)/nested/id.go -verbose



# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run the example program
example: generate_example
	@echo "Running example..."
	cd $(EXAMPLES_DIR)/cmd/simple && go run main.go

# Run the random generation example program
example_gen: generate_example
	@echo "Running random generation example..."
	cd $(EXAMPLES_DIR)/cmd/gen && go run main.go

# Phony targets
.PHONY: all build clean generate test example example_gen