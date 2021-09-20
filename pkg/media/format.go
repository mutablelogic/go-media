package media

import (

	// Packages
	"fmt"

	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"

	// Namespace imports
	. "github.com/djthorpe/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Format struct {
	in  *ffmpeg.AVInputFormat
	out *ffmpeg.AVOutputFormat
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewInputFormat(ctx *ffmpeg.AVInputFormat) *Format {
	if ctx == nil {
		return nil
	}
	return &Format{ctx, nil}
}

func NewOutputFormat(ctx *ffmpeg.AVOutputFormat) *Format {
	if ctx == nil {
		return nil
	}
	return &Format{nil, ctx}
}

func (f *Format) Release() error {
	f.in = nil
	f.out = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f *Format) String() string {
	str := "<format"
	if name := f.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if desc := f.Description(); desc != "" {
		str += fmt.Sprintf(" description=%q", desc)
	}
	if ext := f.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if mimetype := f.MimeType(); mimetype != "" {
		str += fmt.Sprintf(" mimetype=%q", mimetype)
	}
	if flags := f.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (f *Format) Name() string {
	if f.in != nil {
		return f.in.Name()
	} else if f.out != nil {
		return f.out.Name()
	} else {
		return ""
	}
}

func (f *Format) Description() string {
	if f.in != nil {
		return f.in.Description()
	} else if f.out != nil {
		return f.out.Description()
	} else {
		return ""
	}
}

func (f *Format) Ext() string {
	if f.in != nil {
		return f.in.Ext()
	} else if f.out != nil {
		return f.out.Ext()
	} else {
		return ""
	}
}

func (f *Format) MimeType() string {
	if f.in != nil {
		return f.in.MimeType()
	} else if f.out != nil {
		return f.out.MimeType()
	} else {
		return ""
	}
}

func (f *Format) Flags() MediaFlag {
	if f.in != nil {
		return MEDIA_FLAG_DECODER
	} else if f.out != nil {
		return MEDIA_FLAG_ENCODER
	}
	return MEDIA_FLAG_NONE
}
