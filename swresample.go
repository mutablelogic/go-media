package media

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SWResample is an interface to the ffmpeg swresample library
// which resamples audio.
type SWResample interface {
	// Create a new empty context object for conversion. Returns a
	// cancel function which can interrupt the conversion.
	NewContext() SWResampleContext

	// Convert the input data to the output data, until io.EOF is
	// returned or an error occurs, for uint8 data.
	ConvertBytes(SWResampleContext, SWResampleConvertBytes) error
}

// SWResampleConvert is a function that accepts an "output" buffer of data,
// which can be nil if the conversion has not started yet, and should return
// the next buffer of input data. Return any error
// for the conversion to stop (io.EOF should be returned at the end of
// any data conversion)
type SWResampleConvertBytes func(SWResampleContext, []byte) ([]byte, error)

type SWResampleContext interface {
	// Set the input audio format
	SetIn(AudioFormat) error

	// Set the output audio format
	SetOut(AudioFormat) error
}
