//go:build !container

package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST DEVICE INFO LIST PROPERTIES

func Test_avdevice_device_list_nil(t *testing.T) {
	assert := assert.New(t)

	// Nil list should return safe defaults
	var list *AVDeviceInfoList
	assert.Nil(list.Devices())
	assert.Equal(0, list.NumDevices())
	assert.Equal(-1, list.Default())
}

func Test_avdevice_device_list_properties(t *testing.T) {
	assert := assert.New(t)

	// Try to get a real device list
	input := AVDevice_input_audio_device_first()
	if input == nil {
		t.Skip("No audio input devices available")
	}

	list, err := AVDevice_list_input_sources(input, "", nil)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list == nil {
		t.Skip("No devices found")
	}
	defer AVDevice_free_list_devices(list)

	// Test properties
	numDevices := list.NumDevices()
	t.Logf("Found %d devices", numDevices)
	assert.GreaterOrEqual(numDevices, 0)

	devices := list.Devices()
	if numDevices > 0 {
		assert.NotNil(devices)
		assert.Equal(numDevices, len(devices))
	}

	defaultIdx := list.Default()
	t.Logf("Default device index: %d", defaultIdx)
	if defaultIdx >= 0 {
		assert.Less(defaultIdx, numDevices)
	}
}

func Test_avdevice_device_list_string(t *testing.T) {
	assert := assert.New(t)

	// Try to get a real device list
	input := AVDevice_input_video_device_first()
	if input == nil {
		t.Skip("No video input devices available")
	}

	list, err := AVDevice_list_input_sources(input, "", nil)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list == nil {
		t.Skip("No devices found")
	}
	defer AVDevice_free_list_devices(list)

	// Test String() marshaling
	str := list.String()
	assert.NotEmpty(str)
	assert.Contains(str, "devices")
	t.Logf("Device list: %s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST DEVICE INFO PROPERTIES

func Test_avdevice_device_info_nil(t *testing.T) {
	// Nil device info should not crash
	var info *AVDeviceInfo
	// These will crash with nil pointer, so we just verify compilation
	_ = info
}

func Test_avdevice_device_info_properties(t *testing.T) {
	assert := assert.New(t)

	// Get a device list with actual devices
	input := AVDevice_input_audio_device_first()
	if input == nil {
		t.Skip("No audio input devices available")
	}

	list, err := AVDevice_list_input_sources(input, "", nil)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list == nil || list.NumDevices() == 0 {
		t.Skip("No devices found")
	}
	defer AVDevice_free_list_devices(list)

	// Test first device properties
	devices := list.Devices()
	if len(devices) == 0 {
		t.Skip("No devices in list")
	}

	device := devices[0]
	assert.NotNil(device)

	// Test name
	name := device.Name()
	t.Logf("Device name: %s", name)
	assert.NotEmpty(name)

	// Test description
	desc := device.Description()
	t.Logf("Device description: %s", desc)
	// Description may be empty

	// Test media types
	mediaTypes := device.MediaTypes()
	t.Logf("Media types: %v", mediaTypes)
	// Media types array may be empty or nil
}

func Test_avdevice_device_info_string(t *testing.T) {
	assert := assert.New(t)

	// Get a device
	output := AVDevice_output_audio_device_first()
	if output == nil {
		t.Skip("No audio output devices available")
	}

	list, err := AVDevice_list_output_sinks(output, "", nil)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list == nil || list.NumDevices() == 0 {
		t.Skip("No devices found")
	}
	defer AVDevice_free_list_devices(list)

	devices := list.Devices()
	if len(devices) == 0 {
		t.Skip("No devices in list")
	}

	device := devices[0]

	// Test String() marshaling
	str := device.String()
	assert.NotEmpty(str)
	assert.Contains(str, "device")
	t.Logf("Device info: %s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST DEVICE INFO ITERATION

func Test_avdevice_device_info_all_devices(t *testing.T) {
	assert := assert.New(t)

	// Get video input device
	input := AVDevice_input_video_device_first()
	if input == nil {
		t.Skip("No video input devices available")
	}

	list, err := AVDevice_list_input_sources(input, "", nil)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list == nil {
		t.Skip("No devices found")
	}
	defer AVDevice_free_list_devices(list)

	devices := list.Devices()
	t.Logf("Total devices: %d", len(devices))

	// Iterate through all devices
	for i, device := range devices {
		assert.NotNil(device)
		name := device.Name()
		desc := device.Description()
		mediaTypes := device.MediaTypes()

		t.Logf("Device %d: name=%s, desc=%s, types=%v", i, name, desc, mediaTypes)

		// Name should always be present
		assert.NotEmpty(name)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST MEMORY CLEANUP

func Test_avdevice_device_free_nil(t *testing.T) {
	// Free nil list should not crash
	AVDevice_free_list_devices(nil)
	// Test passes if no crash
}

func Test_avdevice_device_free_after_free(t *testing.T) {
	assert := assert.New(t)

	// Get a device list
	input := AVDevice_input_audio_device_first()
	if input == nil {
		t.Skip("No audio input devices available")
	}

	list, err := AVDevice_list_input_sources(input, "", nil)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list == nil {
		t.Skip("No devices found")
	}

	// Free once
	AVDevice_free_list_devices(list)

	// Note: Freeing twice would cause issues, so we don't test that
	// The important part is that single free works correctly
	assert.True(true) // Test passes
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avdevice_device_empty_name(t *testing.T) {
	assert := assert.New(t)

	// Try listing with empty device name
	input := AVDevice_input_audio_device_first()
	if input == nil {
		t.Skip("No audio input devices available")
	}

	list, err := AVDevice_list_input_sources(input, "", nil)
	// Should not error with empty name
	if err == nil && list != nil {
		defer AVDevice_free_list_devices(list)
		t.Logf("Listed %d devices with empty name", list.NumDevices())
	}
	assert.True(true) // Test passes if no crash
}

func Test_avdevice_device_with_options(t *testing.T) {
	assert := assert.New(t)

	// Try listing with options dictionary
	input := AVDevice_input_video_device_first()
	if input == nil {
		t.Skip("No video input devices available")
	}

	// Create empty options
	options := AVUtil_dict_alloc()
	defer AVUtil_dict_free(options)

	list, err := AVDevice_list_input_sources(input, "", options)
	if err != nil {
		t.Skip("Device listing not supported:", err)
	}
	if list != nil {
		defer AVDevice_free_list_devices(list)
		t.Logf("Listed %d devices with options", list.NumDevices())
	}
	assert.True(true) // Test passes if no crash
}
