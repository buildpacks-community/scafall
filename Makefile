## variables
BIN_DIR := $(GOPATH)/bin
GO_ACC := $(BIN_DIR)/go-acc@latest
CODE_COVERAGE_FILE:= coverage
CODE_COVERAGE_FILE_TXT := $(CODE_COVERAGE_FILE).txt

all: build test

build:
	go build -o scafall main.go

test: test-clean test-unit test-integration test-system

test-clean:
	@echo "	cleaning test cache"
	go clean -testcache ./...

$(GO_ACC):
	@echo "	installing testing tools"
	which go-acc || go install -v github.com/ory/go-acc@latest
	$(eval export PATH=$(GO_ACC):$(PATH))

test-unit: $(GO_ACC) test-clean
	@echo "	running unit tests"
	go-acc ./pkg/... -o $(CODE_COVERAGE_FILE_TXT)

test-integration: test-clean
	go test ./test_integration/

test-system: test-clean
	go test ./test_system/