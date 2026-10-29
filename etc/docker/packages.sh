#!/bin/sh
# packages.sh prints the Debian packages go-media's libraries need, to build
# them (dev) or to run what's linked with them (runtime), for an FFmpeg
# license (gpl or lgpl, as the Makefile's FFMPEG_LICENSE).
#
#   apt install -y $(etc/docker/packages.sh dev gpl)
#
# The libraries are static, so only the libraries they link with are needed
# to run. An lgpl build only has decoders and libraries which FFmpeg allows
# in an LGPL build: libheif's HEVC and AV1 decoders, which FFmpeg also uses,
# the compression and XML libraries, and OpenSSL, for HTTPS. A gpl build
# also has the encoders and other libraries FFmpeg is built with, including
# x264 and x265, which make it GPL.
set -e

usage() {
	echo "usage: $0 dev|runtime gpl|lgpl" >&2
	exit 2
}

[ $# -eq 2 ] || usage
case "$1" in
dev)
	# The tools to build the libraries, and the headers of what they link
	# with. OpenSSL links statically with zstd, so FFmpeg only finds it with
	# libzstd-dev
	common="ca-certificates build-essential cmake nasm curl pkg-config \
		libde265-dev libaom-dev libdav1d-dev libxml2-dev liblzma-dev zlib1g-dev libssl-dev libzstd-dev"
	gpl="libfreetype-dev libmp3lame-dev libopus-dev libvorbis-dev libvpx-dev \
		libx264-dev libx265-dev libnuma-dev libzvbi-dev libvulkan-dev"
	;;
runtime)
	# The shared libraries which the static libraries link with
	common="ca-certificates libde265-0 libaom3 libdav1d7 libxml2 liblzma5 zlib1g \
		libssl3t64 libzstd1 libstdc++6 libgomp1"
	gpl="libfreetype6 libmp3lame0 libopus0 libvorbis0a libvorbisenc2 libvpx9 \
		libx264-164 libx265-215 libnuma1 libzvbi0 libvulkan1"
	;;
*)
	usage
	;;
esac

case "$2" in
gpl) echo $common $gpl ;;
lgpl) echo $common ;;
*) usage ;;
esac
