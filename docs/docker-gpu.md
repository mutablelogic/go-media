# Docker GPU Acceleration with Vulkan

The go-media Docker image includes Vulkan support for hardware-accelerated video encoding/decoding. To use GPU acceleration from within the container, you must pass through GPU devices from the host.

## Prerequisites

Install GPU drivers and Vulkan tools on your host system:

```bash
# NVIDIA
apt install nvidia-driver-XXX nvidia-container-toolkit vulkan-tools
systemctl restart docker

# AMD/Intel
apt install mesa-vulkan-drivers vulkan-tools

# Jetson/Tegra (already includes drivers)
apt install vulkan-tools

# Raspberry Pi
apt install mesa-vulkan-drivers vulkan-tools

# Verify
vulkaninfo --summary
```

## Running the Container with GPU Access

### NVIDIA GPUs

```bash
docker run --rm -it --gpus all \
  ghcr.io/mutablelogic/go-media:latest \
  list-codecs
```

### AMD/Intel GPUs

```bash
docker run --rm -it \
  --device=/dev/dri:/dev/dri \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  ghcr.io/mutablelogic/go-media:latest \
  list-codecs
```

### NVIDIA Jetson/Tegra

```bash
docker run --rm -it --runtime nvidia \
  --device=/dev/nvhost-ctrl --device=/dev/nvhost-ctrl-gpu \
  --device=/dev/nvhost-prof-gpu --device=/dev/nvmap \
  --device=/dev/nvhost-gpu --device=/dev/nvhost-as-gpu \
  -v /usr/lib/aarch64-linux-gnu/tegra:/usr/lib/aarch64-linux-gnu/tegra:ro \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  ghcr.io/mutablelogic/go-media:latest \
  list-codecs
```

### Raspberry Pi

```bash
docker run --rm -it \
  --device=/dev/dri:/dev/dri \
  --device=/dev/video10:/dev/video10 \
  --device=/dev/video11:/dev/video11 \
  --device=/dev/video12:/dev/video12 \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  ghcr.io/mutablelogic/go-media:latest \
  list-codecs
```

## Troubleshooting

```bash
# Verify GPU on host
vulkaninfo --summary

# Check GPU in container (NVIDIA)
docker run --rm -it --gpus all \
  ghcr.io/mutablelogic/go-media:latest \
  vulkaninfo --summary

# Check GPU in container (AMD/Intel/Pi)
docker run --rm -it --device=/dev/dri:/dev/dri \
  ghcr.io/mutablelogic/go-media:latest \
  vulkaninfo --summary

# Fix /dev/dri permissions
sudo usermod -aG video,render $USER
```

## References

- [NVIDIA Container Toolkit Documentation](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html)
- [FFmpeg Hardware Acceleration Guide](https://trac.ffmpeg.org/wiki/HWAccelIntro)
- [Vulkan Documentation](https://www.khronos.org/vulkan/)
