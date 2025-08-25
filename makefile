# Makefile for building, testing, and running a Go application.
# Prerequisites: Go, podman-compose or docker-compose, git, golangci-lint (for linting).
# Use `make help` to see available targets.
# Comments with `##` are shown in the help output.

BUILD_DIR := bin
BINARY_NAME := myapp
TEMP_DIR := tmp/$(BINARY_NAME)
MAIN_FILE := cmd/api/server.go
SSL_CONFIG_FILE := openssl.cnf
DIFF_OUTPUT_FILE := $(TEMP_DIR)/diff_output.txt
TEST_FLAG ?= -v

# Default to podman, but fallback to docker.
COMPOSE_CMD := $(shell command -v podman-compose >/dev/null 2>&1 && echo "podman-compose" || (command -v docker-compose >/dev/null 2>&1 && echo "docker-compose" || echo ""))

#========================
# ---- Build Targets ----
#========================

.PHONY: build
build: check-go ## Build binary app
	@mkdir -p $(BUILD_DIR)
	@echo "Building the binary..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

.PHONY: run
run: check-go ## Run server
	@echo "Starting the server..."
	@go run $(MAIN_FILE)

.PHONY: run-cli
run-cli: check-go build ## Run CLI app
	@echo "Starting the CLI app..."
	@$(BUILD_DIR)/$(BINARY_NAME) &

.PHONY: start
start: compose-up run ## Start database containers and server (web)

.PHONY: start-cli
start-cli: compose-up run-cli ## Start database containers and CLI app

.PHONY: stop-cli
stop-cli: ## Stop the CLI app
	@echo "Stopping $(BINARY_NAME)..."
	@-pkill -SIGTERM -f "$(BUILD_DIR)/$(BINARY_NAME)" || true
	@echo "Stopped $(BINARY_NAME)!"

.PHONY: restart
restart: compose-down clean compose-up run ## Restart the web app (stop containers, clean, start containers, run)

.PHONY: restart-cli
restart-cli: stop-cli compose-down clean compose-up run ## Restart the CLI app (stop process, stop containers, clean, start containers, run in background)


#============================
# ---- Container Targets ----
#============================

.PHONY: compose-up
compose-up: ## Start database containers
	@if [ -z "$(COMPOSE_CMD)" ]; then echo "Error: neither podman-compose nor docker-compose installed"; exit 1; fi
	@echo "Starting DB Containers with $(COMPOSE_CMD)..."
	@$(COMPOSE_CMD) up -d || { echo "Failed to start containers!"; exit 1; }
	@echo "DB Containers Started!"

.PHONY: compose-down
compose-down: ## Stop database containers
	@if [ -z "$(COMPOSE_CMD)" ]; then echo "Error: neither podman-compose nor docker-compose installed"; exit 1; fi
	@echo "Stopping DB Containers..."
	@$(COMPOSE_CMD) down || { echo "Failed to stop containers!"; exit 1; }
	@echo "DB Containers Stopped!"


#===============================
# ---- Code Quality Targets ----
#===============================

.PHONY: clean
clean: ## Remove build artifacts, coverage files, and diff files
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR) $(TEMP_DIR)/coverage.out $(TEMP_DIR)/diff_output.txt
	@echo "Cleanup complete!"

.PHONY: fmt
fmt: check-go ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: lint
lint: check-go ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Error: golangci-lint not installed"; exit 1; }
	@echo "Running linter..."
	@golangci-lint run

.PHONY: pretty
pretty: fmt lint ## Format Go code and run linter


#==========================
# ---- Testing Targets ---- 
#==========================

.PHONY: test
test: check-go ## Run all tests with race detection (use TEST_FLAG=-v for verbose output)
	@go test $(TEST_FLAG) -race ./...

.PHONY: coverage
coverage: check-go ## Display test coverage (use TEST_FLAG=-v for verbose output)
	@go test -cover $(TEST_FLAG) -race ./...

.PHONY: cover
cover: check-go ## Generate and view coverage report (run 'make clean' to remove coverage files)
	@mkdir -p $(TEMP_DIR)
	@go test -coverprofile=$(TEMP_DIR)/coverage.out -tags integration ./... && go tool cover -html=$(TEMP_DIR)/coverage.out || rm -f $(TEMP_DIR)/coverage.out

#==========================
# ---- Utility Targets ---- 
#==========================

