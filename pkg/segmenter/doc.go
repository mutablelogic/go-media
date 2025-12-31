// Package segmenter provides audio segmentation functionality.
//
// The segmenter reads audio from any supported format (using FFmpeg) and
// outputs fixed-duration chunks of raw PCM samples. It supports:
//
//   - Fixed segment sizes (e.g., 30 seconds)
//   - Silence-based segmentation (break on silence boundaries)
//   - Sample rate conversion
//   - Mono output (single channel)
//
// Use cases include:
//   - Speech-to-text preprocessing (splitting audio into segments)
//   - Audio analysis (processing chunks at a time)
//   - Streaming audio processing
package segmenter
