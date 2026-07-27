//go:build darwin

package ffmpeg

/*
#cgo pkg-config: libavdevice libavformat libavutil
#cgo LDFLAGS: -framework CoreAudio -framework CoreFoundation
#include <libavdevice/avdevice.h>
#include <libavutil/avutil.h>
#include <libavutil/error.h>
#include <libavutil/mem.h>
#include <CoreAudio/CoreAudio.h>
#include <errno.h>
#include <string.h>
#include <stdlib.h>

// ff_device_list_new allocates an empty AVDeviceInfoList, to be populated with
// ff_device_list_set and freed with avdevice_free_list_devices - just like a
// list returned by avdevice_list_output_sinks.
static AVDeviceInfoList *ff_device_list_new(void) {
	AVDeviceInfoList *list = av_mallocz(sizeof(*list));
	if (list) {
		list->default_device = -1;
	}
	return list;
}

// ff_device_list_free frees a list allocated with ff_device_list_new.
static void ff_device_list_free(AVDeviceInfoList *list) {
	avdevice_free_list_devices(&list);
}

// ff_device_list_set places a device at a specific position in list, growing
// (and null-padding intervening slots in) the devices array as needed.
// avdevice_free_list_devices and AVDeviceInfoList.Devices() already treat a
// NULL entry as "no device here", so the gaps are handled transparently -
// this lets the reported index match what FFmpeg's -audio_device_index
// option expects, even though it enumerates devices CoreAudio reports as
// having no output stream too (we just never fill in those slots).
static int ff_device_list_set(AVDeviceInfoList *list, int index, const char *name, const char *description, int media_type, int is_default) {
	if (!list || !name || !name[0] || index < 0) {
		return AVERROR(EINVAL);
	}

	if (index >= list->nb_devices) {
		AVDeviceInfo **devices = av_realloc_array(list->devices, index + 1, sizeof(*devices));
		if (!devices) {
			return AVERROR(ENOMEM);
		}
		for (int i = list->nb_devices; i <= index; i++) {
			devices[i] = NULL;
		}
		list->devices = devices;
		list->nb_devices = index + 1;
	}

	AVDeviceInfo *dev = av_mallocz(sizeof(*dev));
	if (!dev) {
		return AVERROR(ENOMEM);
	}
	dev->device_name = av_strdup(name);
	dev->device_description = av_strdup(description && description[0] ? description : name);
	dev->media_types = av_malloc(sizeof(enum AVMediaType));
	if (!dev->device_name || !dev->device_description || !dev->media_types) {
		av_freep(&dev->device_name);
		av_freep(&dev->device_description);
		av_freep(&dev->media_types);
		av_free(dev);
		return AVERROR(ENOMEM);
	}
	dev->media_types[0] = (enum AVMediaType)media_type;
	dev->nb_media_types = 1;

	list->devices[index] = dev;
	if (is_default) {
		list->default_device = index;
	}
	return 0;
}

// ff_coreaudio_list_output_devices places every CoreAudio device with at
// least one output stream into list at its real CoreAudio device index
// (matching what FFmpeg's audiotoolbox muxer expects for its
// "audio_device_index" option), tagging the system default output device.
// This exists because the audiotoolbox muxer doesn't implement the
// get_device_list callback that avdevice_list_output_sinks relies on - it
// only logs its device list as a side effect of actually starting playback.
static int ff_coreaudio_list_output_devices(AVDeviceInfoList *list) {
	AudioObjectPropertyAddress devicesProp = {
		kAudioHardwarePropertyDevices,
		kAudioObjectPropertyScopeGlobal,
		kAudioObjectPropertyElementMain,
	};

	UInt32 dataSize = 0;
	if (AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &devicesProp, 0, NULL, &dataSize) != noErr || dataSize == 0) {
		return 0;
	}
	int numDevices = dataSize / sizeof(AudioDeviceID);
	AudioDeviceID *devices = (AudioDeviceID *)av_malloc(dataSize);
	if (!devices) {
		return AVERROR(ENOMEM);
	}
	if (AudioObjectGetPropertyData(kAudioObjectSystemObject, &devicesProp, 0, NULL, &dataSize, devices) != noErr) {
		av_free(devices);
		return 0;
	}

	AudioDeviceID defaultOutput = 0;
	UInt32 defSize = sizeof(defaultOutput);
	AudioObjectPropertyAddress defProp = {
		kAudioHardwarePropertyDefaultOutputDevice,
		kAudioObjectPropertyScopeGlobal,
		kAudioObjectPropertyElementMain,
	};
	AudioObjectGetPropertyData(kAudioObjectSystemObject, &defProp, 0, NULL, &defSize, &defaultOutput);

	for (int i = 0; i < numDevices; i++) {
		AudioDeviceID devID = devices[i];

		AudioObjectPropertyAddress streamsProp = {
			kAudioDevicePropertyStreams,
			kAudioObjectPropertyScopeOutput,
			kAudioObjectPropertyElementMain,
		};
		UInt32 streamsSize = 0;
		if (AudioObjectGetPropertyDataSize(devID, &streamsProp, 0, NULL, &streamsSize) != noErr || streamsSize == 0) {
			continue;
		}

		CFStringRef nameRef = NULL;
		UInt32 nameSize = sizeof(nameRef);
		AudioObjectPropertyAddress nameProp = {
			kAudioObjectPropertyName,
			kAudioObjectPropertyScopeGlobal,
			kAudioObjectPropertyElementMain,
		};
		AudioObjectGetPropertyData(devID, &nameProp, 0, NULL, &nameSize, &nameRef);

		char namebuf[256] = {0};
		if (nameRef) {
			CFStringGetCString(nameRef, namebuf, sizeof(namebuf), kCFStringEncodingUTF8);
			CFRelease(nameRef);
		}
		if (namebuf[0] == 0) {
			continue;
		}

		int ret = ff_device_list_set(list, i, namebuf, namebuf, AVMEDIA_TYPE_AUDIO, devID == defaultOutput);
		if (ret < 0) {
			av_free(devices);
			return ret;
		}
	}

	av_free(devices);
	return 0;
}
*/
import "C"

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// AVFoundationDevice describes a single avfoundation input device. Unlike
// AVDeviceInfo, it carries an Index that is local to its MediaType - video
// and audio devices are independently numbered from 0, matching what
// FFmpeg's avfoundation "<video_index>:<audio_index>" input URL syntax
// expects.
type AVFoundationDevice struct {
	Index     int
	Name      string
	MediaType AVMediaType
	IsDefault bool
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Example line: [0] FaceTime HD Camera
	avfoundationDevicePattern = regexp.MustCompile(`\[(\d+)\]\s+(.+)$`)
)

