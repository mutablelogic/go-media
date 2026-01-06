# Paths to packages
GO=$(shell which go)
DOCKER=$(shell which docker)
PKG_CONFIG=$(shell which pkg-config)

# Source version
#FFMPEG_VERSION ?= ffmpeg-7.1.1
#SYS_VERSION ?= ffmpeg71
FFMPEG_VERSION ?= ffmpeg-8.0.1
SYS_VERSION ?= ffmpeg80
CHROMAPRINT_VERSION ?= chromaprint-1.5.1

# CGO configuration - set CGO vars for C++ libraries
CGO_ENV=PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries"

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
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))
BUILD_TAG := ${DOCKER_REGISTRY}/go-media-${OS}-${ARCH}:${VERSION}
PREFIX ?= ${BUILD_DIR}/install

###############################################################################
# TARGETS

.PHONY: all
all: clean cmd

.PHONY: cmd
cmd: ffmpeg chromaprint $(CMD_DIR)

$(CMD_DIR): go-dep go-tidy mkdir 
	@echo Build cmd $(notdir $@)
	@${CGO_ENV} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

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
ffmpeg-configure: mkdir pkconfig-dep ${BUILD_DIR}/${FFMPEG_VERSION} ffmpeg-dep
	@echo "Configuring ${FFMPEG_VERSION} => ${PREFIX}"	
	@cd ${BUILD_DIR}/${FFMPEG_VERSION} && ./configure \
		--disable-doc --disable-programs \
		--prefix="$(shell realpath ${PREFIX})" \
		--enable-static --pkg-config="${PKG_CONFIG}" --pkg-config-flags="--static" --extra-libs="-lpthread" \
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

# Download chromaprint sources
${BUILD_DIR}/${CHROMAPRINT_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/$(CHROMAPRINT_VERSION)" ]; then \
		echo "Downloading $(CHROMAPRINT_VERSION)"; \
		mkdir -p $(BUILD_DIR)/${CHROMAPRINT_VERSION}; \
		curl -L -o $(BUILD_DIR)/chromaprint.tar.gz https://github.com/acoustid/chromaprint/releases/download/v1.5.1/$(CHROMAPRINT_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/chromaprint.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/chromaprint.tar.gz; \
	fi


# Configure chromaprint
# Note: FFmpeg 8.0 removed avfft API, so we use vDSP on macOS or kissfft on other platforms
# kissfft is bundled with chromaprint and requires no external dependencies
ifeq ($(shell uname -s),Darwin)
    FFT_LIB := vdsp
else
    FFT_LIB := kissfft
endif

.PHONY: chromaprint-configure
chromaprint-configure: mkdir ${BUILD_DIR}/${CHROMAPRINT_VERSION} ffmpeg
	@echo "Configuring ${CHROMAPRINT_VERSION} => ${PREFIX} (FFT_LIB=$(FFT_LIB))"
	FFMPEG_DIR="$(shell realpath ${PREFIX})" cmake \
		-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=0 \
		-DBUILD_TESTS=0 \
		-DBUILD_TOOLS=0 \
		-DFFT_LIB=$(FFT_LIB) \
		-DFFMPEG_ROOT="$(shell realpath ${PREFIX})" \
		-DCMAKE_PREFIX_PATH="$(shell realpath ${PREFIX})" \
		-DCMAKE_LIBRARY_PATH="$(shell realpath ${PREFIX})/lib" \
		-DCMAKE_INCLUDE_PATH="$(shell realpath ${PREFIX})/include" \
		--install-prefix "$(shell realpath ${PREFIX})" \
		-S ${BUILD_DIR}/${CHROMAPRINT_VERSION} \
		-B ${BUILD_DIR}

# Build chromaprint
.PHONY: chromaprint-build
chromaprint-build: chromaprint-configure
	@echo "Building ${CHROMAPRINT_VERSION}"
	@cd $(BUILD_DIR) && make -j2

# Install chromaprint
# Create a modified pkg-config file that ensures correct linking order for C++
.PHONY: chromaprint
chromaprint: chromaprint-build
	@echo "Installing ${CHROMAPRINT_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR) && make install
	@sed -i.bak 's/Libs: -L\${libdir} -lchromaprint/Libs: -L\${libdir} -lchromaprint -lstdc++ -lavutil/g' "${PREFIX}/lib/pkgconfig/libchromaprint.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libchromaprint.pc.bak"

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
	@echo ... test pkg/file
	@${GO} test ./pkg/file

.PHONY: test
test-chromaprint:
	@echo ... test pkg/chromaprint
	@${CGO_ENV} ${GO} test ./pkg/segmenter
	@${CGO_ENV} ${GO} test ./pkg/chromaprint

.PHONY: test-sys
test-sys: go-dep go-tidy ffmpeg-build
	@echo Test
	@echo ... test sys/${SYS_VERSION}
	@${CGO_ENV} ${GO} test ./sys/${SYS_VERSION}

.PHONY: test-ffmpeg
test-ffmpeg: test-sys
	@echo Test
	@echo ... test pkg/ffmpeg/...
	@${CGO_ENV} ${GO} test ./pkg/ffmpeg/...


###############################################################################
# DEPENDENCIES, ETC

.PHONY: go-dep
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

.PHONY: docker-dep
docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

.PHONY: pkconfig-dep
pkconfig-dep:
	@test -f "${PKG_CONFIG}" && test -x "${PKG_CONFIG}"  || (echo "Missing pkg-config binary" && exit 1)


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
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libass && echo "--enable-libass"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists fdk-aac && echo "--enable-libfdk-aac"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists lame && echo "--enable-libmp3lame"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists freetype2 && echo "--enable-libfreetype"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vorbis && echo "--enable-libvorbis"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists opus && echo "--enable-libopus"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists x264 && echo "--enable-libx264"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists x265 && echo "--enable-libx265"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists xvid && echo "--enable-libxvid"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vpx && echo "--enable-libvpx"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libgcrypt && echo "--enable-gcrypt"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists aom && echo "--enable-libaom"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libbluray && echo "--enable-libbluray"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists dav1d && echo "--enable-libdav1d"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vulkan && echo "--enable-vulkan"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists zvbi-0.2 && echo "--enable-libzvbi"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists soxr && echo "--enable-libsoxr"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libopenjp2 && echo "--enable-libopenjpeg"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists rav1e && echo "--enable-librav1e"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists SvtAv1Enc && echo "--enable-libsvtav1"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists srt && echo "--enable-libsrt"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libwebp && echo "--enable-libwebp"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists zimg && echo "--enable-libzimg"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vidstab && echo "--enable-libvidstab"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libvmaf && echo "--enable-libvmaf"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists openh264 && echo "--enable-libopenh264"))
	@echo "FFmpeg configuration: $(FFMPEG_CONFIG)"
