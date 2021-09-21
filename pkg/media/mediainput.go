package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"

	// Packages
	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaInput struct {
	ctx *ffmpeg.AVFormatContext
	s   map[int]*Stream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMediaInput(ctx *ffmpeg.AVFormatContext) *MediaInput {
	// Create object
	this := new(MediaInput)
	this.ctx = ctx

	// Create streams
	streams := this.ctx.Streams()
	if streams == nil {
		return nil
	}
	this.s = make(map[int]*Stream, len(streams))
	for _, stream := range streams {
		key := stream.Index()
		this.s[key] = NewStream(stream, nil)
	}

	// success
	return this
}

func (m *MediaInput) Release() error {
	var result error

	// Release streams
	for _, stream := range m.s {
		if err := stream.Release(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// If there is custom io, free it
	if m.ctx != nil {
		if io := m.ctx.IOContext(); io != nil {
			io.Free()
			m.ctx.SetIOContext(nil)
		}
	}

	// Close input
	if m.ctx != nil {
		m.ctx.CloseInput()
	}

	// Release resources
	m.s = nil
	m.ctx = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *MediaInput) String() string {
	str := "<media input"
	if io := m.CustomIO(); io {
		str += " customio"
	}
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

func (m *MediaInput) URL() *url.URL {
	if m.ctx == nil {
		return nil
	}
	return m.ctx.Url()
}

func (m *MediaInput) CustomIO() bool {
	if m.ctx == nil {
		return false
	}
	if io := m.ctx.IOContext(); io != nil {
		return true
	} else {
		return false
	}
}

func (m *MediaInput) Streams() []*Stream {
	if m.ctx == nil {
		return nil
	}
	result := make([]*Stream, len(m.s))
	for i, s := range m.s {
		result[i] = s
	}
	return result
}

func (m *MediaInput) Metadata() *Metadata {
	if m.ctx == nil {
		return nil
	}
	return NewMetadata(m.ctx.Metadata())
}

func (m *MediaInput) Flags() MediaFlag {
	if m.ctx == nil {
		return MEDIA_FLAG_NONE
	}
	flags := MEDIA_FLAG_DECODER
	if m.ctx.Flags()&ffmpeg.AVFMT_NOFILE == 0 {
		flags |= MEDIA_FLAG_FILE
	}
	for _, stream := range m.Streams() {
		flags |= stream.Flags()
	}

	// Add other flags with likely media file type
	metadata := m.Metadata()
	if flags&MEDIA_FLAG_AUDIO != 0 && metadata.Value(MEDIA_KEY_ALBUM) != nil {
		flags |= MEDIA_FLAG_ALBUM
	}
	if flags&MEDIA_FLAG_ALBUM != 0 && metadata.Value(MEDIA_KEY_ALBUM_ARTIST) != nil && metadata.Value(MEDIA_KEY_TITLE) != nil {
		flags |= MEDIA_FLAG_ALBUM_TRACK
	}
	if flags&MEDIA_FLAG_ALBUM != 0 {
		if compilation, ok := metadata.Value(MEDIA_KEY_COMPILATION).(bool); ok && compilation {
			flags |= MEDIA_FLAG_ALBUM_COMPILATION
		}
	}
	return flags
}

func (m *MediaInput) StreamForIndex(i int) *Stream {
	if m.ctx == nil {
		return nil
	}
	if i < 0 || i >= len(m.s) {
		return nil
	}
	return m.s[i]
}

func (m *MediaInput) StreamsForFlag(flag MediaFlag) []int {
	if m.ctx == nil {
		return nil
	}
	result := []int{}
	for key, stream := range m.s {
		if stream.Flags()&flag != 0 {
			result = append(result, key)
		}
	}
	return result
}

// Iterate over packets in the input stream
func (m *MediaInput) Read(ctx context.Context, streams []int, fn DecodeIteratorFunc) error {
	if fn == nil || m.ctx == nil {
		return ErrBadParameter.With("Read")
	}
	if len(streams) == 0 {
		for index := range m.s {
			streams = append(streams, index)
		}
	}

	// Create decode contexts
	var result error
	streammap := NewStreamMap()
	for _, i := range streams {
		if stream, exists := m.s[i]; !exists {
			result = multierror.Append(result, ErrNotFound.Withf("Stream with index %v", i))
		} else if err := streammap.Set(stream); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Bail out if any errors
	if result != nil {
		return result
	}

	// Create a packet
	packet := ffmpeg.NewAVPacket()
	if packet == nil {
		return ErrInternalAppError.With("NewAVPacket")
	}
	defer packet.Free()

	// Iterate over incoming packets, callback when packet should
	// be processed. Return if parent context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.ctx.ReadPacket(packet); err == io.EOF {
				// End of stream
				break FOR_LOOP
			} else if err != nil {
				return err
			} else if stream := streammap.Get(packet.Stream()); stream != nil {
				// Call decode function with packet
				err := fn(ctx, packet)
				packet.Release()
				if errors.Is(err, io.EOF) {
					// End of stream requested with no error
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
		}
	}

	// Return success
	return nil
}

/*

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ITERATE OVER PACKETS

// Iterate over packets in the input stream
func (this *inputctx) Read(ctx context.Context, streams []int, fn gopi.DecodeIteratorFunc) error {
	// Lock for writing as ReadPacket modifies state
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Check parameters
	if fn == nil || this.ctx == nil {
		return gopi.ErrBadParameter.WithPrefix("DecodeIterator")
	}

	// If streams argument is empty or nil, select all streams
	if len(streams) == 0 {
		for index := range this.streams {
			streams = append(streams, index)
		}
	}

	// Create decode context map and call close on each on exit
	contextmap := make(map[int]*decodectx, len(this.streams))
	defer func() {
		for _, ctx := range contextmap {
			ctx.Close()
		}
	}()

	// Create a stream map (which is used to map streams
	// from input->output)
	streammap := NewStreamMap()

	// Create decode contexts
	for _, index := range streams {
		if stream, exists := this.streams[index]; exists == false {
			return gopi.ErrInternalAppError.WithPrefix("DecodeIterator")
		} else if decodectx := NewDecodeContext(stream, streammap); decodectx == nil {
			return gopi.ErrInternalAppError.WithPrefix("DecodeIterator")
		} else if err := streammap.Set(stream, nil); err != nil {
			return err
		} else {
			contextmap[index] = decodectx
		}
	}

	// Create a packet
	packet := ffmpeg.NewAVPacket()
	if packet == nil {
		return gopi.ErrInternalAppError.WithPrefix("DecodeIterator")
	}
	defer packet.Free()

	// Iterate over incoming packets, callback when packet should
	// be processed. Return if parent context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := this.ctx.ReadPacket(packet); err == io.EOF {
				// End of stream
				break FOR_LOOP
			} else if err != nil {
				return err
			} else if ctx, exists := contextmap[packet.Stream()]; exists {
				// Call decode function with packet
				err := fn(ctx, packet)
				packet.Release()
				if errors.Is(err, io.EOF) {
					// End of stream requested with no error
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
		}
	}

	// Return success
	return nil
}

func (this *inputctx) DecodeFrameIterator(ctx gopi.MediaDecodeContext, packet gopi.MediaPacket, fn gopi.DecodeFrameIteratorFunc) error {
	// Check parameters
	if ctx == nil || fn == nil {
		return gopi.ErrBadParameter.WithPrefix("DecodeFrameIterator")
	}

	// Get internal context object and check more parameters
	ctx_, ok := ctx.(*decodectx)
	if ok == false || packet == nil {
		return gopi.ErrBadParameter.WithPrefix("DecodeFrameIterator")
	}

	// Lock context for writing
	//ctx_.RWMutex.Lock()
	//defer ctx_.RWMutex.Unlock()

	// Decode packet
	if err := ctx_.DecodePacket(packet); err != nil {
		return fmt.Errorf("DecodeFrameIterator: %w", err)
	}

	// Iterate through frames
	for {
		// Return frames until no more available
		if frame, err := ctx_.DecodeFrame(); errors.Is(err, io.EOF) {
			return err
		} else if err != nil {
			return fmt.Errorf("DecodeFrameIterator: %w", err)
		} else if frame == nil {
			// Not enough data, so return without processing frame
			return nil
		} else {
			err := fn(frame)
			frame.Release()
			if err != nil {
				return err
			}
		}
	}
}

*/
