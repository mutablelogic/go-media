ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:1.25 AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .

# Install dependencies
RUN set -x && apt update -y \
  && apt install -y ca-certificates lsb-release build-essential cmake nasm curl pkg-config \
  && apt install -y libfreetype-dev libmp3lame-dev libopus-dev libvorbis-dev libvpx-dev libx264-dev libx265-dev libnuma-dev \
  && apt install -y libvulkan-dev

# Build all the commands
RUN set -x && OS=${OS} ARCH=${ARCH} make all

# Copy commands to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:trixie-slim
ARG OS
ARG ARCH
ARG SOURCE
RUN set -x && apt update -y \
  && apt install -y ca-certificates lsb-release \
  && apt install -y libfreetype6 libmp3lame0 libopus0 libvorbis0a libvorbisenc2 libvpx9 libx264-164 libx265-215 libnuma1 \
  && apt install -y libvulkan1 vulkan-tools

COPY --from=builder /usr/src/app/build/* /usr/local/bin/

# Label the image
LABEL org.opencontainers.image.source=${SOURCE}

# Always run gomedia with any provided arguments
ENTRYPOINT [ "/usr/local/bin/gomedia" ]
EXPOSE 80 443
STOPSIGNAL SIGQUIT
