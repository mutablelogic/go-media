package media

import (
	"fmt"
	"net/url"
	"strconv"

	// Packages
	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaOutput struct {
	ctx  *ffmpeg.AVFormatContext
	avio *ffmpeg.AVIOContext
	s    map[int]*Stream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMediaOutput(ctx *ffmpeg.AVFormatContext) *MediaOutput {
	// Create object
	this := new(MediaOutput)
	this.ctx = ctx
	this.s = make(map[int]*Stream)

	// success
	return this
}

func (m *MediaOutput) Release() error {
	// Write trailer
	var result error
	if m.ctx != nil {
		if err := m.ctx.WriteTrailer(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Close files
	if m.avio != nil {
		m.avio.Flush()
		if err := m.avio.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Release streams
	for _, stream := range m.s {
		stream.Release()
	}

	// Close media
	if m.ctx != nil {
		m.ctx.Free()
	}

	// Release resources
	m.ctx = nil
	m.avio = nil
	m.s = nil

	// Return success
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *MediaOutput) String() string {
	str := "<media output"
	if url := m.URL(); url != nil {
		str += " url=" + strconv.Quote(url.String())
	}
	if streams := m.Streams(); len(streams) > 0 {
		str += " streams=" + fmt.Sprint(streams)
	}
	if flags := m.Flags(); flags != MEDIA_FLAG_NONE {
		str += " flags=" + fmt.Sprint(flags)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (m *MediaOutput) URL() *url.URL {
	if m.ctx == nil {
		return nil
	}
	return m.ctx.Url()
}

func (m *MediaOutput) Metadata() *Metadata {
	if m.ctx == nil {
		return nil
	}
	return NewMetadata(m.ctx.Metadata())
}

func (m *MediaOutput) IsFile() bool {
	if m.ctx == nil {
		return false
	}
	return m.ctx.Flags()&ffmpeg.AVFMT_NOFILE == 0
}

func (m *MediaOutput) Flags() MediaFlag {
	if m.ctx == nil {
		return MEDIA_FLAG_NONE
	}
	flags := MEDIA_FLAG_ENCODER
	if m.IsFile() {
		flags |= MEDIA_FLAG_FILE
	}
	return flags
}

func (m *MediaOutput) Streams() []*Stream {
	if m.ctx == nil {
		return nil
	}
	result := make([]*Stream, len(m.s))
	for i, s := range m.s {
		result[i] = s
	}
	return result
}

func (m *MediaOutput) StreamForIndex(i int) *Stream {
	if m.ctx == nil {
		return nil
	}
	if i < 0 || i >= len(m.s) {
		return nil
	}
	return m.s[i]
}
