# When you run make VERBOSE=1, executed commands will be printed before
# executed, verbose flags are turned on and quiet flags are turned off for
# various commands. Use V_FLAG in places where you can toggle on/off verbosity
# using -v. Use Q_FLAG in places where you can toggle on/off quiet mode using
# -q.
Q = @
Q_FLAG = -q
V_FLAG =
ifeq ($(VERBOSE),1)
       Q =
       Q_FLAG = 
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

# DOCKER_COMPOSE_BIN := $(shell command -v $(DOCKER_COMPOSE_BIN_NAME) 2> /dev/null)

# # Define and get the vakue for UNAME_S variable from shell
# UNAME_S := $(shell uname -s)

# # This is a fix for a non-existing user in passwd file when running in a docker
# # container and trying to clone repos of dependencies
# GIT_COMMITTER_NAME ?= "user"
# GIT_COMMITTER_EMAIL ?= "user@example.com"
# export GIT_COMMITTER_NAME
# export GIT_COMMITTER_EMAIL

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
test-coverage: ./vendor ./out/cover.out

.PHONY: test-coverage-html
test-coverage-html: ./vendor ./out/cover.out
	$(Q)go tool cover -html=./out/cover.out

.PHONY: docker-image
docker-image: Dockerfile
	$(Q)docker build ${Q_FLAG} \
		--build-arg GO_PACKAGE_PATH=${GO_PACKAGE_PATH} \
		--build-arg VERBOSE=${VERBOSE} \
		. \
		-t ${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}:${GIT_COMMIT_ID}

.PHONY: docker-run
docker-run: docker-image
	$(Q)docker run -it --rm -p 8080:8080 ${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}:${GIT_COMMIT_ID}

./out/app-server: ./vendor $(shell find . -path ./vendor -prune -o -name '*.go' -print)
	$(Q)go build -v ${LDFLAGS} -o ./out/app-server

./vendor: Gopkg.toml Gopkg.lock
	$(Q)dep ensure ${V_FLAG} -vendor-only

./out/cover.out:
	$(Q)go test ${V_FLAG} ./... -failfast -coverprofile=cover.out -covermode=set -outputdir=./out	

# # Build go tool to analysis the code
# $(GOLINT_BIN):
# 	cd $(VENDOR_DIR)/github.com/golang/lint/golint && go build -v
# $(GOCYCLO_BIN):
# 	cd $(VENDOR_DIR)/github.com/fzipp/gocyclo && go build -v

# # Pack all migration SQL files into a compilable Go file
# migration/sqlbindata.go: $(GO_BINDATA_BIN) $(wildcard migration/sql-files/*.sql) migration/sqlbindata_test.go
# 	$(GO_BINDATA_BIN) \
# 		-o migration/sqlbindata.go \
# 		-pkg migration \
# 		-prefix migration/sql-files \
# 		-nocompress \
# 		migration/sql-files

# migration/sqlbindata_test.go: $(GO_BINDATA_BIN) $(wildcard migration/sql-test-files/*.sql)
# 	$(GO_BINDATA_BIN) \
# 		-o migration/sqlbindata_test.go \
# 		-pkg migration_test \
# 		-prefix migration/sql-test-files \
# 		-nocompress \
# 		migration/sql-test-files

# spacetemplate/template_assets.go: $(GO_BINDATA_BIN) $(wildcard spacetemplate/assets/*.yaml)
# 	$(GO_BINDATA_BIN) \
# 		-o spacetemplate/template_assets.go \
# 		-pkg spacetemplate \
# 		-prefix spacetemplate/assets \
# 		-nocompress \
# 		spacetemplate/assets

# swagger/swagger_assets.go: $(GO_BINDATA_BIN) $(wildcard swagger/*.json)
# 	$(GO_BINDATA_BIN) \
# 		-o swagger/swagger_assets.go \
# 		-pkg swagger \
# 		-prefix swagger \
# 		-nocompress \
# 		swagger

# # These are binary tools from our vendored packages
# $(GOAGEN_BIN): $(VENDOR_DIR)
# 	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v
# $(GO_BINDATA_BIN): $(VENDOR_DIR)
# 	cd $(VENDOR_DIR)/github.com/jteeuwen/go-bindata/go-bindata && go build -v
# $(FRESH_BIN): $(VENDOR_DIR)
# 	cd $(VENDOR_DIR)/github.com/pilu/fresh && go build -v
# $(GO_JUNIT_BIN): $(VENDOR_DIR)
# 	cd $(VENDOR_DIR)/github.com/jstemmer/go-junit-report && go build -v

# CLEAN_TARGETS += clean-artifacts
# .PHONY: clean-artifacts
# ## Removes the ./bin directory.
# clean-artifacts:
# 	-rm -rf $(INSTALL_PREFIX)

