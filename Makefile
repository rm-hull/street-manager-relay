# Variables
BINARY_NAME = street-manager-relay
PACKAGE_NAME = generated
GENERATED_DIR = generated

# Schema definitions: URL|output_filename|go_filename
SCHEMAS = \
	https://department-for-transport-streetmanager.github.io/street-manager-docs/api-documentation/json/event-notifier-message.json|event-notifier-message.json|event_notifier_message.go

# 	https://department-for-transport-streetmanager.github.io/street-manager-docs/api-documentation/json/api-notification-event-notifier-message.json|api-notification-event.json|api-notification_event.go \

# Extract components from schema definitions
JSON_FILES = $(foreach schema,$(SCHEMAS),$(GENERATED_DIR)/$(word 2,$(subst |, ,$(schema))))
GO_BINDINGS = $(foreach schema,$(SCHEMAS),$(GENERATED_DIR)/$(word 3,$(subst |, ,$(schema))))

# Default target
all: build

# Ensure generated directory exists
$(GENERATED_DIR):
	mkdir -p $(GENERATED_DIR)

# Download all JSON schema files
download: $(JSON_FILES)

# Download individual JSON files
$(GENERATED_DIR)/api-notification-event.json: | $(GENERATED_DIR)
	@echo "Downloading api-notification-event.json..."
	curl -L -o $@ "https://department-for-transport-streetmanager.github.io/street-manager-docs/api-documentation/json/api-notification-event-notifier-message.json"

$(GENERATED_DIR)/event-notifier-message.json: | $(GENERATED_DIR)
	@echo "Downloading event-notifier-message.json..."
	curl -L -o $@ "https://department-for-transport-streetmanager.github.io/street-manager-docs/api-documentation/json/event-notifier-message.json"

# Generate Go bindings from JSON schema
generate: $(GO_BINDINGS)

# Generate individual Go bindings
$(GENERATED_DIR)/api-notification_event.go: $(GENERATED_DIR)/api-notification-event.json
	@echo "Generating Go bindings for api-notification_event.go..."
	npx quicktype --src-lang schema $< --out $@ --lang go --package $(PACKAGE_NAME)

$(GENERATED_DIR)/event_notifier_message.go: $(GENERATED_DIR)/event-notifier-message.json
	@echo "Generating Go bindings for event_notifier_message.go..."
	npx quicktype --src-lang schema $< --out $@ --lang go --package $(PACKAGE_NAME)

# Build the Go binary
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_BINDINGS)
	@echo "Building Go binary..."
	go build -tags=jsoniter -ldflags="-w -s" -o $(BINARY_NAME) .

# Run the application
run: $(GO_BINDINGS)
	@echo "Running application..."
	go run ./...

# Test target (depends on generated bindings)
test: $(GO_BINDINGS)
	@echo "Running tests..."
	gotestsum --junitfile=./test-reports/junit.xml --format github-actions -- -v -coverprofile=profile.cov -coverpkg=./... ./...


# Clean generated files
clean: clean-bindings clean-json
	rm -f $(BINARY_NAME)

# Install dependencies (quicktype)
deps:
	@echo "Installing dependencies..."
	npm install -g quicktype

# Force regeneration of bindings
regen: clean-bindings generate

# Clean only bindings (keeps downloaded JSON)
clean-bindings:
	rm -f $(GENERATED_DIR)/*.go

# Force redownload
redownload: clean-json download

# Clean only JSON files
clean-json:
	rm -f $(GENERATED_DIR)/*.json

# Debug: show what files we expect
debug:
	@echo "JSON_FILES: $(JSON_FILES)"
	@echo "GO_BINDINGS: $(GO_BINDINGS)"

# Show help
help:
	@echo "Available targets:"
	@echo "  all       - Build the application (default)"
	@echo "  download  - Download the JSON schema files"
	@echo "  generate  - Generate Go bindings from JSON schema"
	@echo "  build     - Build the Go binary"
	@echo "  run       - Run the built application"
	@echo "  test      - Run Go tests"
	@echo "  clean     - Remove all generated files"
	@echo "  deps      - Install required dependencies"
	@echo "  regen     - Regenerate bindings (clean + generate)"
	@echo "  redownload- Force redownload of JSON files"
	@echo "  debug     - Show expected file paths"
	@echo "  help      - Show this help message"

# Declare phony targets
.PHONY: all download generate build run test clean deps regen clean-bindings redownload clean-json debug help