////////////////////////////////////////////////////////////////////////////////
// AVFOUNDATION

// AVFoundationListDevices lists avfoundation input devices by opening the
// demuxer with list_devices=true and scraping its log output, since that's
// the only enumeration avfoundation supports.
//
// This doesn't go through AVDevice_list_input_sources / AVDeviceInfoList
// because avfoundation multiplexes two independently-indexed device
// categories (video, audio) through one demuxer, and the generic
// AVDeviceInfo has no index field and AVDeviceInfoList only has room for one
// overall default device - neither can represent that without collisions.
func AVFoundationListDevices(input *AVInputFormat) []AVFoundationDevice {
	if input == nil || input.Name() != "avfoundation" {
		return nil
	}

	var mu sync.Mutex
	var capturedLines []string

	oldLevel := AVUtil_log_get_level()
	AVUtil_log_set_level(AV_LOG_VERBOSE)
	defer AVUtil_log_set_level(oldLevel)

	AVUtil_log_set_callback(func(level AVLog, message string, userInfo any) {
		if level < AV_LOG_INFO || level > AV_LOG_VERBOSE {
			return
		}
		line := strings.TrimSpace(message)
		lower := strings.ToLower(line)
		if !(strings.Contains(lower, "video devices:") ||
			strings.Contains(lower, "audio devices:") ||
			avfoundationDevicePattern.MatchString(line)) {
			return
		}
		mu.Lock()
		capturedLines = append(capturedLines, line)
		mu.Unlock()
	})
	defer AVUtil_log_set_callback(nil)

	options := AVUtil_dict_alloc()
	if options == nil {
		return nil
	}
	defer AVUtil_dict_free(options)
	AVUtil_dict_set(options, "list_devices", "true", AV_DICT_NONE)

	ctx, _ := AVFormat_open_device(input, options)
	if ctx != nil {
		AVFormat_find_stream_info(ctx, nil)
		AVFormat_close_input(ctx)
	}

	mu.Lock()
	defer mu.Unlock()

	var devices []AVFoundationDevice
	var currentType AVMediaType
	haveType := false
	defaultSeen := make(map[AVMediaType]bool)

	for _, line := range capturedLines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		if strings.Contains(lower, "video devices:") {
			currentType, haveType = AVMEDIA_TYPE_VIDEO, true
			continue
		}
		if strings.Contains(lower, "audio devices:") {
			currentType, haveType = AVMEDIA_TYPE_AUDIO, true
			continue
		}
		if !haveType {
			continue
		}

		matches := avfoundationDevicePattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		name := strings.TrimSpace(matches[2])
		if name == "" {
			continue
		}

		index := 0
		if n, err := strconv.Atoi(matches[1]); err == nil {
			index = n
		}
		isDefault := index == 0 && !defaultSeen[currentType]
		if isDefault {
			defaultSeen[currentType] = true
		}

		devices = append(devices, AVFoundationDevice{
			Index:     index,
			Name:      name,
			MediaType: currentType,
			IsDefault: isDefault,
		})
	}

	return devices
}

////////////////////////////////////////////////////////////////////////////////
// FALLBACKS
//
// FFmpeg's audiotoolbox muxer fails the "sanity checks" needed to report
// AVERROR(ENOSYS) gracefully - it doesn't implement the get_device_list
// callback that avdevice_list_output_sinks relies on. AVDevice_list_output_sinks
// falls back to this platform-specific implementation instead, so callers see
// a uniform API regardless.
//
// avfoundation has the same problem on the input side, but its devices are
// listed via AVFoundationListDevices above instead of a fallback here - see
// its doc comment for why.

// avDeviceListInputSourcesFallback has no darwin-specific input formats left
// to handle: avfoundation is listed via AVFoundationListDevices instead.
func avDeviceListInputSourcesFallback(device *AVInputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	return nil, nil
}

// avDeviceListOutputSinksFallback lists audiotoolbox output devices directly
// via CoreAudio, since the audiotoolbox muxer only logs its device list as a
// side effect of actually starting an AudioQueue for playback.
func avDeviceListOutputSinksFallback(device *AVOutputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	if device == nil || device.Name() != "audiotoolbox" {
		return nil, nil
	}

	list := C.ff_device_list_new()
	if list == nil {
		return nil, nil
	}
	if ret := C.ff_coreaudio_list_output_devices(list); ret < 0 {
		C.ff_device_list_free(list)
		return nil, AVError(int(ret))
	}
	if list.nb_devices == 0 {
		C.ff_device_list_free(list)
		return nil, nil
	}
	return (*AVDeviceInfoList)(list), nil
}
