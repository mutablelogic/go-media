# Paths to packages
GO=$(shell which go)
PKG_CONFIG=$(shell which pkg-config)
CURL=$(shell which curl)

# Build with the modules go.mod requires, rather than a go.work workspace's,
# so go mod tidy doesn't change go.mod to match local checkouts. Set
# GOWORK to a go.work file to use a workspace
GOWORK ?= off
export GOWORK

# Curl retry policy for flaky downloads
CURL_RETRY_FLAGS = --fail --location --retry 5 --retry-delay 10 --retry-all-errors

# Default parallelism
JOBS ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 1)

# Locations
BUILD_DIR ?= build
CMD_DIR := $(wildcard cmd/*)
PREFIX ?= ${BUILD_DIR}/install

# Source version
# libfvad's only release, v1.0, predates its CMake build, so use a commit
FFMPEG_VERSION ?= ffmpeg-8.1.3
SYS_VERSION ?= ffmpeg80
CHROMAPRINT_VERSION ?= chromaprint-1.5.1
LIBEXIF_VERSION ?= 0.6.26
LIBRAW_VERSION ?= 0.22.1
LIBHEIF_VERSION ?= 1.23.1
LIBFVAD_VERSION ?= 532ab666c20d3cfda38bca63abbb0f152706c369

# A library is built once for each version: installing it writes its stamp
# in the installation, and removes the stamps of its other versions, so it
# isn't configured or built again while the stamp exists, such as when the
# installation is restored from a cache. Remove the stamp, or the
# installation, to build it again.
STAMP = ${PREFIX}/.stamp
FFMPEG_STAMP = ${STAMP}-ffmpeg-$(FFMPEG_VERSION:ffmpeg-%=%)$(if $(filter lgpl,$(FFMPEG_LICENSE)),-lgpl)
CHROMAPRINT_STAMP = ${STAMP}-chromaprint-$(CHROMAPRINT_VERSION:chromaprint-%=%)
LIBRAW_STAMP = ${STAMP}-libraw-${LIBRAW_VERSION}
LIBEXIF_STAMP = ${STAMP}-libexif-${LIBEXIF_VERSION}
LIBHEIF_STAMP = ${STAMP}-libheif-${LIBHEIF_VERSION}
LIBFVAD_STAMP = ${STAMP}-libfvad-${LIBFVAD_VERSION}

# stamp writes the stamp of the library which is installed, and removes the
# stamps of its other versions
define stamp
	@rm -f ${STAMP}-$(1)-*
	@touch $@
endef

# Set OS and Architecture (must be before CGO configuration)
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
# Windows sets OS=Windows_NT in the environment, which ?= would keep, so
# replace it
ifeq ($(OS),Windows_NT)
OS := $(shell uname | tr A-Z a-z | sed -E 's/^(msys|mingw|cygwin).*/windows/')
else
OS ?= $(shell uname | tr A-Z a-z | sed -E 's/^(msys|mingw|cygwin).*/windows/')
endif
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')

# CGO configuration - set CGO vars for C++ libraries
ifeq ($(OS),darwin)
# Homebrew dependencies whose lib or include dirs aren't found otherwise. Use
# each keg's own dir, not the general Homebrew prefix, which can hold a
# Homebrew-installed ffmpeg that would be used instead of this one.
BREW ?= $(shell command -v brew 2>/dev/null)
BREW_PREFIX = $(if $(BREW),$(shell $(BREW) --prefix 2>/dev/null))
brew_dirs = $(if $(BREW),$(wildcard $(foreach f,$(1),$(shell $(BREW) --prefix $(f) 2>/dev/null)/$(2))))
EXTRA_LIB_DIRS=$(call brew_dirs,gettext libpng lame,lib)
EXTRA_INCLUDE_DIRS=$(call brew_dirs,lame librsvg,include)
CGO_ENV=PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries $(foreach d,${EXTRA_LIB_DIRS},-L${d})"
FFMPEG_EXTRA_LDFLAGS=--extra-ldflags="$(foreach d,${EXTRA_LIB_DIRS},-L${d})" --extra-cflags="$(foreach d,${EXTRA_INCLUDE_DIRS},-I${d})"
else
CGO_ENV=PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig" CGO_LDFLAGS_ALLOW="-(W|D).*" CGO_LDFLAGS="-lstdc++"
FFMPEG_EXTRA_LDFLAGS=
endif

# On Windows, the Video for Windows capture device (vfwcap) isn't built, as it
# links avicap32.dll and msvfw32.dll, which Windows Server Core doesn't have,
# so an executable linked with ffmpeg can't start in a Server Core container.
# The DirectShow (dshow) and screen (gdigrab) capture devices are still built.
ifeq ($(OS),windows)
FFMPEG_PLATFORM_CONFIG=--disable-indev=vfwcap
endif

# FFMPEG_LICENSE is the license FFmpeg is built under. With gpl (the
# default), it's a GPL version 3 build, which can be redistributed under the
# GPL: GPL, for x264 and x265, and version 3, for OpenSSL 3 and the codecs
# which are version 3. With lgpl, it's an LGPL version 3 build, without the
# libraries which FFmpeg only allows in a GPL build (FFMPEG_GPL_LIBRARIES),
# so x264 and x265 aren't built, and nor is any other encoder they'd provide.
# FFMPEG_NONFREE=1 adds the nonfree components, fdk-aac and libtls, which
# makes a build which FFmpeg reports as "nonfree and unredistributable", so
# is only for your own use.
FFMPEG_LICENSE ?= gpl
FFMPEG_NONFREE ?=
ifeq ($(FFMPEG_LICENSE),gpl)
FFMPEG_LICENSE_CONFIG := --enable-gpl --enable-version3 $(if $(FFMPEG_NONFREE),--enable-nonfree)
else ifeq ($(FFMPEG_LICENSE),lgpl)
FFMPEG_LICENSE_CONFIG := --enable-version3 $(if $(FFMPEG_NONFREE),--enable-nonfree)
else
$(error FFMPEG_LICENSE must be gpl or lgpl, not "$(FFMPEG_LICENSE)")
endif

# gpl_library returns an FFmpeg option, only for a GPL build, for a library
# FFmpeg only allows in one (its EXTERNAL_LIBRARY_GPL_LIST)
gpl_library = $(if $(filter gpl,$(FFMPEG_LICENSE)),$(1))


# Set build flags
VERSION_PKG = github.com/mutablelogic/go-server/pkg/version
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}"

