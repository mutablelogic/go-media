package manager_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	test "github.com/mutablelogic/go-media/profile/test"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
	require "github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestListCodecsByType(t *testing.T) {
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	cases := []schema.CodecType{
		schema.CodecType(ff.AVMEDIA_TYPE_AUDIO),
		schema.CodecType(ff.AVMEDIA_TYPE_VIDEO),
		schema.CodecType(ff.AVMEDIA_TYPE_SUBTITLE),
	}

	for _, codecType := range cases {
		t.Run(codecType.String(), func(t *testing.T) {
			require := require.New(t)

			resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{Type: types.Ptr(codecType)})
			require.NoError(err)
			require.NotNil(resp)
			require.Greater(resp.Count, uint64(0))
			require.Len(resp.Body, int(resp.Count))
			for _, codec := range resp.Body {
				require.Equal(codecType, codec.Type)
				require.NotEmpty(codec.Name)
			}
		})
	}
}

func TestListCodecsNoFilterReturnsAllTypes(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{})
	require.NoError(err)
	require.NotNil(resp)
	require.Len(resp.Body, int(resp.Count))

	seen := make(map[schema.CodecType]bool)
	for _, codec := range resp.Body {
		seen[codec.Type] = true
	}
	require.True(seen[schema.CodecType(ff.AVMEDIA_TYPE_AUDIO)], "expected at least one audio codec")
	require.True(seen[schema.CodecType(ff.AVMEDIA_TYPE_VIDEO)], "expected at least one video codec")
	require.True(seen[schema.CodecType(ff.AVMEDIA_TYPE_SUBTITLE)], "expected at least one subtitle codec")
}

func TestListCodecsNoFilterReturnsEncodersAndDecoders(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{})
	require.NoError(err)

	var sawEncoder, sawDecoder bool
	for _, codec := range resp.Body {
		if codec.IsEncoder {
			sawEncoder = true
		}
		if codec.IsDecoder {
			sawDecoder = true
		}
	}
	require.True(sawEncoder, "expected at least one encoder when IsEncoder is unset")
	require.True(sawDecoder, "expected at least one decoder when IsEncoder is unset")
}

func TestListCodecsFilterIsEncoder(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	encoders, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsEncoder: types.Ptr(true)})
	require.NoError(err)
	require.NotEmpty(encoders.Body)
	for _, codec := range encoders.Body {
		require.True(codec.IsEncoder)
	}

	decoders, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsEncoder: types.Ptr(false)})
	require.NoError(err)
	require.NotEmpty(decoders.Body)
	for _, codec := range decoders.Body {
		require.True(codec.IsDecoder)
	}
}

func TestListCodecsFilterIsDecoder(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	decoders, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsDecoder: types.Ptr(true)})
	require.NoError(err)
	require.NotEmpty(decoders.Body)
	for _, codec := range decoders.Body {
		require.True(codec.IsDecoder)
	}

	nonDecoders, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsDecoder: types.Ptr(false)})
	require.NoError(err)
	require.NotEmpty(nonDecoders.Body)
	for _, codec := range nonDecoders.Body {
		require.False(codec.IsDecoder)
	}
}

func TestListCodecsFilterIsEncoderAndIsDecoderCombined(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	// A single ffmpeg codec entry is always exactly one direction, so
	// requiring both simultaneously should never match anything.
	resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsEncoder: types.Ptr(true), IsDecoder: types.Ptr(true)})
	require.NoError(err)
	require.Empty(resp.Body)
}

func TestListCodecsFilterIsHardware(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	full, err := mgr.ListCodecs(ctx, schema.CodecListRequest{})
	require.NoError(err)

	hardware, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsHardware: types.Ptr(true)})
	require.NoError(err)
	for _, codec := range hardware.Body {
		require.True(codec.IsHardware)
	}

	software, err := mgr.ListCodecs(ctx, schema.CodecListRequest{IsHardware: types.Ptr(false)})
	require.NoError(err)
	for _, codec := range software.Body {
		require.False(codec.IsHardware)
	}

	// Whether or not this ffmpeg build registers any hardware codecs, the two
	// filters must partition the full result set exactly.
	require.Equal(full.Count, hardware.Count+software.Count)
}

