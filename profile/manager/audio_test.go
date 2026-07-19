package manager_test

import (
	"errors"
	"net/url"
	"testing"

	// Packages
	uuid "github.com/google/uuid"
	test "github.com/mutablelogic/go-media/profile/test"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	require "github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestCreateAudioProfile(t *testing.T) {
	if ff.AVCodec_find_encoder_by_name("aac") == nil {
		t.Skip("aac encoder is not available")
	}

	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	profile, err := mgr.CreateAudioProfile(ctx, "aac", url.Values{})
	require.NoError(err)
	require.NotNil(profile)
	require.NotEqual(uuid.Nil, profile.Id)
}

func TestListAudioCodecs(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	result, err := mgr.ListAudioCodecs(ctx)
	require.NoError(err)
	require.NotNil(result)
	require.Equal(uint64(len(result.Body)), result.Count)

	for _, codec := range result.Body {
		require.NotEmpty(codec.Name)
		t.Log("Audio codec:", codec.Name, "description:", codec.Description)
	}

	if ff.AVCodec_find_encoder_by_name("aac") != nil {
		found := false
		for _, codec := range result.Body {
			if codec.Name == "aac" {
				found = true
				break
			}
		}
		require.True(found)
	}
}

func TestGetAudioProfile(t *testing.T) {
	if ff.AVCodec_find_encoder_by_name("aac") == nil {
		t.Skip("aac encoder is not available")
	}

	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	created, err := mgr.CreateAudioProfile(ctx, "aac", url.Values{})
	require.NoError(err)
	require.NotNil(created)
	require.NotEqual(uuid.Nil, created.Id)

	got, err := mgr.GetAudioProfile(ctx, created.Id)
	require.NoError(err)
	require.NotNil(got)

	require.Equal(created.Id, got.Id)
	require.Equal(created.Bitrate, got.Bitrate)
	require.Equal(created.SampleRate, got.SampleRate)
	require.Equal(created.SampleFormat, got.SampleFormat)
	require.Equal(created.Channels, got.Channels)
	require.Equal(created.Opts, got.Opts)
}

func TestDeleteAudioProfile(t *testing.T) {
	if ff.AVCodec_find_encoder_by_name("aac") == nil {
		t.Skip("aac encoder is not available")
	}

	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	created, err := mgr.CreateAudioProfile(ctx, "aac", url.Values{})
	require.NoError(err)
	require.NotNil(created)
	require.NotEqual(uuid.Nil, created.Id)

	deleted, err := mgr.DeleteAudioProfile(ctx, created.Id)
	require.NoError(err)
	require.NotNil(deleted)
	require.Equal(created.Id, deleted.Id)

	_, err = mgr.GetAudioProfile(ctx, created.Id)
	require.Error(err)
	require.True(errors.Is(err, pg.ErrNotFound))
}

func TestDeleteAudioProfileGone(t *testing.T) {
	if ff.AVCodec_find_encoder_by_name("aac") == nil {
		t.Skip("aac encoder is not available")
	}

	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	created, err := mgr.CreateAudioProfile(ctx, "aac", url.Values{})
	require.NoError(err)

	_, err = mgr.DeleteAudioProfile(ctx, created.Id)
	require.NoError(err)

	_, err = mgr.DeleteAudioProfile(ctx, created.Id)
	require.Error(err)
	require.True(errors.Is(err, pg.ErrNotFound))
}
