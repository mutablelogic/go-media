# Paths to packages
GO=$(shell which go)
DOCKER=$(shell which docker)

# Build flags
BUILD_MODULE := $(shell cat go.mod | head -1 | cut -d ' ' -f 2)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')
DOCKER_REGISTRY ?= ghcr.io/mutablelogic

# Paths to locations, etc
BUILD_DIR := "build"
CMD_DIR := $(filter-out cmd/ffmpeg/README.md, $(wildcard cmd/ffmpeg/*))
BUILD_TAG := ${DOCKER_REGISTRY}/go-media-${OS}-${ARCH}:${VERSION}

all: clean cmds

cmds: $(CMD_DIR)

docker: docker-dep
	@echo build docker image: ${BUILD_TAG} for ${OS}/${ARCH}
	@${DOCKER} build \
		--tag ${BUILD_TAG} \
		--build-arg ARCH=${ARCH} \
		--build-arg OS=${OS} \
		--build-arg SOURCE=${BUILD_MODULE} \
		--build-arg VERSION=${VERSION} \
		-f etc/docker/Dockerfile .

docker-push: docker-dep
	@echo push docker image: ${BUILD_TAG}
	@${DOCKER} push ${BUILD_TAG}

test: go-dep
	@echo Test
	@${GO} mod tidy
	@echo ... test sys/ffmpeg71
	@${GO} test ./sys/ffmpeg71
	@echo ... test pkg/ffmpeg
	@${GO} test -v ./pkg/ffmpeg
	@echo ... test sys/chromaprint
	@${GO} test ./sys/chromaprint
	@echo ... test pkg/chromaprint
	@${GO} test ./pkg/chromaprint
	@echo ... test pkg/file
	@${GO} test ./pkg/file
	@echo ... test pkg/generator
	@${GO} test ./pkg/generator
	@echo ... test pkg/image
	@${GO} test ./pkg/image
	@echo ... test pkg
	@${GO} test ./pkg/...

container-test: go-dep
	@echo Test
	@${GO} mod tidy
	@${GO} test --tags=container ./sys/ffmpeg71
	@${GO} test --tags=container ./sys/chromaprint
	@${GO} test --tags=container ./pkg/...
	@${GO} test --tags=container .

cli: go-dep mkdir
	@echo Build media tool
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/media ./cmd/cli

$(CMD_DIR): go-dep mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

FORCE:

go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