# CLEAN_TARGETS += clean-object-files
# .PHONY: clean-object-files
# ## Runs go clean to remove any executables or other object files.
# clean-object-files:
# 	go clean ./...

# CLEAN_TARGETS += clean-generated
# .PHONY: clean-generated
# ## Removes all generated code.
# clean-generated:
# 	-rm -rf ./app
# 	-rm -rf ./client/
# 	-rm -rf ./swagger/
# 	-rm -rf ./tool/cli/
# 	-rm -f ./migration/sqlbindata.go
# 	-rm -f ./migration/sqlbindata_test.go
# 	-rm -rf ./account/tenant
# 	-rm -rf ./notification/client
# 	-rm -rf ./auth/authservice
# 	-rm -f ./spacetemplate/template_assets.go

# CLEAN_TARGETS += clean-sql-generated
# .PHONY: clean-sql-generated
# ## Removes all generated code for SQL migration and tests.
# clean-sql-generated:
# 	-rm -f ./migration/sqlbindata.go
# 	-rm -f ./migration/sqlbindata_test.go

# CLEAN_TARGETS += clean-vendor
# .PHONY: clean-vendor
# ## Removes the ./vendor directory.
# clean-vendor:
# 	-rm -rf $(VENDOR_DIR)

# .PHONY: deps
# ## Download build dependencies.
# deps: $(DEP_BIN) $(VENDOR_DIR)

# # install dep in a the tmp/bin dir of the repo
# $(DEP_BIN):
# 	@echo "Installing 'dep' $(DEP_VERSION) at '$(DEP_BIN_DIR)'..."
# 	mkdir -p $(DEP_BIN_DIR)
# ifeq ($(UNAME_S),Darwin)
# 	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64 -o $(DEP_BIN) 
# 	@cd $(DEP_BIN_DIR) && \
# 	echo "1544afdd4d543574ef8eabed343d683f7211202a65380f8b32035d07ce0c45ef  dep" > dep-darwin-amd64.sha256 && \
# 	shasum -a 256 --check dep-darwin-amd64.sha256
# else
# 	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-linux-amd64 -o $(DEP_BIN)
# 	@cd $(DEP_BIN_DIR) && \
# 	echo "31144e465e52ffbc0035248a10ddea61a09bf28b00784fd3fdd9882c8cbb2315  dep" > dep-linux-amd64.sha256 && \
# 	sha256sum -c dep-linux-amd64.sha256
# endif
# 	@chmod +x $(DEP_BIN)

# $(VENDOR_DIR): Gopkg.toml Gopkg.lock
# 	@echo "checking dependencies..."
# 	@$(DEP_BIN) ensure -v 

# app/controllers.go: $(DESIGNS) $(GOAGEN_BIN) $(VENDOR_DIR) goasupport/jsonapi_errors_stringer/generator.go goasupport/conditional_request/generator.go goasupport/helper_function/generator.go
# 	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
# 	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg ${PACKAGE_NAME}/app
# 	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=${PACKAGE_NAME}/goasupport/conditional_request --out app
# 	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=${PACKAGE_NAME}/goasupport/helper_function --out app
# 	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=${PACKAGE_NAME}/goasupport/jsonapi_errors_stringer --out app
# 	$(GOAGEN_BIN) client -d ${PACKAGE_NAME}/${DESIGN_DIR}
# 	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
# 	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-tenant/design --notool --pkg tenant -o account
# 	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-tenant/design --notool --pkg tenant -o vendor/github.com/fabric8-services/fabric8-auth/account
# 	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-notification/design --notool --pkg client -o notification
# 	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --pkg authservice -o auth

# $(MINIMOCK_BIN):
# 	@echo "building the minimock binary..."
# 	@cd $(VENDOR_DIR)/github.com/gojuno/minimock/cmd/minimock && go build -v minimock.go

# .PHONY: generate-minimock
# generate-minimock: deps $(MINIMOCK_BIN) ## Generate Minimock sources. Only necessary after clean or if changes occurred in interfaces.
# 	@echo "Generating mocks..."
# 	-mkdir -p test/controller
# 	$(MINIMOCK_BIN) -i github.com/fabric8-services/fabric8-wit/controller.ClientGetter -o ./test/controller/client_getter_mock.go -t ClientGetterMock
# 	$(MINIMOCK_BIN) -i github.com/fabric8-services/fabric8-wit/controller.OpenshiftIOClient -o ./test/controller/osio_client_mock.go -t OSIOClientMock
# 	-mkdir -p test/kubernetes
# 	$(MINIMOCK_BIN) -i github.com/fabric8-services/fabric8-wit/kubernetes.KubeClientInterface -o ./test/kubernetes/kube_client_mock.go -t KubeClientMock

