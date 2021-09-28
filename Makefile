# Paths to packages
GO=$(shell which go)

# Paths to locations, etc
BUILD_DIR := "build"
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))
PLUGIN_DIR := $(wildcard plugin/*)

# Build flags
BUILD_MODULE = "github.com/djthorpe/go-media"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

all: clean test plugins server cmd

cmd: dependencies mkdir $(CMD_DIR)

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} github.com/mutablelogic/go-server/cmd/server

plugins: $(PLUGIN_DIR)
	@echo Build plugin httpserver 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/httpserver.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/httpserver
	@echo Build plugin log 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/log.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/log
	@echo Build plugin static 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/static.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/static

$(CMD_DIR): FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

test:
	@echo Test sys/ffmpeg
	@${GO} test ./sys/ffmpeg
	@echo Test pkg/media
	@${GO} test ./pkg/media

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
