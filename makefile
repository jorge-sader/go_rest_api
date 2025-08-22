# Create SSL certificate
.PHONY: new-ssl-cert
new-ssl-cert:
	openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -config openssl.cnf

# Run all tests
.PHONY: test
test: ## Run all tests with race detection
	@go test -v -race ./...

# Display test coverage
.PHONY: coverage
coverage: ## Display test coverage
	@go test -cover -race ./...

# Generate and view coverage report
.PHONY: cover
cover: ## Generate and view coverage report
	@go test -coverprofile=coverage.out -tags integration ./... && go tool cover -html=coverage.out || rm -f coverage.out

# Install the CLI to a custom destination
.PHONY: install
install: build ## Install CLI to DEST_DIR
	@if [ -n "$(DEST_DIR)" ]; then cp $(BUILD_DIR)/$(BINARY_NAME) $(DEST_DIR)/$(BINARY_NAME); echo "Installed to $(DEST_DIR)"; else echo "DEST_DIR not set, skipping install"; fi

# Stage all files for Git commit
.PHONY: stage-all
stage-all: ## Stage all files for Git
	@echo "Staging all files..."
	@git add .
	@echo "All files staged!"

# Unstage all files
.PHONY: unstage-all
unstage-all: ## Unstage all files
	@echo "Unstaging all files..."
	@if git rev-parse HEAD >/dev/null 2>&1; then git restore --staged .; else echo "No commits yet, nothing to unstage"; fi
	@echo "All files unstaged!"

# Copy staged diff to clipboard
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

# Generate diff and copy to clipboard
.PHONY: diff
diff: stage-all diff-to-clipboard unstage-all ## Generate and copy diff to clipboard
	@echo "Diff content on clipboard."

# Save diff to file
.PHONY: diff-file
diff-file: stage-all ## Save diff to diff_output.txt
	@echo "Saving diff to diff_output.txt..."
	@git diff --staged > diff_output.txt
	@echo "Diff saved to diff_output.txt."
	@$(MAKE) unstage-all

# Show available targets
.PHONY: help
help: ## Show available make targets
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'