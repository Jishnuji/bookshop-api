.PHONY: dc run test lint install-lint-dev

dc:
	docker-compose up  --remove-orphans --build

build:
	go build -race -o app cmd/main.go

run:
	go build -o app cmd/main.go && \
	HTTP_ADDR=:8080 \
	DEBUG_ERRORS=1 \
	DSN="postgres://postgres:@127.0.0.1:5432/bookshop?sslmode=disable" \
	MIGRATIONS_PATH="file://./internal/app/migrations" \
	./app

test:
	go test -race ./internal/app/services
	go test -race ./internal/app/domain
	go test -race ./internal/app/transport/httpserver/httpserver_test

# Installation dev version golangci-lint to Go 1.25 support
install-lint-dev:
	@echo "Installing golangci-lint from source for Go 1.25 support..."
	@if [ -d "/tmp/golangci-lint" ]; then \
		cd /tmp/golangci-lint && git pull; \
	else \
		git clone https://github.com/golangci/golangci-lint.git /tmp/golangci-lint; \
	fi
	@cd /tmp/golangci-lint && go install ./cmd/golangci-lint

# Fallback to stable latest version
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	golangci-lint run ./...

generate:
	go generate ./...


