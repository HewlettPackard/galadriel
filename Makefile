DIR := ${CURDIR}


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

go-bin-path: go-check
	@echo "$(go_bin_dir):${PATH}"

install-toolchain: install-protoc install-protoc-gen-go | go-check

# install-protoc: $(protoc_bin)

# install-protoc-gen-go: $(protoc_gen_go_bin)

# $(protoc_gen_go_bin): | go-check
# 	@echo "Installing protoc-gen-go $(protoc_gen_go_version)..."
# 	$(E)rm -rf $(protoc_gen_go_base_dir)
# 	$(E)mkdir -p $(protoc_gen_go_dir)
# 	$(E)GOBIN=$(protoc_gen_go_dir) $(go_path) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(protoc_gen_go_version)


# $(protoc_bin):
# 	@echo "Installing protoc $(protoc_version)..."
# 	$(E)rm -rf $(dir $(protoc_dir))
# 	$(E)mkdir -p $(protoc_dir)
# 	$(E)curl -sSfL $(protoc_url) -o $(build_dir)/tmp.zip; unzip -q -d $(protoc_dir) $(build_dir)/tmp.zip; rm $(build_dir)/tmp.zip





# protos := \
# 	proto/jwtglue/jwtglue.proto

# The following vars are used in rule construction
comma := ,
null  :=
space := $(null) 

.PHONY: build
build: bin/spire-bridge-server

# This is the master template for compiling Go binaries
define binary_rule
.PHONY: $1
$1: | go-check bin/
	@echo Building $1...
	$(E)$(go_path) go build -o $1 $2
endef

# This dynamically generates targets for each binary using
# the binary_rule template above
$(eval $(call binary_rule,bin/spire-bridge-server,./cmd/jwtglue))


# #
# # code generation
# # 
# #

# .PHONY: generate

# generate: $(protos:.proto=.pb.go)

# #proto/jwtglue/jwtglue.pb.go: proto/jwtglue/jwtglue.proto
# #	@echo "got to $@"

# %.pb.go: %.proto $(protoc_bin) $(protoc_gen_go_bin)
# 	@echo "got to $@"

# #%_.pb.go: %.proto $(protoc_bin) $(protoc_gen_go_bin) FORCE | bin/protoc-gen-go-spire
# #	@echo "generating $@..."
# #	$(E) PATH="$(protoc_gen_go_dir):$(PATH)" $(protoc_bin) \
# #		-I proto \
# #		--go-spire_out=. \
# #		--go-spire_opt=module=github.com/dfeldman/jwtglue \
# #		--go-spire_opt=mode=plugin \
# #		$<

CONTAINER_OPTIONS = docker podman
CONTAINER_EXEC := $(foreach exec,$(CONTAINER_OPTIONS),\
     $(if $(shell which $(exec)),$(exec)))


api-doc-build: 
	$(CONTAINER_EXEC) build -f dev/api/Dockerfile -t galadriel-api-doc:latest .

api-doc: api-doc-build
	$(CONTAINER_EXEC) run --rm \
		--name galadriel-api-doc \
		-p 8000:8000 \
		--mount type=bind,source=${DIR}/dev/api/,target=/app/api,readonly \
		galadriel-api-doc:latest

test: test-unit

test-unit:
	go test -cover ./...

coverage:
	go test -v -coverprofile ./out/coverage/coverage.out ./... && \
	go tool cover -html=./out/coverage/coverage.out -o ./out/coverage/index.html
