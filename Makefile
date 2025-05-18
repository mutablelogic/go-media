# Paths to packages
GO=$(shell which go)
DOCKER=$(shell which docker)

# Source version
FFMPEG_VERSION=ffmpeg-7.1.1
CHROMAPRINT_VERSION=chromaprint-1.5.1

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
PREFIX ?= ${BUILD_DIR}/install

###############################################################################
# TARGETS

.PHONY: all
all: clean ffmpeg cli

.PHONY: cmds
cmds: $(CMD_DIR)

.PHONY: cli
cli: go-dep go-tidy mkdir
	@echo Build media tool
	@PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/media ./cmd/media

$(CMD_DIR): go-dep go-tidy mkdir
	@echo Build cmd $(notdir $@)
	@PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

###############################################################################
# FFMPEG

# Download ffmpeg sources
${BUILD_DIR}/${FFMPEG_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/$(FFMPEG_VERSION)" ]; then \
		echo "Downloading $(FFMPEG_VERSION)"; \
		mkdir -p $(BUILD_DIR)/${FFMPEG_VERSION}; \
		curl -L -o $(BUILD_DIR)/ffmpeg.tar.gz https://ffmpeg.org/releases/$(FFMPEG_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/ffmpeg.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/ffmpeg.tar.gz; \
	fi

# Configure ffmpeg
.PHONY: ffmpeg-configure
ffmpeg-configure: mkdir ${BUILD_DIR}/${FFMPEG_VERSION} ffmpeg-dep
	@echo "Configuring ${FFMPEG_VERSION} => ${PREFIX}"	
	@cd ${BUILD_DIR}/${FFMPEG_VERSION} && ./configure \
		--enable-static --disable-doc --disable-programs \
		--prefix="$(shell realpath ${PREFIX})" \
	  	--pkg-config-flags="--static" \
		--extra-libs="-lpthread" \
		--enable-gpl --enable-nonfree ${FFMPEG_CONFIG}

# Build ffmpeg
.PHONY: ffmpeg-build
ffmpeg-build: ffmpeg-configure
	@echo "Building ${FFMPEG_VERSION}"
	@cd $(BUILD_DIR)/$(FFMPEG_VERSION) && make -j2

# Install ffmpeg
.PHONY: ffmpeg
ffmpeg: ffmpeg-build
	@echo "Installing ${FFMPEG_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR)/$(FFMPEG_VERSION) && make install

###############################################################################
# CHROMAPRINT

# Download ffmpeg sources
${BUILD_DIR}/${CHROMAPRINT_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/$(CHROMAPRINT_VERSION)" ]; then \
		echo "Downloading $(CHROMAPRINT_VERSION)"; \
		mkdir -p $(BUILD_DIR)/${CHROMAPRINT_VERSION}; \
		curl -L -o $(BUILD_DIR)/chromaprint.tar.gz https://github.com/acoustid/chromaprint/releases/download/v1.5.1/$(CHROMAPRINT_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/chromaprint.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/chromaprint.tar.gz; \
	fi


# Configure chromaprint
.PHONY: chromaprint-configure
chromaprint-configure: mkdir ${BUILD_DIR}/${CHROMAPRINT_VERSION}
	@echo "Configuring ${CHROMAPRINT_VERSION} => ${PREFIX}"	
	cmake \
		-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=0 \
		-DBUILD_TESTS=0 \
		-DBUILD_TOOLS=0 \
		--install-prefix "$(shell realpath ${PREFIX})" \
		-S ${BUILD_DIR}/${CHROMAPRINT_VERSION} \
		-B ${BUILD_DIR}

# Build chromaprint
.PHONY: chromaprint-build
chromaprint-build: chromaprint-configure
	@echo "Building ${CHROMAPRINT_VERSION}"
	@cd $(BUILD_DIR) && make -j2

# Install chromaprint
.PHONY: chromaprint
chromaprint: chromaprint-build
	@echo "Installing ${CHROMAPRINT_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR) && make install

###############################################################################
# DOCKER

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

###############################################################################
# TESTS

.PHONY: test
test: test-ffmpeg test-chromaprint

.PHONY: test-chromaprint
test-chromaprint: go-dep go-tidy
	@echo Test
	@echo ... test sys/chromaprint
	@PKG_CONFIG_PATH=$(shell realpath ${PREFIX})/lib/pkgconfig ${GO} test ./sys/chromaprint

.PHONY: test-ffmpeg
test-ffmpeg: go-dep go-tidy
	@echo Test
	@echo ... test sys/ffmpeg71
	@PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" ${GO} test ./sys/ffmpeg71
	@echo ... test pkg/segmenter
	@PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" ${GO} test ./pkg/segmenter
	@echo ... test pkg/chromaprint
	@PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" ${GO} test ./pkg/chromaprint


#	@echo ... test pkg/ffmpeg
#	@${GO} test -v ./pkg/ffmpeg
#	@echo ... test sys/chromaprint
#	@${GO} test ./sys/chromaprint
#	@echo ... test pkg/chromaprint
#	@${GO} test ./pkg/chromaprint
#	@echo ... test pkg/file
#	@${GO} test ./pkg/file
#	@echo ... test pkg/generator
#	@${GO} test ./pkg/generator
#	@echo ... test pkg/image
#	@${GO} test ./pkg/image
#	@echo ... test pkg
#	@${GO} test ./pkg/...

container-test: go-dep go-tidy
	@echo Test
	@${GO} mod tidy
	@${GO} test --tags=container ./sys/ffmpeg71
	@${GO} test --tags=container ./sys/chromaprint
	@${GO} test --tags=container ./pkg/...
	@${GO} test --tags=container .


###############################################################################
# DEPENDENCIES, ETC

.PHONY: go-dep
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

.PHONY: docker-dep
docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

.PHONY: mkdir
mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}
	@install -d ${PREFIX}

.PHONY: go-tidy
go-tidy:
	@echo Tidy
	@${GO} mod tidy

.PHONY: clean
clean: go-tidy
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} clean -cache

# Check for FFmpeg dependencies
.PHONY: ffmpeg-dep
ffmpeg-dep:
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists libass && echo "--enable-libass"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists fdk-aac && echo "--enable-libfdk-aac"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists lame && echo "--enable-libmp3lame"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists freetype2 && echo "--enable-libfreetype"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists theora && echo "--enable-libtheora"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists vorbis && echo "--enable-libvorbis"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists vpx && echo "--enable-libvpx"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists x264 && echo "--enable-libx264"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists x265 && echo "--enable-libx265"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists opus && echo "--enable-libopus"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell pkg-config --exists xvid && echo "--enable-libxvid"))
	@echo "FFmpeg configuration: $(FFMPEG_CONFIG)"