###############################################################################
# TARGETS

.PHONY: all
all: clean cmd

.PHONY: cmd
cmd: ffmpeg chromaprint libexif libraw libheif libfvad $(CMD_DIR)

$(CMD_DIR): go-dep go-tidy chromaprint-dep mkdir
	@echo Build cmd $(notdir $@)
	@${CGO_ENV} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@)$(if $(filter windows,$(OS)),.exe) ./$@

###############################################################################
# FFMPEG

# Download ffmpeg sources
${BUILD_DIR}/${FFMPEG_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/$(FFMPEG_VERSION)" ]; then \
		echo "Downloading $(FFMPEG_VERSION)"; \
		$(CURL) $(CURL_RETRY_FLAGS) -o "$(BUILD_DIR)/ffmpeg.tar.gz" https://ffmpeg.org/releases/$(FFMPEG_VERSION).tar.gz && \
		tar -xzf "$(BUILD_DIR)/ffmpeg.tar.gz" -C "$(BUILD_DIR)" && \
		rm -f "$(BUILD_DIR)/ffmpeg.tar.gz"; \
	fi

# Configure ffmpeg
.PHONY: ffmpeg-configure
ffmpeg-configure: mkdir pkconfig-dep ${BUILD_DIR}/${FFMPEG_VERSION} ffmpeg-dep
	@echo "Configuring ${FFMPEG_VERSION} => ${PREFIX}"	
	@cd ${BUILD_DIR}/${FFMPEG_VERSION} && ./configure \
		--disable-doc --disable-programs \
		--prefix="$(shell realpath ${PREFIX})" \
		--enable-static --pkg-config="${PKG_CONFIG}" --pkg-config-flags="--static" --extra-libs="-lpthread" ${FFMPEG_EXTRA_LDFLAGS} \
		${FFMPEG_LICENSE_CONFIG} ${FFMPEG_PLATFORM_CONFIG} ${FFMPEG_CONFIG}

# Build ffmpeg
.PHONY: ffmpeg-build
ffmpeg-build: ffmpeg-configure
	@echo "Building ${FFMPEG_VERSION} with ${JOBS} jobs"
	@cd $(BUILD_DIR)/$(FFMPEG_VERSION) && make -j$(JOBS)

# Install ffmpeg
.PHONY: ffmpeg-install
ffmpeg-install: ffmpeg-build
	@echo "Installing ${FFMPEG_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR)/$(FFMPEG_VERSION) && make install

# Install ffmpeg, unless it's installed already
.PHONY: ffmpeg
ffmpeg: ${FFMPEG_STAMP}
${FFMPEG_STAMP}:
	@$(MAKE) --no-print-directory ffmpeg-install
	$(call stamp,ffmpeg)

###############################################################################
# CHROMAPRINT

# Download chromaprint sources
${BUILD_DIR}/${CHROMAPRINT_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/$(CHROMAPRINT_VERSION)" ]; then \
		echo "Downloading $(CHROMAPRINT_VERSION)"; \
		mkdir -p $(BUILD_DIR)/${CHROMAPRINT_VERSION}; \
		$(CURL) $(CURL_RETRY_FLAGS) -o $(BUILD_DIR)/chromaprint.tar.gz https://github.com/acoustid/chromaprint/releases/download/v1.5.1/$(CHROMAPRINT_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/chromaprint.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/chromaprint.tar.gz; \
	fi

# Configure chromaprint. FFmpeg 8.0 removed avfft, so use vDSP on macOS, and
# chromaprint's bundled kissfft elsewhere
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

# Build chromaprint, with cmake --build as cmake may generate Ninja files
# rather than a Makefile (eg. MSYS2)
.PHONY: chromaprint-build
chromaprint-build: chromaprint-configure
	@echo "Building ${CHROMAPRINT_VERSION} with ${JOBS} jobs"
	@cmake --build ${BUILD_DIR} -j$(JOBS)

# Install chromaprint, adding C++ and libavutil to its pkg-config libs
.PHONY: chromaprint-install
chromaprint-install: chromaprint-build
	@echo "Installing ${CHROMAPRINT_VERSION} => ${PREFIX}"
	@cmake --install ${BUILD_DIR}
	@sed -i.bak 's/Libs: -L\${libdir} -lchromaprint/Libs: -L\${libdir} -lchromaprint -lstdc++ -lavutil/g' "${PREFIX}/lib/pkgconfig/libchromaprint.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libchromaprint.pc.bak"

# Install chromaprint, unless it's installed already, and again when ffmpeg,
# which it's built with, is installed after it
.PHONY: chromaprint
chromaprint: ${CHROMAPRINT_STAMP}
${CHROMAPRINT_STAMP}: ${FFMPEG_STAMP}
	@$(MAKE) --no-print-directory chromaprint-install
	$(call stamp,chromaprint)

###############################################################################
# LIBRAW

# Download libraw sources
${BUILD_DIR}/libraw-${LIBRAW_VERSION}:
	if [ ! -d "$(BUILD_DIR)/libraw-$(LIBRAW_VERSION)" ]; then \
		echo "Downloading $(LIBRAW_VERSION)"; \
		$(CURL) $(CURL_RETRY_FLAGS) -o $(BUILD_DIR)/libraw.tar.gz https://www.libraw.org/data/LibRaw-${LIBRAW_VERSION}.tar.gz; \
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

# Install libraw, appending -lz (DNG deflate) and -lm to its pkg-config libs,
# and -lws2_32 on Windows (for htonl/ntohl)
.PHONY: libraw-install
libraw-install: libraw-build
	@echo "Installing ${LIBRAW_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR)/libraw-$(LIBRAW_VERSION) && make install
	@sed -i.bak '/^Libs:/ s/$$/ -lz -lm$(if $(filter windows,${OS}), -lws2_32)/' "${PREFIX}/lib/pkgconfig/libraw.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libraw.pc.bak"
	@${GO} clean -cache

# Install libraw, unless it's installed already
.PHONY: libraw
libraw: ${LIBRAW_STAMP}
${LIBRAW_STAMP}:
	@$(MAKE) --no-print-directory libraw-install
	$(call stamp,libraw)

###############################################################################
# LIBEXIF

# Download libexif sources
${BUILD_DIR}/libexif-${LIBEXIF_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/libexif-$(LIBEXIF_VERSION)" ]; then \
		echo "Downloading $(LIBEXIF_VERSION)"; \
		mkdir -p $(BUILD_DIR)/libexif-${LIBEXIF_VERSION}; \
		$(CURL) $(CURL_RETRY_FLAGS) -o $(BUILD_DIR)/libexif.tar.gz https://github.com/libexif/libexif/releases/download/v$(LIBEXIF_VERSION)/libexif-$(LIBEXIF_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/libexif.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/libexif.tar.gz; \
	fi

.PHONY: libexif-configure
# --disable-nls: translated strings aren't needed, and would require -lintl
# where gettext is found (eg. MSYS2)
libexif-configure: mkdir ${BUILD_DIR}/libexif-${LIBEXIF_VERSION}
	@echo "Configuring libexif-${LIBEXIF_VERSION} => ${PREFIX}"
	@cd ${BUILD_DIR}/libexif-${LIBEXIF_VERSION} && ./configure \
		--disable-docs --disable-nls --enable-year2038  \
		--prefix="$(shell realpath ${PREFIX})" \
		--enable-static --disable-shared

# Build libexif
.PHONY: libexif-build
libexif-build: libexif-configure
	@echo "Building libexif-${LIBEXIF_VERSION} with ${JOBS} jobs"
	@cd $(BUILD_DIR)/libexif-$(LIBEXIF_VERSION) && make -j$(JOBS)

# Install libexif
.PHONY: libexif-install
libexif-install: libexif-build
	@echo "Installing libexif-${LIBEXIF_VERSION} => ${PREFIX}"
	@cd $(BUILD_DIR)/libexif-$(LIBEXIF_VERSION) && make install
	@sed -i.bak 's|-lexif$$|-lexif -lm|' "${PREFIX}/lib/pkgconfig/libexif.pc"
	@rm -f "${PREFIX}/lib/pkgconfig/libexif.pc.bak"

# Install libexif, unless it's installed already
.PHONY: libexif
libexif: ${LIBEXIF_STAMP}
${LIBEXIF_STAMP}:
	@$(MAKE) --no-print-directory libexif-install
	$(call stamp,libexif)

###############################################################################
# LIBHEIF

# Download libheif sources
${BUILD_DIR}/libheif-${LIBHEIF_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/libheif-$(LIBHEIF_VERSION)" ]; then \
		echo "Downloading $(LIBHEIF_VERSION)"; \
		mkdir -p $(BUILD_DIR)/libheif-${LIBHEIF_VERSION}; \
		$(CURL) $(CURL_RETRY_FLAGS) -o $(BUILD_DIR)/libheif.tar.gz https://github.com/strukturag/libheif/releases/download/v$(LIBHEIF_VERSION)/libheif-$(LIBHEIF_VERSION).tar.gz; \
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
.PHONY: libheif-install
libheif-install: libheif-build
	@echo "Installing libheif-${LIBHEIF_VERSION} => ${PREFIX}"
	@cmake --install ${BUILD_DIR}/libheif-${LIBHEIF_VERSION}
	@if ! grep -q 'libavcodec' "${PREFIX}/lib/pkgconfig/libheif.pc"; then \
		sed -i.bak 's/^Requires.private:[[:space:]]*/Requires.private: libavcodec libavformat libavutil /' "${PREFIX}/lib/pkgconfig/libheif.pc"; \
		rm -f "${PREFIX}/lib/pkgconfig/libheif.pc.bak"; \
	fi

# Install libheif, unless it's installed already, and again when ffmpeg,
# which it's built with, is installed after it
.PHONY: libheif
libheif: ${LIBHEIF_STAMP}
${LIBHEIF_STAMP}: ${FFMPEG_STAMP}
	@$(MAKE) --no-print-directory libheif-install
	$(call stamp,libheif)

###############################################################################
# LIBFVAD

# Download libfvad sources
${BUILD_DIR}/libfvad-${LIBFVAD_VERSION}:
	@if [ ! -d "$(BUILD_DIR)/libfvad-$(LIBFVAD_VERSION)" ]; then \
		echo "Downloading libfvad-$(LIBFVAD_VERSION)"; \
		$(CURL) $(CURL_RETRY_FLAGS) -o $(BUILD_DIR)/libfvad.tar.gz https://github.com/dpirch/libfvad/archive/$(LIBFVAD_VERSION).tar.gz; \
		tar -xzf $(BUILD_DIR)/libfvad.tar.gz -C $(BUILD_DIR); \
		rm -f $(BUILD_DIR)/libfvad.tar.gz; \
	fi

.PHONY: libfvad-configure
libfvad-configure: mkdir ${BUILD_DIR}/libfvad-${LIBFVAD_VERSION}
	@echo "Configuring libfvad-${LIBFVAD_VERSION} => ${PREFIX}"
	@cmake \
		-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
		-DCMAKE_BUILD_TYPE=Release \
		-DCMAKE_INSTALL_PREFIX="$(shell realpath ${PREFIX})" \
		-S ${BUILD_DIR}/libfvad-${LIBFVAD_VERSION} \
		-B ${BUILD_DIR}/libfvad-${LIBFVAD_VERSION}/build

# Build libfvad
.PHONY: libfvad-build
libfvad-build: libfvad-configure
	@echo "Building libfvad-${LIBFVAD_VERSION} with ${JOBS} jobs"
	@cmake --build ${BUILD_DIR}/libfvad-${LIBFVAD_VERSION}/build -j$(JOBS)

# Install libfvad, and write its pkg-config file, which only its autotools
# build installs
.PHONY: libfvad-install
libfvad-install: libfvad-build
	@echo "Installing libfvad-${LIBFVAD_VERSION} => ${PREFIX}"
	@cmake --install ${BUILD_DIR}/libfvad-${LIBFVAD_VERSION}/build
	@install -d "${PREFIX}/lib/pkgconfig"
	@printf '%s\n' \
		'prefix=$(shell realpath ${PREFIX})' \
		'libdir=$${prefix}/lib' \
		'includedir=$${prefix}/include' \
		'' \
		'Name: libfvad' \
		'Description: Voice activity detection (VAD) library' \
		'URL: https://github.com/dpirch/libfvad' \
		'Version: 1.0.1' \
		'Libs: -L$${libdir} -lfvad' \
		'Cflags: -I$${includedir}' \
		> "${PREFIX}/lib/pkgconfig/libfvad.pc"

# Install libfvad, unless it's installed already
.PHONY: libfvad
libfvad: ${LIBFVAD_STAMP}
${LIBFVAD_STAMP}:
	@$(MAKE) --no-print-directory libfvad-install
	$(call stamp,libfvad)

###############################################################################
# MUSAN

# MUSAN is a corpus of speech, music and noise (https://www.openslr.org/17/),
# which test-musan uses to show how well the segmenter classifies audio. It's
# an 11GB download, which make clean removes along with the rest of the build.
# MUSAN_LIMIT is the number of files for each label, or zero for all of them
MUSAN_DIR ?= ${BUILD_DIR}/musan
MUSAN_LIMIT ?= 50
MUSAN_STAMP = ${MUSAN_DIR}/.stamp

# Download and extract MUSAN, resuming a download which was interrupted
${MUSAN_STAMP}:
	@echo "Downloading MUSAN => ${MUSAN_DIR}"
	@install -d ${MUSAN_DIR}
	@${CURL} ${CURL_RETRY_FLAGS} -C - -o ${MUSAN_DIR}.tar.gz https://www.openslr.org/resources/17/musan.tar.gz
	@echo "Extracting MUSAN => ${MUSAN_DIR}"
	@tar -xzf ${MUSAN_DIR}.tar.gz -C ${MUSAN_DIR} --strip-components=1
	@rm -f ${MUSAN_DIR}.tar.gz
	@touch $@

.PHONY: musan
musan: ${MUSAN_STAMP}

###############################################################################
# TESTS

.PHONY: test
test: ffmpeg chromaprint libexif libraw libheif libfvad test-ffmpeg test-chromaprint test-exif test-raw test-heif test-libfvad test-gomedia

.PHONY: test-chromaprint
test-chromaprint: chromaprint
	@echo ... test sys/chromaprint
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/chromaprint

.PHONY: test-libfvad
# segmenter depends on ffmpeg and libfvad, so is tested here
test-libfvad: ffmpeg libfvad
	@echo ... test sys/libfvad segmenter
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libfvad
	@${CGO_ENV} ${GO} test ${ARGS} ./segmenter/...

.PHONY: test-musan
# Classify the MUSAN corpus, writing the stats of each segment to a CSV file
test-musan: ffmpeg libfvad musan
	@echo ... classify ${MUSAN_DIR}
	@${CGO_ENV} ${GO} run ./cmd/segmenter classify --limit ${MUSAN_LIMIT} --csv ${MUSAN_DIR}.csv \
		${MUSAN_DIR}/speech ${MUSAN_DIR}/music ${MUSAN_DIR}/noise

.PHONY: test-exif
test-exif: libexif
	@echo ... test sys/libexif
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libexif

.PHONY: test-raw
test-raw: libraw
	@echo ... test sys/libraw
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libraw

.PHONY: test-heif
test-heif: libheif
	@echo ... test sys/libheif
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/libheif

.PHONY: test-ffmpeg
test-ffmpeg: go-dep go-tidy ffmpeg
	@echo ... test sys/${SYS_VERSION} reader writer frame
	@${CGO_ENV} ${GO} test ${ARGS} ./sys/${SYS_VERSION}
	@${CGO_ENV} ${GO} test ${ARGS} ./reader/...
	@${CGO_ENV} ${GO} test ${ARGS} ./writer/...
	@${CGO_ENV} ${GO} test ${ARGS} ./frame/...

.PHONY: test-gomedia
test-gomedia: ffmpeg
	@echo ... test task profile gomedia
	@${CGO_ENV} ${GO} test ${ARGS} ./task/...
	@${CGO_ENV} ${GO} test ${ARGS} ./profile/...
	@${CGO_ENV} ${GO} test ${ARGS} ./gomedia/...

###############################################################################
# LIBRARIES ZIP

# Package the libraries go-media links with, as installed in PREFIX with their
# headers, pkg-config files and stamps, as a zip which can be unpacked
# anywhere, as the pkg-config files are made relative to where they are, with
# ${pcfiledir}. Unpacked into build/install, make doesn't build them again.
# The install path is replaced in the form it's in and, under MSYS2, the form
# Windows tools such as cmake write
LIBRARIES_ZIP = ${BUILD_DIR}/go-media-${VERSION}-${FFMPEG_LICENSE}-${OS}-${ARCH}.zip

.PHONY: libraries-zip
libraries-zip: ffmpeg chromaprint libexif libraw libheif libfvad
	@echo "Packaging ${PREFIX} => ${LIBRARIES_ZIP}"
	@rm -rf "${BUILD_DIR}/zip" "${LIBRARIES_ZIP}" && install -d "${BUILD_DIR}/zip"
	@cp -R "${PREFIX}/." "${BUILD_DIR}/zip/"
	@prefix="$$(realpath "${PREFIX}")"; windows="$$(cygpath -m "$$prefix" 2>/dev/null || true)"; \
	for f in "${BUILD_DIR}"/zip/lib/pkgconfig/*.pc; do \
		sed -i.bak -e "s|$$prefix|\$${pcfiledir}/../..|g" $${windows:+-e "s|$$windows|\$${pcfiledir}/../..|g"} "$$f" && rm -f "$$f.bak"; \
	done
	@install -d "${BUILD_DIR}/zip/share/doc/go-media" && cp LICENSE NOTICE etc/licenses/GPL-3.0.txt "${BUILD_DIR}/zip/share/doc/go-media/"
	@cd "${BUILD_DIR}/zip" && zip -qr "$(abspath ${LIBRARIES_ZIP})" .
	@rm -rf "${BUILD_DIR}/zip"

###############################################################################
# DOCKER

# Build Linux images for this machine's architecture with etc/docker/Dockerfile.linux:
# the gomedia image (make docker), and the images of the libraries go-media
# links with, built under FFMPEG_LICENSE, to build with (make docker-dev) and
# to run with (make docker-runtime)
DOCKER ?= $(shell which docker)
DOCKER_REPO ?= ghcr.io/mutablelogic/gomedia
DOCKER_LIBRARIES_REPO ?= ghcr.io/mutablelogic/go-media
DOCKER_SPDX = Apache-2.0 AND $(if $(filter lgpl,$(FFMPEG_LICENSE)),LGPL,GPL)-3.0-or-later
docker_build = ${DOCKER} build --file etc/docker/Dockerfile.linux --platform linux/${ARCH} \
	--build-arg ARCH=${ARCH} --build-arg LICENSE=${FFMPEG_LICENSE} --build-arg "SPDX=${DOCKER_SPDX}" \
	--target $(1) --tag $(2) .

.PHONY: docker
docker: docker-dep
	@echo "Building ${DOCKER_REPO}:${VERSION}-linux-${ARCH}"
	@$(call docker_build,gomedia,${DOCKER_REPO}:${VERSION}-linux-${ARCH})

.PHONY: docker-dev
docker-dev: docker-dep
	@echo "Building ${DOCKER_LIBRARIES_REPO}:${VERSION}-${FFMPEG_LICENSE}-dev"
	@$(call docker_build,dev,${DOCKER_LIBRARIES_REPO}:${VERSION}-${FFMPEG_LICENSE}-dev)

.PHONY: docker-runtime
docker-runtime: docker-dep
	@echo "Building ${DOCKER_LIBRARIES_REPO}:${VERSION}-${FFMPEG_LICENSE}-runtime"
	@$(call docker_build,runtime,${DOCKER_LIBRARIES_REPO}:${VERSION}-${FFMPEG_LICENSE}-runtime)

###############################################################################
# DEPENDENCIES, ETC

.PHONY: docker-dep
docker-dep:
	@test -f "$(DOCKER)" && test -x "$(DOCKER)" || (echo "Missing docker binary" && exit 1)

.PHONY: go-dep
go-dep:
	@test -f "$(GO)" && test -x "$(GO)"  || (echo "Missing go binary" && exit 1)

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
	@install -d ${BUILD_DIR}
	@echo 'module buildartifacts' > ${BUILD_DIR}/go.mod
	@${GO} mod tidy

.PHONY: clean
clean: go-tidy
	@echo Clean
	@rm -fr $(BUILD_DIR)

# Check for FFmpeg dependencies
.PHONY: ffmpeg-dep
ffmpeg-dep:
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libass && echo "--enable-libass"))
	$(if $(FFMPEG_NONFREE),$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists fdk-aac && echo "--enable-libfdk-aac")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists lame && echo "--enable-libmp3lame"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists freetype2 && echo "--enable-libfreetype"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vorbis && echo "--enable-libvorbis"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists opus && echo "--enable-libopus"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists x264 && echo "--enable-libx264")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists x265 && echo "--enable-libx265")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists xvid && echo "--enable-libxvid")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vpx && echo "--enable-libvpx"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libgcrypt && echo "--enable-gcrypt"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists aom && echo "--enable-libaom"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libbluray && echo "--enable-libbluray"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists dav1d && echo "--enable-libdav1d"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vulkan && echo "--enable-vulkan"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists zvbi-0.2 && echo "--enable-libzvbi"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists soxr && echo "--enable-libsoxr"))
# Not with Homebrew's openjpeg, which fails to link statically
	$(if $(and $(BREW_PREFIX),$(filter $(BREW_PREFIX)/%,$(shell ${PKG_CONFIG} --variable=libdir libopenjp2 2>/dev/null))),,$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libopenjp2 && echo "--enable-libopenjpeg")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists rav1e && echo "--enable-librav1e"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists SvtAv1Enc && echo "--enable-libsvtav1"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists srt && echo "--enable-libsrt"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libwebp && echo "--enable-libwebp"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists zimg && echo "--enable-libzimg"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists vidstab && echo "--enable-libvidstab")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libvmaf && echo "--enable-libvmaf"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists openh264 && echo "--enable-libopenh264"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists gsm && echo "--enable-libgsm"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists speex && echo "--enable-libspeex"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists twolame && echo "--enable-libtwolame"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists theora && echo "--enable-libtheora"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists shine && echo "--enable-libshine"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libilbc && echo "--enable-libilbc"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists codec2 && echo "--enable-libcodec2"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists lc3 && echo "--enable-liblc3"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libmysofa && echo "--enable-libmysofa"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libgme && echo "--enable-libgme"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libmodplug && echo "--enable-libmodplug"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libopenmpt && echo "--enable-libopenmpt"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists opencore-amrnb && echo "--enable-libopencore-amrnb"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists opencore-amrwb && echo "--enable-libopencore-amrwb"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vo-amrwbenc && echo "--enable-libvo-amrwbenc"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists rubberband && echo "--enable-librubberband")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libbs2b && echo "--enable-libbs2b"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists kvazaar && echo "--enable-libkvazaar"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists davs2 && echo "--enable-libdavs2")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists uavs3d && echo "--enable-libuavs3d"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(call gpl_library,$(shell ${PKG_CONFIG} --exists xavs2 && echo "--enable-libxavs2")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vvenc && echo "--enable-libvvenc"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists xeve && echo "--enable-libxeve"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists xevd && echo "--enable-libxevd"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists oapv && echo "--enable-liboapv"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libjxl && echo "--enable-libjxl"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists fontconfig && echo "--enable-libfontconfig"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists fribidi && echo "--enable-libfribidi"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists harfbuzz && echo "--enable-libharfbuzz"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists aribb24 && echo "--enable-libaribb24"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libaribcaption && echo "--enable-libaribcaption"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists caca && echo "--enable-libcaca"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists librsvg-2.0 && echo "--enable-librsvg"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists lcms2 && echo "--enable-lcms2"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists lensfun && echo "--enable-liblensfun"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists tesseract && echo "--enable-libtesseract"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libqrencode && echo "--enable-libqrencode"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libplacebo && echo "--enable-libplacebo"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists openssl && echo "--enable-openssl"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists mbedtls && echo "--enable-mbedtls"))
	$(if $(FFMPEG_NONFREE),$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libtls && echo "--enable-libtls")))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists librtmp && echo "--enable-librtmp"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libssh && echo "--enable-libssh"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists librist && echo "--enable-librist"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists librabbitmq && echo "--enable-librabbitmq"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libzmq && echo "--enable-libzmq"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libxml-2.0 && echo "--enable-libxml2"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists smbclient && echo "--enable-libsmbclient"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists snappy && echo "--enable-libsnappy"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists dvdread && echo "--enable-libdvdread"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists dvdnav && echo "--enable-libdvdnav"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libcdio libcdio_paranoia && echo "--enable-libcdio"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libpulse && echo "--enable-libpulse"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists jack && echo "--enable-libjack"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libv4l2 && echo "--enable-libv4l2"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libdc1394-2 && echo "--enable-libdc1394"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libiec61883 && echo "--enable-libiec61883"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists openal && echo "--enable-openal"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists openvino && echo "--enable-libopenvino"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists opencv4 && echo "--enable-libopencv"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists pocketsphinx && echo "--enable-pocketsphinx"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists vapoursynth && echo "--enable-vapoursynth"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists lilv-0 && echo "--enable-lv2"))
	$(eval FFMPEG_CONFIG := $(FFMPEG_CONFIG) $(shell ${PKG_CONFIG} --exists libklvanc && echo "--enable-libklvanc"))
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

# Check for Chromaprint dependencies
.PHONY: chromaprint-dep
chromaprint-dep:
	$(eval BUILD_FLAGS := $(BUILD_FLAGS) $(shell PKG_CONFIG_PATH="$(shell realpath ${PREFIX})/lib/pkgconfig:$$PKG_CONFIG_PATH" $(PKG_CONFIG) --exists libchromaprint && echo "--tags chromaprint"))
