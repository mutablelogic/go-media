// Package segmenter provides audio segmentation functionality.
//
// The segmenter reads audio from any supported format (using the reader
// package) and outputs segments of mono float32 samples at a chosen sample
// rate. It supports:
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