.PHONY: new-ssl-cert
new-ssl-cert: ## Create new SSL certificate from $(SSL_CONFIG_FILE)
	@test -f $(SSL_CONFIG_FILE) || { echo "Error: $(SSL_CONFIG_FILE) not found"; exit 1; }
	@openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -config $(SSL_CONFIG_FILE)

.PHONY: stage-all
stage-all: ## Stage all files for Git
	@echo "Staging all files..."
	@git add .
	@echo "All files staged!"

.PHONY: unstage-all
unstage-all: ## Unstage all files
	@echo "Unstaging all files..."
	@if git rev-parse HEAD >/dev/null 2>&1; then git restore --staged .; else echo "No commits yet, nothing to unstage"; fi
	@echo "All files unstaged!"

.PHONY: diff-to-clipboard
diff-to-clipboard: ## Copy staged diff to clipboard
	@echo "Copying diff to clipboard..."
	@if [ "$$(uname)" = "Darwin" ]; then \
	    git diff --staged | pbcopy && echo "Diff copied (macOS)"; \
	  elif [ "$$(uname)" = "Linux" ]; then \
	    if command -v xclip >/dev/null; then \
	      git diff --staged | xclip -selection clipboard && echo "Diff copied (Linux/xclip)"; \
	    elif command -v wl-copy >/dev/null; then \
	      git diff --staged | wl-copy && echo "Diff copied (Linux/wl-copy)"; \
	    else \
	      echo "Warning: xclip or wl-copy required for Linux"; exit 1; \
	    fi; \
	  elif [ "$$(uname -o 2>/dev/null)" = "Msys" ] || [ "$$(uname -o 2>/dev/null)" = "Cygwin" ]; then \
	    git diff --staged | clip && echo "Diff copied (Windows)"; \
	  else \
	    echo "Warning: Unsupported OS for clipboard copy"; exit 1; \
	  fi

.PHONY: diff
diff: stage-all diff-to-clipboard unstage-all ## Generate and copy diff to clipboard
	@echo "Diff content on clipboard."

.PHONY: diff-file
diff-file: stage-all ## Save diff to $(DIFF_OUTPUT_FILE)
	@mkdir -p $(TEMP_DIR)
	@echo "Saving diff to $(DIFF_OUTPUT_FILE)..."
	@git diff --staged > $(DIFF_OUTPUT_FILE)
	@echo "Diff saved to $(DIFF_OUTPUT_FILE)."
	@$(MAKE) unstage-all


#=============================
# ---- Validation Targets ---- 
#=============================

.PHONY: check-go ## Checks that Go is installed. Throws error if not
check-go:
	@command -v go >/dev/null 2>&1 || { echo "Error: Go is not installed"; exit 1; }


#========================
# ---- CI/CD Targets ----
#========================

.PHONY: ci
ci: fmt lint test build ## Run all checks for CI/CD
	@echo "CI checks completed!"


#===============
# ---- Help ---- 
#===============

# .PHONY: help
# help: ## Show available make targets
# 	@echo "Available targets:"
# 	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# .PHONY: help
# help: ## Show available make targets
# 	@echo "Available targets (grouped by category):"
# 	@echo "Shared Targets (web and CLI):"
# 	@grep -E '^(build|test|coverage|cover|clean|fmt|lint|pretty|ci|check-go|new-ssl-cert|stage-all|unstage-all|diff|diff-file):.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
# 	@echo "Web App Targets:"
# 	@grep -E '^(run|start|restart):.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
# 	@echo "CLI App Targets:"
# 	@grep -E '^.*-cli:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
# 	@echo "Container Targets:"
# 	@grep -E '^(compose-up|compose-down):.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: help
help: ## Show available make targets
	@echo "========================================"
	@echo "=---------------- HELP ----------------="
	@echo "========================================"
	@echo ""
	@echo "Available targets (grouped by category):"
	
	@echo ""
	@echo "Shared Targets (web and CLI):"
	@grep -E '^(build|test|coverage|cover|clean|fmt|lint|pretty|ci|check-go|new-ssl-cert|stage-all|unstage-all|diff|diff-file):.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-15s\033[0m %s\n", $$1, $$2}'

	@echo ""
	@echo "Web App Targets:"
	@grep -E '^(run|start|restart):.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-15s\033[0m %s\n", $$1, $$2}'

	@echo ""
	@echo "CLI App Targets:"
	@grep -E '^[a-zA-Z_-]+-cli:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-15s\033[0m %s\n", $$1, $$2}'

	@echo ""
	@echo "Container Targets:"
	@grep -E '^(compose-up|compose-down):.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-15s\033[0m %s\n", $$1, $$2}'