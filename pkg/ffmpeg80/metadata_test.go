package ffmpeg

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST BASIC OPERATIONS

func Test_metadata_new(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("title", "Test Title")
	assert.NotNil(m)
	assert.Equal("title", m.Key())
	assert.Equal("Test Title", m.Value())
}

func Test_metadata_nil(t *testing.T) {
	assert := assert.New(t)

	var m *Metadata
	assert.Empty(m.Key())
	assert.Empty(m.Value())
	assert.Nil(m.Bytes())
	assert.Nil(m.Image())
	assert.Nil(m.Any())
	assert.Empty(m.String())
}

////////////////////////////////////////////////////////////////////////////////
// TEST KEY AND VALUE

func Test_metadata_key(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("artist", "Test Artist")
	assert.Equal("artist", m.Key())
}

func Test_metadata_value_string(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("album", "Test Album")
	assert.Equal("Test Album", m.Value())
	assert.Equal("Test Album", m.Any())
}

func Test_metadata_value_int(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("year", 2024)
	assert.Equal("2024", m.Value())
	assert.Equal(2024, m.Any())
}

func Test_metadata_value_nil(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("empty", nil)
	assert.Empty(m.Value())
	assert.Nil(m.Any())
}

////////////////////////////////////////////////////////////////////////////////
// TEST BYTES

func Test_metadata_bytes_from_bytes(t *testing.T) {
	assert := assert.New(t)

	data := []byte{0x01, 0x02, 0x03, 0x04}
	m := NewMetadata("data", data)

	result := m.Bytes()
	assert.NotNil(result)
	assert.Equal(data, result)
}

func Test_metadata_bytes_from_string(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("text", "hello world")

	result := m.Bytes()
	assert.NotNil(result)
	assert.Equal([]byte("hello world"), result)
}

func Test_metadata_bytes_from_int(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("number", 123)

	result := m.Bytes()
	assert.Nil(result)
}

func Test_metadata_bytes_nil_value(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("empty", nil)
	assert.Nil(m.Bytes())
}

////////////////////////////////////////////////////////////////////////////////
// TEST IMAGE

func Test_metadata_image_valid_png(t *testing.T) {
	assert := assert.New(t)

	// Create a simple 1x1 PNG image
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	assert.NoError(err)

	m := NewMetadata(MetaArtwork, buf.Bytes())

	result := m.Image()
	assert.NotNil(result)
	assert.Equal(1, result.Bounds().Dx())
	assert.Equal(1, result.Bounds().Dy())
}

func Test_metadata_image_invalid_data(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata(MetaArtwork, []byte{0x00, 0x01, 0x02})

	result := m.Image()
	assert.Nil(result)
}

func Test_metadata_image_from_string(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("text", "not an image")

	result := m.Image()
	assert.Nil(result)
}

func Test_metadata_image_nil_value(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("empty", nil)
	assert.Nil(m.Image())
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_metadata_marshal_json(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("title", "Test Title")

	data, err := json.Marshal(m)
	assert.NoError(err)
	assert.NotNil(data)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(err)
	assert.Equal("title", result["key"])
	assert.Equal("Test Title", result["value"])
}

func Test_metadata_marshal_json_nil(t *testing.T) {
	assert := assert.New(t)

	var m *Metadata

	data, err := json.Marshal(m)
	assert.NoError(err)
	assert.Equal("null", string(data))
}

func Test_metadata_marshal_json_nil_value(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("empty", nil)

	data, err := json.Marshal(m)
	assert.NoError(err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(err)
	assert.Equal("empty", result["key"])
	_, hasValue := result["value"]
	assert.False(hasValue) // omitempty should exclude nil values
}

func Test_metadata_string(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("artist", "Test Artist")

	str := m.String()
	assert.NotEmpty(str)
	assert.Contains(str, "artist")
	assert.Contains(str, "Test Artist")
}

func Test_metadata_string_nil(t *testing.T) {
	assert := assert.New(t)

	var m *Metadata
	assert.Empty(m.String())
}

////////////////////////////////////////////////////////////////////////////////
// TEST METADATA CONSTANT

func Test_metadata_artwork_constant(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("artwork", MetaArtwork)
}

////////////////////////////////////////////////////////////////////////////////
// TEST VALUE MIMETYPE DETECTION

func Test_metadata_value_bytes_mimetype(t *testing.T) {
	assert := assert.New(t)

	// Create a simple 1x1 PNG
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	png.Encode(&buf, img)

	m := NewMetadata(MetaArtwork, buf.Bytes())

	// Value() should return mimetype for byte slices
	value := m.Value()
	assert.Contains(value, "image") // Should detect it as an image
}

////////////////////////////////////////////////////////////////////////////////
// TEST ANY METHOD

func Test_metadata_any_string(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("key", "value")
	assert.Equal("value", m.Any())
}

func Test_metadata_any_bytes(t *testing.T) {
	assert := assert.New(t)

	data := []byte{0x01, 0x02}
	m := NewMetadata("data", data)
	assert.Equal(data, m.Any())
}

func Test_metadata_any_int(t *testing.T) {
	assert := assert.New(t)

	m := NewMetadata("number", 42)
	assert.Equal(42, m.Any())
}

func Test_metadata_any_nil(t *testing.T) {
	assert := assert.New(t)

	var m *Metadata
	assert.Nil(m.Any())
}

////////////////////////////////////////////////////////////////////////////////
// TEST INTERFACE COMPLIANCE

func Test_metadata_interface_compliance(t *testing.T) {
	assert := assert.New(t)

	// Verify the type assertion at package level compiles
	m := NewMetadata("test", "value")
	assert.NotNil(m)

	// Verify all interface methods exist
	_ = m.Key()
	_ = m.Value()
	_ = m.Bytes()
	_ = m.Image()
	_ = m.Any()
}
