package media

import (
	"fmt"
	"strings"

	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type format_in struct {
	ctx   *ffmpeg.AVInputFormat
	flags MediaFlag
}

type format_out struct {
	ctx      *ffmpeg.AVOutputFormat
	flags    MediaFlag
	audio    *codec
	video    *codec
	subtitle *codec
}

// Ensure *format_in *format_out comply with Media interface
var _ MediaFormat = (*format_in)(nil)
var _ MediaFormat = (*format_out)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a input format container
func NewInputFormat(ctx *ffmpeg.AVInputFormat, flags MediaFlag) *format_in {
	this := new(format_in)

	if ctx == nil {
		return nil
	} else {
		this.ctx = ctx
		this.flags = flags
	}

	// Return success
	return this
}

// Create a output format container
func NewOutputFormat(ctx *ffmpeg.AVOutputFormat, flags MediaFlag) *format_out {
	this := new(format_out)

	if ctx == nil {
		return nil
	} else {
		this.ctx = ctx
		this.flags = flags
	}

	// Default codecs
	if id := this.ctx.DefaultAudioCodec(); id != ffmpeg.AV_CODEC_ID_NONE {
		this.audio = NewCodecEncoder(id)
	}
	if id := this.ctx.DefaultVideoCodec(); id != ffmpeg.AV_CODEC_ID_NONE {
		this.video = NewCodecEncoder(id)
	}
	if id := this.ctx.DefaultSubtitleCodec(); id != ffmpeg.AV_CODEC_ID_NONE {
		this.subtitle = NewCodecEncoder(id)
	}

	// Return success
	return this
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (format *format_in) String() string {
	str := "<media.format"
	if flags := format.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if name := format.Name(); len(name) > 0 {
		if len(name) == 1 {
			str += fmt.Sprintf(" name=%q", name[0])
		} else {
			str += fmt.Sprintf(" name=%q", name)
		}
	}
	if description := format.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if mimetype := format.MimeType(); len(mimetype) > 0 {
		str += fmt.Sprintf(" mimetype=%q", mimetype)
	}
	if ext := format.Ext(); len(ext) > 0 {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	return str + ">"
}

func (format *format_out) String() string {
	str := "<media.format"
	if flags := format.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if name := format.Name(); len(name) > 0 {
		if len(name) == 1 {
			str += fmt.Sprintf(" name=%q", name[0])
		} else {
			str += fmt.Sprintf(" name=%q", name)
		}
	}
	if description := format.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if mimetype := format.MimeType(); len(mimetype) > 0 {
		str += fmt.Sprintf(" mimetype=%q", mimetype)
	}
	if ext := format.Ext(); len(ext) > 0 {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - IN

// Return MEDIA_FLAG_ENCODER and MEDIA_FLAG_DEVICE flags
func (format *format_in) Flags() MediaFlag {
	return format.flags
}

// Return the name of the media format
func (format *format_in) Name() []string {
	return toExt("", format.ctx.Name())
}

// Return a longer description of the media format
func (format *format_in) Description() string {
	return format.ctx.Description()
}

// Return mimetype
func (format *format_in) MimeType() []string {
	return toExt("", format.ctx.MimeType())
}

// Return file extensions
func (format *format_in) Ext() []string {
	return toExt(".", format.ctx.Ext())
}

// Return nil
func (format *format_in) DefaultAudioCodec() Codec {
	return nil
}

func (format *format_in) DefaultVideoCodec() Codec {
	return nil
}

func (format *format_in) DefaultSubtitleCodec() Codec {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OUT

// Return MEDIA_FLAG_DECODER, MEDIA_FLAG_DEVICE, MEDIA_FLAG_AUDIO,
// MEDIA_FLAG_VIDEO and MEDIA_FLAG_SUBTITLE flags
func (format *format_out) Flags() MediaFlag {
	flags := format.flags
	if format.audio != nil {
		flags |= MEDIA_FLAG_AUDIO
	}
	if format.video != nil {
		flags |= MEDIA_FLAG_VIDEO
	}
	if format.subtitle != nil {
		flags |= MEDIA_FLAG_SUBTITLE
	}
	return flags
}

// Return the name of the media format
func (format *format_out) Name() []string {
	return toExt("", format.ctx.Name())
}

// Return a longer description of the media format
func (format *format_out) Description() string {
	return format.ctx.Description()
}

// Return mimetype
func (format *format_out) MimeType() []string {
	return toExt("", format.ctx.MimeType())
}

// Return file extensions
func (format *format_out) Ext() []string {
	return toExt(".", format.ctx.Ext())
}

func (format *format_out) DefaultAudioCodec() Codec {
	return format.audio
}

func (format *format_out) DefaultVideoCodec() Codec {
	return format.video
}

func (format *format_out) DefaultSubtitleCodec() Codec {
	return format.subtitle
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toExt(prefix, exts string) []string {
	result := make([]string, 0, 3)
	if ext := strings.TrimSpace(exts); ext != "" {
		for _, ext := range strings.Split(ext, ",") {
			result = append(result, prefix+strings.ToLower(ext))
		}
	}
	return result
}
