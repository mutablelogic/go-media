# recommended way to build is using:
# docker build \
#   --build-arg PLATFORM=linux --build-arg ARCH=amd64 --build-arg VERSION=bullseye \
#   -f etc/docker/Dockerfile .
#
ARG PLATFORM
ARG ARCH
ARG VERSION
FROM --platform=${PLATFORM}/${ARCH} golang:1.19-${VERSION} AS builder

# Run makefile to build command-line tools
WORKDIR /usr/src/app
COPY . .
RUN make

FROM --platform=${PLATFORM}/${ARCH} ubuntu:${VERSION}
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
