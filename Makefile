SHELL = /usr/bin/env bash -o nounset -o pipefail -o errexit -c

ifneq ("$(wildcard .env)","")
	include .env
	export
endif

APP_NAME = ketchup
NOTIFIER_NAME = notifier
PACKAGES ?= ./...

MAIN_SOURCE = cmd/ketchup/api.go
MAIN_RUNNER = go run $(MAIN_SOURCE)
ifeq ($(DEBUG), true)
	MAIN_RUNNER = dlv debug $(MAIN_SOURCE) --
endif

NOTIFIER_SOURCE = cmd/notifier/notifier.go
NOTIFIER_RUNNER = go run $(NOTIFIER_SOURCE)
ifeq ($(DEBUG), true)
	NOTIFIER_RUNNER = dlv debug $(NOTIFIER_SOURCE)) --
endif

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sort

## name: Output app name
.PHONY: name
name:
	@printf "$(APP_NAME)"

## version: Output last commit sha
.PHONY: version
version:
	@printf "$(shell git rev-parse --short HEAD)"

## version-date: Output last commit date
.PHONY: version-date
version-date:
	@printf "$(shell git log -n 1 "--date=format:%Y%m%d%H%M" "--pretty=format:%cd")"

## app: Build whole app
.PHONY: app
app: init dev

## dev: Build app
.PHONY: dev
dev: format style test build

## init: Bootstrap your application. e.g. fetch some data files, make some API calls, request user input etc...
.PHONY: init
init:
	@curl --disable --silent --show-error --location --max-time 30 "https://raw.githubusercontent.com/ViBiOh/scripts/main/bootstrap" | bash -s -- "-c" "git_hooks" "coverage" "release"
	go install "github.com/golang/mock/mockgen@v1.6.0"
	go install "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	go install "golang.org/x/tools/cmd/goimports@latest"
	go install "mvdan.cc/gofumpt@latest"
	go mod tidy

## format: Format code. e.g Prettier (js), format (golang)
.PHONY: format
format:
	goimports -w $(shell find . -name "*.go")
	gofumpt -w $(shell find . -name "*.go") 2>/dev/null

## style: Check lint, code styling rules. e.g. pylint, phpcs, eslint, style (java) etc ...
.PHONY: style
style:
	golangci-lint run

## mocks: Generate mocks
.PHONY: mocks
mocks:
	find . -name "mocks" -type d -exec rm -r "{}" \+
	go generate -run mockgen $(PACKAGES)
	mockgen -destination pkg/mocks/pgx.go -package mocks -mock_names Row=Row,Rows=Rows github.com/jackc/pgx/v4 Row,Rows

## test: Shortcut to launch all the test tasks (unit, functional and integration).
.PHONY: test
test:
	scripts/coverage
	$(MAKE) bench

## bench: Shortcut to launch benchmark tests.
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build the application.
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/$(APP_NAME) $(MAIN_SOURCE)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/$(NOTIFIER_NAME) $(NOTIFIER_SOURCE)

## run: Locally run the application, e.g. node index.js, python -m myapp, go run myapp etc ...
.PHONY: run
run:
	$(MAIN_RUNNER)

## run-notifier: Locally run the notifier of the application
.PHONY: run-notifier
run-notifier:
	$(NOTIFIER_RUNNER)

.PHONY: sidecars
sidecars:
	docker run --detach --name "ketchup-pg" --publish "127.0.0.1:$(KETCHUP_DB_PORT):5432" --env "POSTGRES_USER=$(KETCHUP_DB_USER)" --env "POSTGRES_PASSWORD=$(KETCHUP_DB_PASS)" --env "POSTGRES_DB=$(KETCHUP_DB_NAME)" "postgres"

.PHONY: sidecars-off
sidecars-off:
	docker rm --force --volumes "ketchup-pg"
