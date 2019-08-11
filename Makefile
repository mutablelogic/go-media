# Go parameters
GOCMD=go
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
PKG_CONFIG_PATH="pkg-config"
GOFLAGS=-ldflags "-s"

install:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) $(GOINSTALL) $(GOFLAGS) ./ffmpeg

test:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) $(GOTEST) -v ./ffmpeg
