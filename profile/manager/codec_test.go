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
