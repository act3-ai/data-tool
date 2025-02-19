
all: build

REV=$(shell git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD)

# This is the default. It can be overridden in the main Makefile after
# including build.make.
REGISTRY_NAME=zot.lion.act3-ace.ai

# Images are named after the command contained in them.
IMAGE_REPO=$(REGISTRY_NAME)/ace/data/tool

# Tool versions

# renovate: datasource=go depName=sigs.k8s.io/controller-tools
CONTROLLER_GEN_VERSION?=v0.17.1
# renovate: datasource=go depName=github.com/elastic/crd-ref-docs
CRD_REF_DOCS_VERSION?=v0.1.0
# renovate: datasource=go depName=github.com/google/ko
KO_VERSION?=v0.17.1
# renovate: datasource=go depName=github.com/golangci/golangci-lint
GOLANGCI_LINT_VERSION?=v1.64.5

REGISTRY_CONTAINER ?= data-test-registry
TELEMETRY_CONTAINER ?= data-test-telemetry

CONTAINER_RUNTIME?=podman
# CONTAINER_RUN_ARGS=--network=slirp4netns # needed on podman on linux from homebrew right now

.PHONY: generate
generate: tool/controller-gen
	go generate ./...

.PHONY: build
build: dt

.PHONY: dt
dt: generate
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/ace-dt ./cmd/ace-dt

.PHONY: dt
dt-linux: generate
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o ci-dist/ace-dt/linux/amd64/bin/ace-dt ./cmd/ace-dt

.PHONY: clean
clean:
	- rm -rf bin

.PHONY: template
template:
	- rm -rf internal/mirror/testing/testdata/large/oci
	go run ./cmd/ace-dt run-recipe internal/mirror/testing/testdata/large/recipe.jsonl

	- rm -rf internal/mirror/testing/testdata/small/oci
	go run ./cmd/ace-dt run-recipe internal/mirror/testing/testdata/small/recipe.jsonl

.PHONY: test
test:

.PHONY: test-go
test: test-go
test-go:
	go test -short ./...
	
.PHONY: lint
test: lint
lint: tool/golangci-lint
	tool/golangci-lint run
	- go run github.com/nikolaydubina/smrcptr@latest ./...

.PHONY: test-functional
test-functional: start-services

	$(eval REGISTRY_HOST := $(shell $(CONTAINER_RUNTIME) port $(REGISTRY_CONTAINER) 5000/tcp))
	$(eval TELEMETRY_HOST := $(shell $(CONTAINER_RUNTIME) port $(TELEMETRY_CONTAINER) 8100/tcp))
	# We disable caching since the tests depend on external data not monitored by "go test" with -count=1
	TEST_REGISTRY=$(REGISTRY_HOST) TEST_TELEMETRY=http://$(TELEMETRY_HOST) go test -count=1 ./...
	
	# $(MAKE) stop-services

.PHONY: cover
cover:
	go clean -testcache
	- rm coverage.txt
	go test -count=1 ./... -coverprofile coverage.txt -coverpkg=$(shell go list )/...
	./filter-coverage.sh < coverage.txt > coverage.txt.filtered
	go tool cover -func coverage.txt.filtered

.PHONY: bench
bench:
	go test -benchmem -run=^$$ -bench=. ./...

.PHONY: start-services
start-services: stop-services
	$(CONTAINER_RUNTIME) run $(CONTAINER_RUN_ARGS) -d -p 127.0.0.1::5000 -e REGISTRY_STORAGE_DELETE_ENABLED=true --name $(REGISTRY_CONTAINER) docker.io/library/registry:2
	$(CONTAINER_RUNTIME) run $(CONTAINER_RUN_ARGS) -d -p 127.0.0.1::8100 -e ACE_TELEMETRY_LISTEN=:8100 -e ACE_TELEMETRY_DSN=file:/tmp/test.db --name $(TELEMETRY_CONTAINER) --pull always reg.git.act3-ace.com/ace/data/telemetry:latest serve -v

	# Get the ports for the services
	echo -n TEST_REGISTRY= >> .env.test
	$(CONTAINER_RUNTIME) port $(REGISTRY_CONTAINER) 5000/tcp >> .env.test
	echo -n TEST_TELEMETRY=http:// >> .env.test
	$(CONTAINER_RUNTIME) port $(TELEMETRY_CONTAINER) 8100/tcp >> .env.test
	cat .env.test

