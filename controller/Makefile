# go option
GO             ?= go
PKG            := go mod vendor
LDFLAGS        := -w -s
GOFLAGS        :=
GOPRIVATE      := 
GOPROXY        := https://proxy.golang.org
GO_EXTRA_FLAGS := -v
TAGS           :=
BINDIR         := $(CURDIR)/bin
BIN_NAME       := ck8sd
PKGDIR         := github.com/cswarm/$(BIN_NAME)
CGO_ENABLED    := 1

.PHONY: all
all: build

.PHONY: pre-commit
pre-commit: fmt

.PHONY: build
build: gogenerate gobuild

.PHONY: debug
debug:
	KUBECONFIG=../kubeconfig.yaml WATCH_NAMESPACE=default OPERATOR_NAME=ck8sd ./bin/ck8sd

.PHONY: test
test:
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) ./scripts/test.sh

.PHONY: docs
docs:
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) ./scripts/docs.sh

.PHONY: dep
dep:
	@echo " ===> Installing dependencies via <=== "
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) $(PKG)

.PHONY: gogenerate
gogenerate:
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) go generate ./...

.PHONY: gobuild
gobuild:
	@echo " ===> building releases in ./bin/... <=== "
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BINDIR)/$(BIN_NAME) -ldflags "$(LDFLAGS)" $(GO_EXTRA_FLAGS) $(PKGDIR)/cmd/manager

.PHONY: docker-build
docker-build:
	@echo " ===> building docker image <==="
	@ssh-add -L
	@echo " ===> If you run into credential issues, ensure that your key is in your SSH agent (ssh-add <ssh-key-path>) <==="
	DOCKER_BUILDKIT=1 docker build --ssh default -t codegen -f deployments/codegen/Dockerfile . --build-arg VERSION=${VERSION} 

.PHONY: fmt
fmt:
	@echo " ===> Running goimports <==="
	find  . -path ./vendor -prune -o -type f -name '*.go' -print | xargs -n 1 goimports -w
