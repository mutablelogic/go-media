// Package thumbnailer finds the scenes of a video, and representative
// thumbnails of it.
//
// Frames are sampled from the keyframes, when the video has them at a
// useful rate, or by seeking to evenly spaced times otherwise. Each sampled
// frame is measured for how good a thumbnail it makes (brightness, contrast,
// sharpness, colourfulness and so on), and compared with the one before it
// to find where scenes change. A thumbnail is then the best frame of a
// scene, and thumbnails are chosen from different scenes which don't look
// alike.
package thumbnailer
