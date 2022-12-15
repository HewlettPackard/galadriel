-include Makefile-poc.mk

DIR := ${CURDIR}

.DEFAULT_GOAL = help

E:=@
ifeq ($(V),1)
	E=
endif

cyan := $(shell which tput > /dev/null && tput setaf 6 2>/dev/null || echo "")
reset := $(shell which tput > /dev/null && tput sgr0 2>/dev/null || echo "")
bold  := $(shell which tput > /dev/null && tput bold 2>/dev/null || echo "")

.PHONY: default 

default: build

all: build test

############################################################################
# OS/ARCH detection
############################################################################
os1=$(shell uname -s)
os2=
ifeq ($(os1),Darwin)
os1=darwin
os2=osx
else ifeq ($(os1),Linux)
os1=linux
os2=linux
else ifeq (,$(findstring MYSYS_NT-10-0-, $(os1)))
os1=windows
os2=windows
else
$(error unsupported OS: $(os1))
endif

arch1=$(shell uname -m)
ifeq ($(arch1),x86_64)
arch2=amd64
else ifeq ($(arch1),aarch64)
arch2=arm64
else ifeq ($(arch1),arm64)
arch2=arm64
else
$(error unsupported ARCH: $(arch1))
endif


############################################################################
# Vars
############################################################################

build_dir := $(DIR)/.build/$(os1)-$(arch1)

go_version_full := $(shell cat .go-version)
go_version := $(go_version_full:.0=)
go_dir := $(build_dir)/go/$(go_version)

ifeq ($(os1),windows)
	go_bin_dir = $(go_dir)/go/bin
	go_url = https://storage.googleapis.com/golang/go$(go_version).$(os1)-$(arch2).zip
	exe=".exe"
else 
	go_bin_dir = $(go_dir)/bin
	go_url = https://storage.googleapis.com/golang/go$(go_version).$(os1)-$(arch2).tar.gz
	exe=
endif

go_path := PATH="$(go_bin_dir):$(PATH)"

oapi_codegen_version = 1.11.0
oapi_codegen_dir = $(build_dir)/protoc/$(protoc_version):q

sqlc_dir = $(build_dir)/sqlc/$(sqlc_version)
sqlc_bin = $(sqlc_dir)/sqlc
sqlc_version = 1.16.0
sqlc_config_file = $(DIR)/db/sqlc.yaml
ifeq ($(os1),windows)
	sqlc_url = https://github.com/kyleconroy/sqlc/releases/download/v${sqlc_version}/sqlc_${sqlc_version}_windows_amd64.zip
else ifeq ($(os1),darwin)
	sqlc_url = https://github.com/kyleconroy/sqlc/releases/download/v${sqlc_version}/sqlc_${sqlc_version}_darwin_$(arch2).zip
else
	sqlc_url = https://github.com/kyleconroy/sqlc/releases/download/v${sqlc_version}/sqlc_${sqlc_version}_linux_amd64.zip
endif

go-check:
ifeq (go$(go_version), $(shell $(go_path) go version 2>/dev/null | cut -f3 -d' '))
else
	@echo "Installing go $(go_version)..."
	$(E)rm -rf $(dir $(go_dir))
	$(E)mkdir -p $(go_dir)
	$(E)curl -sSfL $(go_url) | tar xz -C $(go_dir) --strip-components=1
endif

## Checks installed go version and prints the path it is installed.
go-bin-path: go-check
	@echo "$(go_bin_dir):${PATH}"

install-toolchain: install-sqlc | go-check

install-sqlc: $(sqlc_bin)

$(sqlc_bin):
	@echo "Installing sqlc $(sqlc_version)..."
	$(E)rm -rf $(dir $(sqlc_dir))
	$(E)mkdir -p $(sqlc_dir)
	$(E)echo $(sqlc_url); curl -sSfL $(sqlc_url) -o $(build_dir)/tmp.zip; unzip -q -d $(sqlc_dir) $(build_dir)/tmp.zip; rm $(build_dir)/tmp.zip


# The following vars are used in rule construction
comma := ,
null  :=
space := $(null) 

.PHONY: build

## Compiles all Galadriel binaries.
build: bin/galadriel-harvester bin/galadriel-server

# This is the master template for compiling Go binaries
define binary_rule
.PHONY: $1
$1: | go-check bin/
	@echo Building $1...
	$(E)$(go_path) go build -o $1 $2
endef

# This dynamically generates targets for each binary using
# the binary_rule template above
$(eval $(call binary_rule,bin/galadriel-harvester,cmd/harvester/main.go))
$(eval $(call binary_rule,bin/galadriel-server,cmd/server/main.go))

bin/:
	@mkdir -p $@

CONTAINER_OPTIONS = docker podman
CONTAINER_EXEC := $(foreach exec,$(CONTAINER_OPTIONS),\
     $(if $(shell which $(exec)),$(exec)))

server-run: build
	./bin/galadriel-server run

## Runs the go unit tests.
test: test-unit

test-unit:
	go test -cover ./...

## Runs unit tests with race detection.
race-test:
	go test -cover -race ./...

## Generates the test coverage for the code with the Go tool.
coverage:
	$(E)mkdir -p out/coverage
	go test -v -coverprofile ./out/coverage/coverage.out ./... && \
	go tool cover -html=./out/coverage/coverage.out -o ./out/coverage/index.html

## Builds docker image for Galadriel Server.
docker-build-server:
	docker build . --target galadriel-server --tag galadriel-server:latest

## Builds docker image for Galadriel Harvester.
docker-build-harvester:
	docker build . --target galadriel-harvester --tag galadriel-harvester:latest

## Builds all docker images.
docker-build: docker-build-server docker-build-harvester

#------------------------------------------------------------------------
# Document file
#------------------------------------------------------------------------

# VARIABLES
NAME = Galadriel
VERSION = 0.1.0
AUTHOR=HPE

# COLORS
GREEN := $(shell tput -Txterm setaf 2)
RESET := $(shell tput -Txterm sgr0)

TARGET_MAX_CHAR_NUM=30

## Shows help.
help:
	@echo "$(bold)Usage:$(reset) make $(cyan)<target>$(reset)"
	@echo
	@echo "$(bold)Build:$(reset)"
	@echo "  $(cyan)build$(reset)                                 - build all Galadriel binaries"
	@echo
	@echo "$(bold)Test:$(reset)"
	@echo "  $(cyan)test$(reset)                                  - run unit tests"
	@echo
	@echo "$(bold)Build and test:$(reset)"
	@echo "  $(cyan)all$(reset)                                   - build all Galadriel binaries, and run unit tests"
	@echo
	@echo "$(bold)Code generation:$(reset)"
	@echo "  $(cyan)generate$(reset)                              - generate datastore sql code"

### Code generation ####
.PHONE: generate generatesql

generate: generatesql

generatesql:
	$(sqlc_bin) generate --file $(sqlc_config_file)
