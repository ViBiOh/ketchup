SHELL = /bin/bash

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
	MAIN_RUNNER = gdlv -d $(shell dirname $(MAIN_SOURCE)) debug --
endif

NOTIFIER_SOURCE = cmd/notifier/notifier.go
NOTIFIER_RUNNER = go run $(NOTIFIER_SOURCE)
ifeq ($(DEBUG), true)
	NOTIFIER_RUNNER = gdlv -d $(shell dirname $(NOTIFIER_SOURCE)) debug --
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

## version: Output last commit sha1
.PHONY: version
version:
	@printf "$(shell git rev-parse --short HEAD)"

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
	go install github.com/kisielk/errcheck@latest
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang/mock/mockgen@v1.6.0
	$(MAKE) mocks
	go mod tidy

## format: Format code. e.g Prettier (js), format (golang)
.PHONY: format
format:
	goimports -w $(shell find . -name "*.go")
	gofmt -s -w $(shell find . -name "*.go")

## style: Check lint, code styling rules. e.g. pylint, phpcs, eslint, style (java) etc ...
.PHONY: style
style:
	golint $(PACKAGES)
	errcheck -ignoretests $(PACKAGES)
	go vet $(PACKAGES)

## mocks: Generate mocks
.PHONY: mocks
mocks:
	find . -name "mocks" -type d -exec rm -r "{}" \+
	mockgen -destination pkg/mocks/user_service.go -mock_names UserService=UserService -package mocks github.com/ViBiOh/ketchup/pkg/middleware UserService
	mockgen -destination pkg/mocks/mailer.go -mock_names Mailer=Mailer -package mocks github.com/ViBiOh/ketchup/pkg/notifier Mailer
	mockgen -destination pkg/mocks/auth.go -mock_names AuthService=Auth -package mocks github.com/ViBiOh/ketchup/pkg/service/user AuthService
	mockgen -destination pkg/mocks/user_store.go -mock_names Store=UserStore -package mocks github.com/ViBiOh/ketchup/pkg/service/user Store

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
