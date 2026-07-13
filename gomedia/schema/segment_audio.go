package schema

import (
	"io"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SegmentAudioRequest struct {
	Reader           io.Reader     `json:"-" kong:"-"`
	OutputDir        string        `json:"output_dir,omitempty" name:"out" help:"Output directory for encoded segment M4A files."`
	Duration         time.Duration `json:"duration,omitempty" name:"duration" help:"Target segment duration (e.g. 30s). Use 0s to disable fixed-size splits."`
	Silence          bool          `json:"silence" name:"silence" help:"Enable silence-based segmentation." negatable:"" default:"true"`
	SilenceDuration  time.Duration `json:"silence_duration,omitempty" name:"silence-duration" help:"Minimum silence duration used for silence-based splitting (e.g. 500ms). Also enables silence splitting."`
	SilenceThreshold float64       `json:"silence_threshold,omitempty" name:"silence-threshold" help:"Silence threshold as RMS energy (0.0-1.0). Also enables silence splitting. 0 uses auto threshold (0.005)."`
}
