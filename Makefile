include versions.mk

cyan := $(shell which tput > /dev/null && tput setaf 6 2>/dev/null || echo "")
reset := $(shell which tput > /dev/null && tput sgr0 2>/dev/null || echo "")
bold := $(shell which tput > /dev/null && tput bold 2>/dev/null || echo "")

# Vars
DIR := $(CURDIR)
.DEFAULT_GOAL = help
E := @
ifeq ($(V),1)
	E =
endif

# OS/ARCH detection
os1 := $(shell uname -s)
os2 :=
ifeq ($(os1),Darwin)
os1 := darwin
os2 := osx
else ifeq ($(os1),Linux)
os1 := linux
os2 := linux
else ifeq (,$(findstring MYSYS_NT-10-0-, $(os1)))
os1 := windows
os2 := windows
else
$(error unsupported OS: $(os1))
endif

arch1 := $(shell uname -m)
ifeq ($(arch1),x86_64)
arch2 := amd64
else ifeq ($(arch1),aarch64)
arch2 := arm64
else ifeq ($(arch1),arm64)
arch2 := arm64
else
$(error unsupported ARCH: $(arch1))
endif

# Define build directories and URLs
build_dir := $(DIR)/.build/$(os1)-$(arch1)
go_dir := $(build_dir)/go/$(GO_VERSION)
server_sqlc_config_file := $(DIR)/pkg/server/db/sqlc.yaml
sqlc_dir := $(build_dir)/sqlc/$(SQLC_VERSION)
sqlc_bin := $(sqlc_dir)/sqlc

ifeq ($(os1),windows)
go_bin_dir := $(go_dir)/go/bin
go_url := https://storage.googleapis.com/golang/go$(GO_VERSION).$(os1)-$(arch2).zip
exe := .exe
else
go_bin_dir := $(go_dir)/bin
go_url := https://storage.googleapis.com/golang/go$(GO_VERSION).$(os1)-$(arch2).tar.gz
exe :=
endif

# SQLC download URL
ifeq ($(os1),windows)
sqlc_url := https://github.com/kyleconroy/sqlc/releases/download/v$(SQLC_VERSION)/sqlc_$(SQLC_VERSION)_windows_amd64.zip
else ifeq ($(os1),darwin)
sqlc_url := https://github.com/kyleconroy/sqlc/releases/download/v$(SQLC_VERSION)/sqlc_$(SQLC_VERSION)_darwin_$(arch2).zip
else
sqlc_url := https://github.com/kyleconroy/sqlc/releases/download/v$(SQLC_VERSION)/sqlc_$(SQLC_VERSION)_linux_amd64.zip
endif

# Define Go path
go_path := PATH="$(go_bin_dir):$(PATH)"

# Define master template for compiling Go binaries
define binary_rule
.PHONY: $1
$1: | go-check bin/
	@echo "Building $1..."
	$(E)$(go_path) go build -o $1 $2
endef

# Dynamically generate targets for each binary using the binary_rule template
$(eval $(call binary_rule,bin/galadriel-harvester,cmd/harvester/main.go))
$(eval $(call binary_rule,bin/galadriel-server,cmd/server/main.go))

# Build directories
bin/:
	@mkdir -p $@

# Go check and installation if necessary
go-check:
ifeq (go$(GO_VERSION), $(shell $(go_path) go version 2>/dev/null | cut -f3 -d' '))
else
	@echo "Installing go $(GO_VERSION)..."
	$(E)rm -rf $(dir $(go_dir))
	$(E)mkdir -p $(go_dir)
	$(E)curl -sSfL $(go_url) | tar xz -C $(go_dir) --strip-components=1
endif

# Prints Go binary installation path
go-bin-path: go-check
	@echo "$(go_bin_dir):${PATH}"

# Install necessary toolchains
install-toolchain: install-sqlc | go-check

# Install SQLC
install-sqlc: $(sqlc_bin)
$(sqlc_bin):
	@echo "Installing sqlc $(SQLC_VERSION)..."
	$(E)rm -rf $(dir $(sqlc_dir))
	$(E)mkdir -p $(sqlc_dir)
	$(E)echo $(sqlc_url); curl -sSfL $(sqlc_url) -o $(build_dir)/tmp.zip; unzip -q -d $(sqlc_dir) $(build_dir)/tmp.zip; rm $(build_dir)/tmp.zip

# Rules for building binaries and running tests
default: build
all: build test
build: bin/galadriel-harvester bin/galadriel-server
test: test-unit
test-unit:
	go test -cover ./...
race-test:
	go test -cover -race ./...
coverage:
	$(E)mkdir -p out/coverage
	go test -v -coverprofile ./out/coverage/coverage.out ./... && \
	go tool cover -html=./out/coverage/coverage.out -o ./out/coverage/index.html
clean:
	rm -rf $(build_dir)
	rm -f bin/galadriel-harvester
	rm -f bin/galadriel-server
	rm -rf out/coverage
.PHONY: clean


# Generate SQL and API code
generate-sql-code: install-sqlc $(server_sqlc_config_file)
	@echo "Generating server SQL code..."
	$(sqlc_bin) generate --file $(server_sqlc_config_file)
generate-api-code: $(SPEC_FILES)
	@echo "Generating API code..."
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v$(OAPI_CODEGEN_VERSION)
	cd ./pkg/common/api; $(GOPATH)/bin/oapi-codegen -config schemas.cfg.yaml schemas.yaml
	cd ./pkg/server/api/admin; $(GOPATH)/bin/oapi-codegen -config admin.cfg.yaml admin.yaml
	cd ./pkg/server/api/harvester; $(GOPATH)/bin/oapi-codegen -config harvester.cfg.yaml harvester.yaml
	cd ./pkg/harvester/api/admin; $(GOPATH)/bin/oapi-codegen -config admin.cfg.yaml admin.yaml

# Help rule
help:
	@echo "$(bold)Usage:$(reset) make $(cyan)<target>$(reset)"
	@echo
	@echo "$(bold)Build:$(reset)"
	@echo "  $(cyan)build$(reset)                                 - build all Galadriel binaries"
	@echo
	@echo "$(bold)Test:$(reset)"
	@echo "  $(cyan)test$(reset)                                  - run unit tests"
	@echo
	@echo "$(bold)Toolchain:$(reset)"
	@echo "  $(cyan)install-toolchain$(reset)                    - install required build tools"
	@echo "  $(cyan)go-bin-path$(reset)                          - print path of installed go binary"
	@echo
	@echo "$(bold)Code Generation:$(reset)"
	@echo "  $(cyan)generate-sql-code$(reset)                    - generate sql code using sqlc"
	@echo "  $(cyan)generate-api-code$(reset)                    - generate api code using oapi-codegen"
	@echo
	@echo "$(bold)Cleanup:$(reset)"
	@echo "  $(cyan)clean$(reset)                                - clean build artifacts"
	@echo
.PHONY: help