func TestGetCodecDefaultsToEncoderThenDecoder(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	// "aac" has both an encoder and a decoder registered; the unset default
	// should resolve the encoder for backwards compatibility.
	codec, err := mgr.GetCodec(ctx, "aac", schema.CodecGetRequest{})
	require.NoError(err)
	require.True(codec.IsEncoder)
}

func TestGetCodecIsEncoderFilter(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	decoder, err := mgr.GetCodec(ctx, "aac", schema.CodecGetRequest{IsEncoder: types.Ptr(false)})
	require.NoError(err)
	require.True(decoder.IsDecoder)
	require.False(decoder.IsEncoder)

	encoder, err := mgr.GetCodec(ctx, "aac", schema.CodecGetRequest{IsEncoder: types.Ptr(true)})
	require.NoError(err)
	require.True(encoder.IsEncoder)
}

func TestGetCodecDecoderOptsExcludeEncodeOnlyTemplate(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	// bitrate/sample_rate/sample_format/channel_layout are encode-time knobs
	// that don't apply to decoding — the bitstream dictates them, not the
	// caller — so a decoder's Opts should never include them.
	decoder, err := mgr.GetCodec(ctx, "aac", schema.CodecGetRequest{IsEncoder: types.Ptr(false)})
	require.NoError(err)
	require.True(decoder.IsDecoder)

	encodeOnly := map[string]bool{
		schema.OptionAudioBitrate:  true,
		schema.OptionSampleRate:    true,
		schema.OptionSampleFormat:  true,
		schema.OptionChannelLayout: true,
	}
	for _, opt := range decoder.Opts {
		require.False(encodeOnly[opt.Name], "decoder %q should not expose encode-only option %q", decoder.Name, opt.Name)
	}
}

func TestListCodecsPaging(t *testing.T) {
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	// Filter to audio codecs to keep the result set stable and non-trivial in size.
	audioType := types.Ptr(schema.CodecType(ff.AVMEDIA_TYPE_AUDIO))

	full, err := mgr.ListCodecs(ctx, schema.CodecListRequest{Type: audioType})
	require.NoError(t, err)
	require.NotNil(t, full)
	require.GreaterOrEqual(t, full.Count, uint64(5), "expected at least 5 audio codecs to test paging")

	t.Run("offset and limit mid-list", func(t *testing.T) {
		require := require.New(t)

		limit := uint64(2)
		resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{
			Type:        audioType,
			OffsetLimit: pg.OffsetLimit{Offset: 2, Limit: &limit},
		})
		require.NoError(err)
		require.NotNil(resp)
		require.Equal(full.Count, resp.Count)
		require.Equal(uint64(2), resp.Offset)
		require.Equal(uint64(2), types.Value(resp.Limit))
		require.Equal(full.Body[2:4], resp.Body)
	})

	t.Run("limit clamped near end", func(t *testing.T) {
		require := require.New(t)

		offset := full.Count - 2
		limit := uint64(10)
		resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{
			Type:        audioType,
			OffsetLimit: pg.OffsetLimit{Offset: offset, Limit: &limit},
		})
		require.NoError(err)
		require.Equal(full.Count, resp.Count)
		require.Len(resp.Body, 2)
		require.Equal(uint64(2), types.Value(resp.Limit), "limit should be clamped to the items available after offset")
		require.Equal(full.Body[offset:], resp.Body)
	})

	t.Run("offset beyond total returns no items", func(t *testing.T) {
		require := require.New(t)

		resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{
			Type:        audioType,
			OffsetLimit: pg.OffsetLimit{Offset: full.Count + 10},
		})
		require.NoError(err)
		require.Equal(full.Count, resp.Count)
		require.Empty(resp.Body)
	})

	t.Run("limit of zero returns count only", func(t *testing.T) {
		require := require.New(t)

		limit := uint64(0)
		resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{
			Type:        audioType,
			OffsetLimit: pg.OffsetLimit{Limit: &limit},
		})
		require.NoError(err)
		require.Equal(full.Count, resp.Count)
		require.Empty(resp.Body)
	})

	t.Run("no offset or limit returns everything", func(t *testing.T) {
		require := require.New(t)

		resp, err := mgr.ListCodecs(ctx, schema.CodecListRequest{Type: audioType})
		require.NoError(err)
		require.Equal(full.Count, resp.Count)
		require.Equal(full.Body, resp.Body)
	})
}
