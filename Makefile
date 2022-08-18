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

all: build 

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


go-check:
ifeq (go$(go_version), $(shell $(go_path) go version 2>/dev/null | cut -f3 -d' '))
	@echo "Installing go$(go_version)..."
	$(E)rm -rf $(dir $(go_dir))
	$(E)mkdir -p $(go_dir)
	$(E)curl -sSfL $(go_url) | tar xz -C $(go_dir) --strip-components=1
endif

## Checks installed go version and prints the path it is installed.
go-bin-path: go-check
	@echo "$(go_bin_dir):${PATH}"

# The following vars are used in rule construction
comma := ,
null  :=
space := $(null) 

.PHONY: build

## Compile Go binaries for the Galadriel.
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

api-doc-build:
	$(CONTAINER_EXEC) build -f doc/api/Dockerfile -t galadriel-api-doc:latest .

## Build the API documentation for the Galadriel.
api-doc: api-doc-build
	$(CONTAINER_EXEC) run --rm \
		--name galadriel-api-doc \
		-p 8000:8000 \
		--mount type=bind,source=${DIR}/spec/api,target=/app/api,readonly \
		galadriel-api-doc:latest

## Runs the go unit tests.
test: test-unit

test-unit:
	go test -cover ./...

## Generate the test coverage for the code with the Go tool.
coverage:
	$(E)mkdir -p out/coverage
	go test -v -coverprofile ./out/coverage/coverage.out ./... && \
	go tool cover -html=./out/coverage/coverage.out -o ./out/coverage/index.html

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

TARGET_MAX_CHAR_NUM=20

## Shows help.
help:
	@echo "--------------------------------------------------------------------------------"
	@echo "Author  : ${GREEN}$(AUTHOR)${RESET}"
	@echo "Project : ${GREEN}$(NAME)${RESET}"
	@echo "Version : ${GREEN}$(VERSION)${RESET}"
	@echo "--------------------------------------------------------------------------------"
	@echo ""
	@echo "Usage:"
	@echo "  ${GREEN}make${RESET} <target>"
	@echo "Targets:"
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${GREEN}%-$(TARGET_MAX_CHAR_NUM)s${RESET} %s\n", helpCommand, helpMessage; \
		} \
	} \
{ lastLine = $$0 }' $(MAKEFILE_LIST)
