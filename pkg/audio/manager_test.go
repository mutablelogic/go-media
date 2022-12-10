package audio_test

import (
	"os"
	"testing"
	"time"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/audio"
)

const (
	FILE_S16_MONO_22050   = "../../etc/s16le_22050_1ch_audio.raw"
	FILE_S16_MONO_44100   = "../../etc/s16le_44100_1ch_audio.raw"
	FILE_S16_STEREO_44100 = "../../etc/s16le_44100_2ch_audio.raw"
)

func Test_manager_000(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)

	// Create an audioframe for input
	in, err := NewAudioFrame(AudioFormat{Rate: 22050, Format: SAMPLE_FORMAT_S16, Layout: CHANNEL_LAYOUT_MONO}, time.Second)
	assert.NoError(err)

	// Read data
	r, err := os.Open(FILE_S16_MONO_22050)
	assert.NoError(err)
	defer r.Close()

	assert.NoError(mgr.Convert(in, AudioFormat{Format: SAMPLE_FORMAT_DBL, Rate: 44100}, func(out AudioFrame) error {
		n, err := in.Read(r, 0)
		t.Log("n=", n)
		if err != nil {
			t.Log("err=", err)
		}
		return err
	}))

	// Close
	assert.NoError(in.Close())
	assert.NoError(mgr.Close())
}

/*

 * Once all values have been set, it must be initialized with swr_init(). If
 * you need to change the conversion parameters, you can change the parameters
 * using @ref AVOptions, as described above in the first example; or by using
 * swr_alloc_set_opts2(), but with the first argument the allocated context.
 * You must then call swr_init() again.
 *
 * The conversion itself is done by repeatedly calling swr_convert().
 * Note that the samples may get buffered in swr if you provide insufficient
 * output space or if sample rate conversion is done, which requires "future"
 * samples. Samples that do not require future input can be retrieved at any
 * time by using swr_convert() (in_count can be set to 0).
 * At the end of conversion the resampling buffer can be flushed by calling
 * swr_convert() with NULL in and 0 in_count.
 *
 * The samples used in the conversion process can be managed with the libavutil
 * @ref lavu_sampmanip "samples manipulation" API, including av_samples_alloc()
 * function used in the following example.
 *
 * The delay between input and output, can at any time be found by using
 * swr_get_delay().
 *
 * The following code demonstrates the conversion loop assuming the parameters
 * from above and caller-defined functions get_input() and handle_output():
 * @code
 * uint8_t **input;
 * int in_samples;
 *
 * while (get_input(&input, &in_samples)) {
 *     uint8_t *output;
 *     int out_samples = av_rescale_rnd(swr_get_delay(swr, 48000) +
 *                                      in_samples, 44100, 48000, AV_ROUND_UP);
 *     av_samples_alloc(&output, NULL, 2, out_samples,
 *                      AV_SAMPLE_FMT_S16, 0);
 *     out_samples = swr_convert(swr, &output, out_samples,
 *                                      input, in_samples);
 *     handle_output(output, out_samples);
 *     av_freep(&output);
 * }
 * @endcode
 *
 * When the conversion is finished, the conversion
 * context and everything associated with it must be freed with swr_free().
 * A swr_close() function is also available, but it exists mainly for
 * compatibility with libavresample, and is not required to be called.
 *
 * There will be no memory leak if the data is not completely flushed before
 * swr_free().
 */
