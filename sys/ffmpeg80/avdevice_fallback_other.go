//go:build !darwin

package ffmpeg

// AVFoundationDevice describes a single avfoundation input device. Only
// meaningful on darwin; see the darwin build's doc comment for details.
type AVFoundationDevice struct {
	Index     int
	Name      string
	MediaType AVMediaType
	IsDefault bool
}

// AVFoundationListDevices has no non-darwin implementation: avfoundation
// only exists on darwin.
func AVFoundationListDevices(input *AVInputFormat) []AVFoundationDevice {
	return nil
}

// avDeviceListInputSourcesFallback has no non-darwin implementation: every
// input format on other platforms implements the get_device_list callback
// that avdevice_list_input_sources relies on.
func avDeviceListInputSourcesFallback(device *AVInputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	return nil, nil
}

// avDeviceListOutputSinksFallback has no non-darwin implementation: every
// output format on other platforms implements the get_device_list callback
// that avdevice_list_output_sinks relies on.
func avDeviceListOutputSinksFallback(device *AVOutputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	return nil, nil
}
