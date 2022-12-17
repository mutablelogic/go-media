# Paths to packages
GO=$(shell which go)
DOCKER=$(shell which docker)

# Paths to locations, etc
BUILD_DIR := "build"
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))

# Build flags
BUILD_MODULE := $(shell go list -m)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

all: clean test cmd

cmd: clean dependencies $(CMD_DIR)

$(CMD_DIR): FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

docker:
	@echo Build docker image
	@${DOCKER} build \
	    --tag go-media:$(shell git describe --tags) \
		--build-arg PLATFORM=$(shell ${GO} env GOOS) \
		--build-arg ARCH=$(shell ${GO} env GOARCH) \
		--build-arg VERSION=kinetic \
		-f etc/docker/Dockerfile .

test: clean dependencies
	@echo Test sys/
	@${GO} test ./sys/...
	@echo Test pkg/
	@${GO} test ./pkg/...

dependencies: mkdir
ifeq (,${GO})
        $(error "Missing go binary")
endif

mkdir:
	@echo Mkdir
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
