ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:1.24 AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .

# Install dependencies
RUN set -x \
  && apt update -y \
  && apt install -y ca-certificates lsb-release curl

# Build all the commands
RUN set -x \
  && OS=${OS} ARCH=${ARCH} make ffinstall \
  && OS=${OS} ARCH=${ARCH} make cli

# Copy commands to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:bookworm-slim
ARG OS
ARG ARCH
ARG SOURCE
RUN set -x \
  && apt update -y \
  && apt install -y ca-certificates lsb-release
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY --chmod=755 etc/docker/entrypoint.sh .

# Label the image
LABEL org.opencontainers.image.source=${SOURCE}

# Entrypoint when running the server
ENTRYPOINT [ "/entrypoint.sh" ]
CMD [ "/usr/local/bin/media" ]
EXPOSE 80 443
STOPSIGNAL SIGQUIT
