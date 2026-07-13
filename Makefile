# Paths to packages
GO=$(shell which go)
DOCKER=$(shell which docker)
PKG_CONFIG=$(shell which pkg-config)

# Default parallelism
JOBS ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 1)

# Locations
BUILD_DIR ?= build
CMD_DIR := $(wildcard cmd/*)
PREFIX ?= ${BUILD_DIR}/install

# Source version
FFMPEG_VERSION ?= ffmpeg-8.0.3
SYS_VERSION ?= ffmpeg80
CHROMAPRINT_VERSION ?= chromaprint-1.5.1
LIBEXIF_VERSION ?= 0.6.26
LIBRAW_VERSION ?= 0.22.1
LIBHEIF_VERSION ?= 1.23.1

# Set OS and Architecture (must be before CGO configuration)
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')
DOCKER_REGISTRY ?= ghcr.io/mutablelogic

# CGO configuration - set CGO vars for C++ libraries
ifeq ($(OS),darwin)
CGO_ENV=PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries"
else
CGO_ENV=PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" CGO_LDFLAGS="-lstdc++"
endif


# Set build flags
VERSION_PKG = github.com/mutablelogic/go-server/pkg/version
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}"

# Docker
DOCKER_REPO ?= ghcr.io/mutablelogic/gomedia
DOCKER_SOURCE ?= $(shell cat go.mod | head -1 | cut -d ' ' -f 2)
DOCKER_TAG = ${DOCKER_REPO}:${VERSION}-${OS}-${ARCH}


###############################################################################
# TARGETS

.PHONY: all
all: clean cmd

.PHONY: cmd
cmd: ffmpeg chromaprint libexif libraw libheif $(CMD_DIR)

$(CMD_DIR): go-dep go-tidy sdl-dep chromaprint-dep mkdir
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
	@echo "Building ${FFMPEG_VERSION} with ${JOBS} jobs"
	@cd $(BUILD_DIR)/$(FFMPEG_VERSION) && make -j$(JOBS)

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
	@echo "Building ${CHROMAPRINT_VERSION} with ${JOBS} jobs"
	@cd $(BUILD_DIR) && make -j$(JOBS)

# Install chromaprint
# Create a modified pkg-config file that ensures correct linking order for C++
.PHONY: chromaprint
chromaprint: chromaprint-build
	@echo "Installing ${CHROMAPRINT_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR) && make install
	@sed -i.bak 's/Libs: -L\${libdir} -lchromaprint/Libs: -L\${libdir} -lchromaprint -lstdc++ -lavutil/g' "${PREFIX}/lib/pkgconfig/libchromaprint.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libchromaprint.pc.bak"

###############################################################################
# LIBRAW

# Download libraw sources
${BUILD_DIR}/libraw-${LIBRAW_VERSION}:
	if [ ! -d "$(BUILD_DIR)/libraw-$(LIBRAW_VERSION)" ]; then \
		echo "Downloading $(LIBRAW_VERSION)"; \
		curl -L -o $(BUILD_DIR)/libraw.tar.gz https://www.libraw.org/data/LibRaw-${LIBRAW_VERSION}.tar.gz; \
		tar -xzf $(BUILD_DIR)/libraw.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/libraw.tar.gz; \
		mv $(BUILD_DIR)/LibRaw-${LIBRAW_VERSION} $(BUILD_DIR)/libraw-${LIBRAW_VERSION}; \
	fi

.PHONY: libraw-configure
libraw-configure: mkdir ${BUILD_DIR}/libraw-${LIBRAW_VERSION}
	@echo "Configuring ${LIBRAW_VERSION} => ${PREFIX}"
	@cd ${BUILD_DIR}/libraw-${LIBRAW_VERSION} && ./configure \
		--prefix="$(shell realpath ${PREFIX})" \
		--enable-static --disable-shared \
		LIBS="$(if $(filter darwin,$(OS)),-lc++,-lstdc++)"

# Build libraw
.PHONY: libraw-build
libraw-build: libraw-configure
	@echo "Building libraw-${LIBRAW_VERSION} with ${JOBS} jobs"
	@cd $(BUILD_DIR)/libraw-$(LIBRAW_VERSION) && make -j$(JOBS) lib/libraw.la lib/libraw_r.la

# Install libraw
# Patch pkg-config to add -lz (required for DNG deflate support) and -lm (math functions)
.PHONY: libraw
libraw: libraw-build
	@echo "Installing ${LIBRAW_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR)/libraw-$(LIBRAW_VERSION) && make install
	@sed -i.bak 's|-lraw -lstdc++|-lraw -lstdc++ -lz -lm|' "${PREFIX}/lib/pkgconfig/libraw.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libraw.pc.bak"
	@${GO} clean -cache

###############################################################################
# LIBEXIF

# Download libexif sources
${BUILD_DIR}/libexif-${LIBEXIF_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/libexif-$(LIBEXIF_VERSION)" ]; then \
		echo "Downloading $(LIBEXIF_VERSION)"; \
		mkdir -p $(BUILD_DIR)/libexif-${LIBEXIF_VERSION}; \
		curl -L -o $(BUILD_DIR)/libexif.tar.gz https://github.com/libexif/libexif/releases/download/v$(LIBEXIF_VERSION)/libexif-$(LIBEXIF_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/libexif.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/libexif.tar.gz; \
	fi

.PHONY: libexif-configure
libexif-configure: mkdir ${BUILD_DIR}/libexif-${LIBEXIF_VERSION}
	@echo "Configuring libexif-${LIBEXIF_VERSION} => ${PREFIX}"
	@cd ${BUILD_DIR}/libexif-${LIBEXIF_VERSION} && ./configure \
		--disable-docs --enable-year2038  \
		--prefix="$(shell realpath ${PREFIX})" \
		--enable-static --disable-shared

# Build libexif
.PHONY: libexif-build
libexif-build: libexif-configure
	@echo "Building libexif-${LIBEXIF_VERSION} with ${JOBS} jobs"
	@cd $(BUILD_DIR)/libexif-$(LIBEXIF_VERSION) && make -j$(JOBS)

# Install libexif
.PHONY: libexif
libexif: libexif-build
	@echo "Installing libexif-${LIBEXIF_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR)/libexif-$(LIBEXIF_VERSION) && make install
	@sed -i.bak 's|-lexif$$|-lexif -lm|' "${PREFIX}/lib/pkgconfig/libexif.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libexif.pc.bak"

###############################################################################
# LIBHEIF

# Download libheif sources
${BUILD_DIR}/libheif-${LIBHEIF_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/libheif-$(LIBHEIF_VERSION)" ]; then \
		echo "Downloading $(LIBHEIF_VERSION)"; \
		mkdir -p $(BUILD_DIR)/libheif-${LIBHEIF_VERSION}; \
		curl -L -o $(BUILD_DIR)/libheif.tar.gz https://github.com/strukturag/libheif/releases/download/v$(LIBHEIF_VERSION)/libheif-$(LIBHEIF_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/libheif.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/libheif.tar.gz; \
	fi

.PHONY: libheif-configure
libheif-configure: mkdir pkconfig-dep ${BUILD_DIR}/libheif-${LIBHEIF_VERSION} ffmpeg libheif-dep
	@echo "Configuring libheif-${LIBHEIF_VERSION} => ${PREFIX}"
	@cmake \
		-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=0 \
		-DCMAKE_INSTALL_PREFIX="$(shell realpath ${PREFIX})" \
		-DCMAKE_PREFIX_PATH="$(shell realpath ${PREFIX})" \
		-DCMAKE_LIBRARY_PATH="$(shell realpath ${PREFIX})/lib" \
		-DCMAKE_INCLUDE_PATH="$(shell realpath ${PREFIX})/include" \
		${LIBHEIF_CONFIG} \
		-S ${BUILD_DIR}/libheif-${LIBHEIF_VERSION} \
		-B ${BUILD_DIR}/libheif-${LIBHEIF_VERSION}

# Build libheif
.PHONY: libheif-build
libheif-build: libheif-configure
	@echo "Building libheif-${LIBHEIF_VERSION} with ${JOBS} jobs"
	@cmake --build ${BUILD_DIR}/libheif-${LIBHEIF_VERSION} -j$(JOBS)

# Install libheif
.PHONY: libheif
libheif: libheif-build
	@echo "Installing libheif-${LIBHEIF_VERSION} => ${PREFIX}"
	@cmake --install ${BUILD_DIR}/libheif-${LIBHEIF_VERSION}
	@if ! grep -q 'libavcodec' "${PREFIX}/lib/pkgconfig/libheif.pc"; then \
		sed -i.bak 's/^Requires.private:[[:space:]]*/Requires.private: libavcodec libavformat libavutil /' "${PREFIX}/lib/pkgconfig/libheif.pc"; \
		rm -f "${PREFIX}/lib/pkgconfig/libheif.pc.bak"; \
	fi

###############################################################################
# DOCKER

# Build the docker image
.PHONY: docker
docker: docker-dep
	@echo build docker image ${DOCKER_TAG} OS=${OS} ARCH=${ARCH} SOURCE=${DOCKER_SOURCE} VERSION=${VERSION}
	@${DOCKER} build \
		--tag ${DOCKER_TAG} \
		--provenance=false \
		--build-arg ARCH=${ARCH} \
		--build-arg OS=${OS} \
		--build-arg SOURCE=${DOCKER_SOURCE} \
		--build-arg VERSION=${VERSION} \
		-f etc/docker/Dockerfile .

# Push docker container
.PHONY: docker-push
docker-push: docker-dep 
	@echo push docker image: ${DOCKER_TAG}
	@${DOCKER} push ${DOCKER_TAG}

# Print out the version
.PHONY: docker-version
docker-version: docker-dep 
	@echo "tag=${VERSION}"

###############################################################################
# TESTS

.PHONY: test
test: ffmpeg chromaprint libexif libraw libheif test-ffmpeg test-chromaprint test-exif test-raw test-heif test-metadata test-gomedia

.PHONY: test-chromaprint
test-chromaprint:
	@echo ... test pkg/segmenter pkg/chromaprint
	@${CGO_ENV} ${GO} test ${ARGS} ./pkg/segmenter
	@${CGO_ENV} ${GO} test ${ARGS} ./pkg/chromaprint

.PHONY: test-exif
test-exif:
	@echo ... test sys/libexif pkg/exif
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libexif
	@${CGO_ENV} ${GO} test ${ARGS} ./pkg/exif

.PHONY: test-raw
test-raw:
	@echo ... test sys/libraw pkg/raw
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libraw
	@${CGO_ENV} ${GO} test ${ARGS} ./pkg/raw

.PHONY: test-heif
test-heif:
	@echo ... test sys/libheif pkg/heif
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libheif
	@${CGO_ENV} ${GO} test ${ARGS} ./pkg/heif

.PHONY: test-ffmpeg
test-ffmpeg: go-dep go-tidy
	@echo ... test sys/${SYS_VERSION} pkg/ffmpeg
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/${SYS_VERSION}
	@${CGO_ENV} ${GO} test ${ARGS} ./pkg/ffmpeg/...

.PHONY: test-metadata
test-metadata: 
	@echo ... test metadata/...
	@${CGO_ENV} ${GO} test ${ARGS} ./metadata/...

.PHONY: test-gomedia
test-gomedia: 
	@echo ... test gomedia/...
	@${CGO_ENV} ${GO} test ${ARGS} ./gomedia/...

###############################################################################
# DEPENDENCIES, ETC

.PHONY: go-dep
go-dep:
	@test -f "$(GO)" && test -x "$(GO)"  || (echo "Missing go binary" && exit 1)

.PHONY: docker-dep
docker-dep:
	@test -f "$(DOCKER)" && test -x "$(DOCKER)"  || (echo "Missing docker binary" && exit 1)

.PHONY: pkconfig-dep
pkconfig-dep:
	@test -f "$(PKG_CONFIG)" && test -x "$(PKG_CONFIG)"  || (echo "Missing pkg-config binary" && exit 1)


.PHONY: mkdir
mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}
	@install -d ${PREFIX}

