DOCKER_REPO=registry.blackforestbytes.com
DOCKER_NAME=mikescher/project-siren-cc

NAMESPACE=$(shell git rev-parse --abbrev-ref HEAD)
HASH=$(shell git rev-parse HEAD)

SWAGGO=github.com/swaggo/swag/cmd/swag@v1.16.3
SWAG_BIN=./_build/swagger/swaggobin/swag

GOLANGCI_LINT=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
LINTER_BIN=./_build/linter/golangci-lint/golangci-lint

GOW=github.com/mitranim/gow@latest
GOW_BIN=./_build/gow/gow

.PHONY: run build fmt test lint clean swagger docker push run-docker-local inspect-docker

run: build
	mkdir -p _run-data
	_build/server

build: generate
	mkdir -p _build
	rm -f ./_build/server
	go build -v -buildvcs=false  -o _build/server .

generate: enums ids

enums:
	@for d in ./*  ; do if [ -f "$$d/enums.go" ]; then go generate "$$d/enums.go"; fi; done
	@go generate "enums.go";

ids:
	go generate ids.go

fmt: swagger-setup
	go fmt ./...
	$(SWAG_BIN) fmt

clean:
	rm -rf _build/*
	! which go 2>&1 >> /dev/null || go clean
	! which go 2>&1 >> /dev/null || go clean -testcache

lint: lint-setup
	 $(LINTER_BIN) run ./...

lint-setup:
	mkdir -p "_build/linter/golangci-lint"
	[ -f $(LINTER_BIN) ] || { GOBIN=/tmp/_golangci-lint go install $(GOLANGCI_LINT); cp "/tmp/_golangci-lint/golangci-lint" $(LINTER_BIN); rm -rf "/tmp/_golangci-lint"; }

test:
	go generate ./...
	which gotestsum || go install gotest.tools/gotestsum@latest
	PURGE_TEST_DBS=1 gotestsum --format "testname" -- "./..."

gow: gow-setup
	$(GOW_BIN) run mikescher/project-signal-cc

gow-setup:
	mkdir -p "_build/gow"
	[ -f $(GOW_BIN) ] || { GOBIN=/tmp/_gow go install $(GOW); cp "/tmp/_gow/gow" $(GOW_BIN); rm -rf "/tmp/_gow"; }

push: docker
	docker image push "$(DOCKER_REPO)/$(DOCKER_NAME):$(HASH)"
	docker image push "$(DOCKER_REPO)/$(DOCKER_NAME):$(NAMESPACE)-latest"
	docker image push "$(DOCKER_REPO)/$(DOCKER_NAME):latest"

docker:
	docker build \
    		-t "$(DOCKER_NAME):$(HASH)" \
    		-t "$(DOCKER_NAME):$(NAMESPACE)-latest" \
    		-t "$(DOCKER_NAME):latest" \
    		-t "$(DOCKER_REPO)/$(DOCKER_NAME):$(HASH)" \
    		-t "$(DOCKER_REPO)/$(DOCKER_NAME):$(NAMESPACE)-latest" \
    		-t "$(DOCKER_REPO)/$(DOCKER_NAME):latest" \
    		.

run-docker-local: docker
	mkdir -p .run-data
	docker run --rm \
			   --init \
			   --volume "$(shell pwd)/.run-data/docker-local:/data" \
			   --publish "8000:80" \
			   $(DOCKER_NAME):latest
swagger-setup:
	@mkdir -p "_build/swagger/swaggobin"
	[ -f $(SWAG_BIN) ] || { GOBIN=/tmp/_swaggo go install $(SWAGGO); cp "/tmp/_swaggo/swag" $(SWAG_BIN); rm -rf "/tmp/_swaggo"; }

swagger: swagger-setup
	$(SWAG_BIN) init --quiet --dir .  --generalInfo ./main.go --propertyStrategy camelcase --output _build/swagger/  --outputTypes "json" --overridesFile override.swag --parseDependency
