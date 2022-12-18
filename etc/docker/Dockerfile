# recommended way to build is using:
# docker build \
#   --build-arg PLATFORM=linux --build-arg ARCH=amd64 --build-arg VERSION=focal \
#   -f etc/docker/Dockerfile .
#
# ${VERSION} should be "focal"
ARG PLATFORM
ARG ARCH
ARG VERSION
FROM --platform=${PLATFORM}/${ARCH} ubuntu:${VERSION} AS builder

# update the base packages
ENV DEBIAN_FRONTEND="noninteractive" TZ="Europe/Berlin"
RUN apt-get update -y && apt-get upgrade -y

# install packages
RUN apt-get install -y apt-utils golang make pkg-config ca-certificates lsb-release software-properties-common

# install other build dependencies
# note we need to install ffmpeg 4 from a different repo for bionic
RUN apt-get install -y  \
  libavcodec-dev libavdevice-dev libavfilter-dev \
  libavformat-dev libswresample-dev libavutil-dev libchromaprint-dev

# Run makefile to build command-line tools
WORKDIR /usr/src/app
COPY . .
RUN make

FROM --platform=${PLATFORM}/${ARCH} ubuntu:${VERSION}
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
