# When you run make VERBOSE=1, executed commands will be printed before
# executed, verbose flags are turned on and quiet flags are turned off for
# various commands. Use V_FLAG in places where you can toggle on/off verbosity
# using -v. Use Q_FLAG in places where you can toggle on/off quiet mode using
# -q. Use S_FLAG where you want to toggle on/off silence mode using -s.
Q = @
Q_FLAG = -q
V_FLAG =
S_FLAG = -s
ifeq ($(VERBOSE),1)
       Q =
       Q_FLAG = 
       S_FLAG = 
       V_FLAG = -v
endif

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

# Check all required tools are accessible (only when not running a docker-
# target).
GO_BIN ?= go
DEP_BIN ?= dep
GIT_BIN ?= git
DOCKER_BIN ?= docker
REQUIRED_EXECUTABLES = $(GO_BIN) $(DEP_BIN) $(GIT_BIN)
ifeq ($(MAKECMDGOALS),docker-image)
    REQUIRED_EXECUTABLES = $(DOCKER_BIN)
endif
ifeq ($(MAKECMDGOALS),docker-run)
    REQUIRED_EXECUTABLES = $(DOCKER_BIN)
endif
ifeq ($(VERBOSE),1)
$(info Searching for required executables: $(REQUIRED_EXECUTABLES)...)
endif
K := $(foreach exec,$(REQUIRED_EXECUTABLES),\
        $(if $(shell which $(exec) 2>/dev/null),some string,$(error "ERROR: No "$(exec)" binary found in in PATH!")))

# Create output directory for artifacts and test results
$(shell mkdir -p ./out);

GIT_COMMIT_ID := $(shell git rev-parse HEAD)
ifneq ($(shell git status --porcelain --untracked-files=no),)
       GIT_COMMIT_ID := $(GIT_COMMIT_ID)-dirty
endif
BUILD_TIME = `date -u '+%Y-%m-%dT%H:%M:%SZ'`

# By default the project should be build under 
GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}

# Pass in build time variables to main
LDFLAGS := -ldflags "-X ${GO_PACKAGE_PATH}/appserver.Commit=${GIT_COMMIT_ID} -X ${GO_PACKAGE_PATH}/appserver.BuildTime=${BUILD_TIME}"

.PHONY: all
all: ./out/app-server

.PHONY: clean
clean:
	$(Q)-rm -rf ${V_FLAG} ./out
	$(Q)-rm -rf ${V_FLAG} ./vendor

.PHONY: test
test: ./vendor
	$(Q)go test ${V_FLAG} ./... -failfast

.PHONY: test-coverage
test-coverage: ./out/cover.out

.PHONY: test-coverage-html
test-coverage-html: ./vendor ./out/cover.out
	$(Q)go tool cover -html=./out/cover.out

.PHONY: docker-image
docker-image: Dockerfile
	$(Q)docker build ${Q_FLAG} \
		--network host \
		--build-arg GO_PACKAGE_PATH=${GO_PACKAGE_PATH} \
		--build-arg VERBOSE=${VERBOSE} \
		. \
		-t ${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}:${GIT_COMMIT_ID}

.PHONY: docker-run
docker-run: docker-image
	$(Q)docker run -it --rm -p 8080:8080 --network host ${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}:${GIT_COMMIT_ID}

.PHONY: docker-test
docker-test: docker-image
	$(Q)docker rm -f app-service-test
	$(Q)docker run --name app-service-test  -dt --rm -p 8123:8080 ${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}:${GIT_COMMIT_ID}
	$(Q)curl ${S_FLAG} 0.0.0.0:8123/status?format=yaml
	$(Q)docker rm -f app-service-test

.PHONY: codecov
codecov: ./out/cover.out
	$(Q)CODECOV_TOKEN="c2340826-e1f9-4c12-a2b1-f1e98b3a040e" bash <(curl -s https://codecov.io/bash)

./out/app-server: ./vendor $(shell find . -path ./vendor -prune -o -name '*.go' -print)
	$(Q)go build -v ${LDFLAGS} -o ./out/app-server

./vendor: Gopkg.toml Gopkg.lock
	$(Q)dep ensure ${V_FLAG} -vendor-only

./out/cover.out: ./vendor
	$(Q)go test ${V_FLAG} -race ./... -failfast -coverprofile=cover.out -covermode=atomic -outputdir=./out
