# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=http-fuzzer
BINARY_UNIX=$(BINARY_NAME)_unix

# Build parameters
BUILD_DIR=build
MAIN_FILE=cmd/main.go

# Tool versions
GOLANGCI_LINT_VERSION=v1.55.2

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all build clean test coverage run deps lint vet fmt help install-lint

all: test build

## Build the application
build: 
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(MAIN_FILE)

## Build for unix
build-unix:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) -v $(MAIN_FILE)

## Run the application
run:
	$(GORUN) $(MAIN_FILE)

## Clean build directory
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

## Download dependencies
deps:
	$(GOGET) -v -d ./...
	$(GOMOD) tidy

## Install golangci-lint
install-lint:
	command -v golangci-lint >/dev/null 2>&1 || \
	{ \
		echo "Installing golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	}

## Run linter
lint: install-lint
	golangci-lint run

## Run go vet
vet:
	$(GOCMD) vet ./...

## Format code
fmt:
	$(GOCMD) fmt ./...

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-20s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)