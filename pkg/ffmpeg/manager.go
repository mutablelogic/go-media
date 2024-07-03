package ffmpeg

import (

	// Packages
	media "github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/version"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	opts []Opt
}

var _ media.Manager = (*Manager)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	manager *Manager
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new media manager which enumerates the available codecs, formats
// and devices
func NewManager(opt ...Opt) (*Manager, error) {
	var options opts

	// Return existing manager if it exists
	if manager == nil {
		manager = new(Manager)
	}

	// Set default options
	options.level = ff.AV_LOG_WARNING

	// Apply options
	for _, opt := range opt {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	// Set logging
	ff.AVUtil_log_set_level(options.level)
	if options.callback != nil {
		ff.AVUtil_log_set_callback(func(level ff.AVLog, message string, userInfo any) {
			options.callback(message)
		})
	}

	// Initialise network
	ff.AVFormat_network_init()

	// Set force flag - this is used to resample or resize decoded
	// frames even if the target format is the same as the source format
	if options.force {
		manager.opts = append(manager.opts, OptForce())
	}

	// Return success
	return manager, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Open a media file or device for reading, from a path or url.
// If a format is specified, then the format will be used to open
// the file. You can add additional options to the open call as
// key=value pairs
/*
func (manager *Manager) Open(url string, format media.Format, opts ...string) (media.Media, error) {
	opt := append([]Opt{OptInputOpt(opts...)}, manager.opts...)
	if format != nil {
		opt = append(opt, OptInputFormat(format.Name()))
	}
	return Open(url, opt...)
}

// Open an io.Reader for reading. If a format is specified, then the
// format will be used to open the file. You can add additional options
// to the open call as key=value pairs
func (manager *Manager) NewReader(r io.Reader, format media.Format, opts ...string) (media.Media, error) {
	opt := append([]Opt{OptInputOpt(opts...)}, manager.opts...)
	if format != nil {
		opt = append(opt, OptInputFormat(format.Name()))
	}
	return NewReader(r, opt...)
}
*/

/*
func (manager *Manager) Transcode(context,output_writer,input_reader or file,input_mapping_function) {
	// 1. Read the input and detect the streams
	// 2. Make a mapping to output streams
	// 3. Create an output writer or file, with the mapped streams
	// 4. Create one goroutine which reads the input and passes frames to a channel
	// 5. Create a second goroutine which reads the channel and writes to the output
	// 6. When EOF on the input or context is cancelled, then stop
}
*/

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VERSION

// Return version information as metadata key/value pairs
func (manager *Manager) Version() []media.Metadata {
	var result []media.Metadata
	for _, v := range version.Version() {
		result = append(result, NewMetadata(v.Key, v.Value))
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOGGING

// Log error messages
func (manager *Manager) Errorf(v string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_ERROR, v, args...)
}

// Log warning messages
func (manager *Manager) Warningf(v string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_WARNING, v)
}

// Log info messages
func (manager *Manager) Infof(v string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_INFO, v, args...)
}