.PHONY: go-tidy
go-tidy: go-dep
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

# Check for libheif dependencies
.PHONY: libheif-dep
libheif-dep:
	$(eval LIBHEIF_CONFIG := -DENABLE_PLUGIN_LOADING=OFF -DWITH_EXAMPLES=OFF -DBUILD_TESTING=OFF -DBUILD_DOCUMENTATION=OFF)
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_GDK_PIXBUF=OFF)
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_LIBSHARPYUV=OFF)
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_LIBDE265=$(shell ${PKG_CONFIG} --exists libde265 && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_X265=$(shell ${PKG_CONFIG} --exists x265 && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_AOM_DECODER=$(shell ${PKG_CONFIG} --exists aom && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_AOM_ENCODER=$(shell ${PKG_CONFIG} --exists aom && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_DAV1D=$(shell ${PKG_CONFIG} --exists dav1d && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_SvtEnc=$(shell ${PKG_CONFIG} --exists SvtAv1Enc && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_OpenH264_DECODER=$(shell ${PKG_CONFIG} --exists openh264 && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_X264=$(shell ${PKG_CONFIG} --exists x264 && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_FFMPEG_DECODER=$(shell ${PKG_CONFIG} --exists libavcodec libavformat libavutil && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_JPEG_DECODER=$(shell ${PKG_CONFIG} --exists libjpeg && echo ON || echo OFF))
	$(eval LIBHEIF_CONFIG := $(LIBHEIF_CONFIG) -DWITH_JPEG_ENCODER=$(shell ${PKG_CONFIG} --exists libjpeg && echo ON || echo OFF))
	@echo "libheif configuration: $(LIBHEIF_CONFIG)"

# Check for SDL dependencies
.PHONY: sdl-dep
sdl-dep:
	$(eval BUILD_FLAGS := $(BUILD_FLAGS) $(shell $(PKG_CONFIG) --exists sdl2 && echo "--tags sdl2"))

# Check for Chromaprint dependencies
.PHONY: chromaprint-dep
chromaprint-dep:
	$(eval BUILD_FLAGS := $(BUILD_FLAGS) $(shell PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig:$$PKG_CONFIG_PATH" $(PKG_CONFIG) --exists libchromaprint && echo "--tags chromaprint"))
