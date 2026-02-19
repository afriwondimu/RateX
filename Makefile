.PHONY: build test test-short test-integration clean run-example deps lint fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=ratex

all: deps build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./...

test: deps
	$(GOTEST) -v ./...

test-short: deps
	$(GOTEST) -v -short ./...

test-integration: deps
	$(GOTEST) -v -tags=integration ./...

test-coverage: deps
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out

deps:
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOGET) github.com/stretchr/testify

run-example-basic:
	cd examples/basic && $(GOCMD) run main.go

run-example-redis:
	cd examples/redis && $(GOCMD) run main.go

run-example-custom:
	cd examples/custom && $(GOCMD) run main.go

lint:
	golangci-lint run

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...