.PHONY: stop-services
stop-services:
	- echo "" > .env.test
	@mkdir -p log
	- $(CONTAINER_RUNTIME) logs $(REGISTRY_CONTAINER) > log/registry.log 2>&1
	- $(CONTAINER_RUNTIME) rm -f $(REGISTRY_CONTAINER)
	- $(CONTAINER_RUNTIME) logs $(TELEMETRY_CONTAINER) > log/telemetry.jsonl 2>&1
	- $(CONTAINER_RUNTIME) rm -f $(TELEMETRY_CONTAINER)

.PHONY: integration
integration: export ACE_DT_TELEMETRY_USERNAME = ci-test-user
integration: build start-services
	$(eval REGISTRY_HOST := $(shell $(CONTAINER_RUNTIME) port $(REGISTRY_CONTAINER) 5000/tcp))
	$(eval TELEMETRY_HOST := $(shell $(CONTAINER_RUNTIME) port $(TELEMETRY_CONTAINER) 8100/tcp))

	ACE_DT_TELEMETRY_URL=http://$(TELEMETRY_HOST) env | grep ACE
	- rm -rf integration
	@mkdir -p integration

	# This tests an authenticated pull using custom config
	ACE_DT_TELEMETRY_URL=http://$(TELEMETRY_HOST) bin/ace-dt bottle pull reg.git.act3-ace.com/ace/data/tool/bottle/mnist:v1.6 -d integration/bottle
	ACE_DT_TELEMETRY_URL=http://$(TELEMETRY_HOST) bin/ace-dt bottle push $(REGISTRY_HOST)/bottle/mnist:v1.6 -d integration/bottle
	# add a newline
	cat integration/bottleid ; echo
	curl -sSfvo integration/location.txt http://$(TELEMETRY_HOST)/api/location?bottle_digest=$(shell cat integration/bottle/.dt/bottleid)
	cat integration/location.txt
	grep reg.git.act3-ace.com/ace/data/tool/bottle/mnist integration/location.txt
	grep $(REGISTRY_HOST)/bottle/mnist integration/location.txt
	ACE_DT_TELEMETRY_URL=http://$(TELEMETRY_HOST) bin/ace-dt bottle pull bottle:$(shell cat integration/bottle/.dt/bottleid) -d integration/bottle-pull

	# $(MAKE) stop-services

.PHONY: image
image: tool/ko
	VERSION=$(REV) KO_DOCKER_REPO=$(IMAGE_REPO) tool/ko build -B --platform=all --image-label version=$(REV) ./cmd/ace-dt

tool/controller-gen: tool/.controller-gen.$(CONTROLLER_GEN_VERSION)
	GOBIN=$(PWD)/tool go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

tool/.controller-gen.$(CONTROLLER_GEN_VERSION):
	@rm -f tool/.controller-gen.*
	@mkdir -p tool
	touch $@


tool/crd-ref-docs: tool/.crd-ref-docs.$(CRD_REF_DOCS_VERSION)
	GOBIN=$(PWD)/tool go install github.com/elastic/crd-ref-docs@$(CRD_REF_DOCS_VERSION)

tool/.crd-ref-docs.$(CRD_REF_DOCS_VERSION):
	@rm -f tool/.crd-ref-docs.*
	@mkdir -p tool
	touch $@


tool/ko: tool/.ko.$(KO_VERSION)
	GOBIN=$(PWD)/tool go install github.com/google/ko@$(KO_VERSION)

tool/.ko.$(KO_VERSION):
	@rm -f tool/.ko.*
	@mkdir -p tool
	touch $@


tool/golangci-lint: tool/.golangci-lint.$(GOLANGCI_LINT_VERSION)
	@mkdir -p tool
	GOBIN=$(PWD)/tool go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

tool/.golangci-lint.$(GOLANGCI_LINT_VERSION):
	@rm -f tool/.golangci-lint.*
	@mkdir -p tool
	touch $@


.PHONY: tool
tool: tool/controller-gen tool/crd-ref-docs tool/ko tool/golangci-lint

.PHONY: gendoc
gendoc:
	- rm docs/cli/*
	HOME=HOMEDIR ci-dist/ace-dt/linux/amd64/bin/ace-dt gendocs md --only-commands docs/cli/

.PHONY: apidoc
apidoc: $(addsuffix .md, $(addprefix docs/apis/config.dt.act3-ace.io/, v1alpha1))
docs/apis/%.md: tool/crd-ref-docs $(wildcard pkg/apis/$*/*_types.go) 
	@mkdir -p $(@D)
	tool/crd-ref-docs --config=apidocs.yaml --renderer=markdown --source-path=pkg/apis/$* --output-path=$@