# .PHONY: migrate-database
# ## Compiles the server and runs the database migration with it
# migrate-database: $(BINARY_SERVER_BIN)
# 	$(BINARY_SERVER_BIN) -migrateDatabase

# .PHONY: generate
# ## Generate GOA sources. Only necessary after clean of if changed `design` folder.
# generate: app/controllers.go migration/sqlbindata.go spacetemplate/template_assets.go generate-minimock swagger/swagger_assets.go

# .PHONY: regenerate
# ## Runs the "clean-generated" and the "generate" target
# regenerate: clean-generated generate

# .PHONY: dev
# dev: prebuild-check deps generate $(FRESH_BIN) docker-compose-up
# 	F8_DEVELOPER_MODE_ENABLED=true $(FRESH_BIN)

# .PHONY: docker-compose-up
# docker-compose-up:
# ifeq ($(UNAME_S),Darwin)
# 	@echo "Running docker-compose with macOS network settings"
# 	docker-compose -f docker-compose.macos.yml up -d db auth swagger_ui
# else
# 	@echo "Running docker-compose with Linux network settings"
# 	docker-compose up -d db auth swagger_ui
# endif

# MINISHIFT_IP = `minishift ip`
# MINISHIFT_URL = http://$(MINISHIFT_IP)
# # make sure you have a entry in /etc/hosts for "minishift.local MINISHIFT_IP"
# MINISHIFT_HOSTS_ENTRY = http://minishift.local

# # Setup AUTH image URL, use default image path and default tag if not provided
# AUTH_IMAGE_DEFAULT=quay.io/openshiftio/fabric8-services-fabric8-auth
# AUTH_IMAGE_TAG ?= latest
# AUTH_IMAGE_URL=$(AUTH_IMAGE_DEFAULT):$(AUTH_IMAGE_TAG)

# .PHONY: dev-wit-openshift
# ## Starts minishift and creates/uses a project named wit-openshift
# ## Deploys DB, DB-AUTH, AUTH services from minishift/kedge/*.yml
# ## Runs WIT service on local machine
# dev-wit-openshift: prebuild-check deps generate $(FRESH_BIN)
# 	minishift start
# 	./check_hosts.sh
# 	-eval `minishift oc-env` &&  oc login -u developer -p developer && oc new-project wit-openshift
# 	AUTH_IMAGE_URL=$(AUTH_IMAGE_URL) \
# 	AUTH_WIT_URL=$(MINISHIFT_URL):8080 \
# 	kedge apply -f minishift/kedge/db.yml -f minishift/kedge/db-auth.yml -f minishift/kedge/auth.yml
# 	sleep 3s
# 	F8_AUTH_URL=$(MINISHIFT_HOSTS_ENTRY):31000 \
# 	F8_POSTGRES_HOST=$(MINISHIFT_IP) \
# 	F8_POSTGRES_PORT=32000 \
# 	F8_DEVELOPER_MODE_ENABLED=true \
# 	$(FRESH_BIN)

# .PHONY: dev-wit-openshift-clean
# ## Removes the project created by `make dev-wit-openshift`
# dev-wit-openshift-clean:
# 	-eval `minishift oc-env` &&  oc login -u developer -p developer && oc delete project wit-openshift --force

# include ./.make/test.mk

# ifneq ($(OS),Windows_NT)
# ifdef DOCKER_BIN
# include ./.make/docker.mk
# endif
# endif

# $(INSTALL_PREFIX):
# # Build artifacts dir
# 	mkdir -p $(INSTALL_PREFIX)

# $(TMP_PATH):
# 	mkdir -p $(TMP_PATH)

# .PHONY: show-info
# show-info:
# 	$(call log-info,"$(shell go version)")
# 	$(call log-info,"$(shell go env)")

# .PHONY: prebuild-check
# prebuild-check: $(TMP_PATH) $(INSTALL_PREFIX) $(CHECK_GOPATH_BIN) show-info
# # Check that all tools where found
# ifndef GIT_BIN
# 	$(error The "$(GIT_BIN_NAME)" executable could not be found in your PATH)
# endif
# ifndef DEP_BIN
# 	$(error The "$(DEP_BIN_NAME)" executable could not be found in your PATH)
# endif
# 	@$(CHECK_GOPATH_BIN) -packageName=$(PACKAGE_NAME) || (echo "Project lives in wrong location"; exit 1)

# $(CHECK_GOPATH_BIN): .make/check_gopath.go
# ifndef GO_BIN
# 	$(error The "$(GO_BIN_NAME)" executable could not be found in your PATH)
# endif
# ifeq ($(OS),Windows_NT)
# 	@go build -o "$(shell cygpath --windows '$(CHECK_GOPATH_BIN)')" .make/check_gopath.go
# else
# 	@go build -o $(CHECK_GOPATH_BIN) .make/check_gopath.go
# endif