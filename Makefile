## variables
GOCMD?=go
CODE_COVERAGE_FILE:= coverage
CODE_COVERAGE_FILE_TXT := $(CODE_COVERAGE_FILE).txt
PACKAGE_BASE=github.com/AidanDelaney/scafall
SRC=$(shell find . -type f -name '*.go' -not -path "*/testdata/*")

all: build verify test

build:
	go build -o scafall main.go

test: lint test-unit test-integration test-system

install-golangci-lint:
	@echo "> Installing golangci-lint..."
	cd tools && $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint

install-go-acc:
	@echo "	installing go-acc"
	cd tools && $(GOCMD) install github.com/ory/go-acc

test-unit: install-go-acc
	@echo "	running unit tests"
	go-acc ./pkg/... -o $(CODE_COVERAGE_FILE_TXT)

test-integration:
	go test ./test_integration/ -count=1

test-system:
	go test ./test_system/ -count=1

install-goimports:
	@echo "> Installing goimports..."
	cd tools && $(GOCMD) install golang.org/x/tools/cmd/goimports

verify-format: install-goimports
	@echo "> Formating code..."
	@goimports -l -local ${PACKAGE_BASE} ${SRC}

format: install-goimports
	@echo "> Formating code..."
	@goimports -l -w -local ${PACKAGE_BASE} ${SRC}

lint: install-golangci-lint
	golangci-lint run

verify: lint verify-format