VERSION=$(shell git describe --tags 2>/dev/null || git rev-parse --short=7 HEAD 2>/dev/null || echo -n unknown)
DOCKER_IMAGE=w1ck3dg0ph3r/goce
DOCKER_TAG=$(DOCKER_IMAGE):latest

GO_LDFLAGS="-X main.version=$(VERSION)"

GOLANGCILINT_VERSION=v1.61.0

default: build

version:
	@echo "version: $(VERSION)"

build: build-ui build-api

build-ui:
	@cd ui &&\
	pnpm install &&\
	pnpm vite build --mode=localhost

build-api:
	go build -ldflags $(GO_LDFLAGS) .

test: test-api test-ui

test-api:
	@mkdir -p ui/dist &&\
	echo -e "User-agent: *\nDisallow: /api" > ui/dist/robots.txt &&\
	go test -v -count=1 -race ./...

test-ui:
	@cd ui &&\
	pnpm type-check

lint: lint-api lint-ui

lint-api: install-golangcilint
	golangci-lint run

lint-ui:
	@cd ui &&\
	pnpm lint

image:
	docker build \
		--build-arg version=$(VERSION) \
		-t $(DOCKER_TAG) .

image-local:
	docker build \
		--build-arg version=$(VERSION) \
		--build-arg ui_mode=localhost \
		-t $(DOCKER_TAG) .

install-golangcilint:
	@./scripts/install-golangcilint.sh $(GOLANGCILINT_VERSION)
