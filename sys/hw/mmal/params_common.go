/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

import "github.com/djthorpe/mmal"

////////////////////////////////////////////////////////////////////////////////
// PARAMETERS

const (
	MMAL_PARAM_NONE                 paramType = iota
	MMAL_PARAM_SUPPORTED_ENCODINGS            // Takes a MMAL_PARAMETER_ENCODING_T
	MMAL_PARAM_URI                            // Takes a MMAL_PARAMETER_URI_T
	MMAL_PARAM_CHANGE_EVENT_REQUEST           // Takes a MMAL_PARAMETER_CHANGE_EVENT_REQUEST_T
	MMAL_PARAM_ZERO_COPY                      // Takes a MMAL_PARAMETER_BOOLEAN_T
	MMAL_PARAM_BUFFER_REQUIREMENTS            // Takes a MMAL_PARAMETER_BUFFER_REQUIREMENTS_T
	MMAL_PARAM_STATISTICS                     // Takes a MMAL_PARAMETER_STATISTICS_T
	MMAL_PARAM_CORE_STATISTICS                // Takes a MMAL_PARAMETER_CORE_STATISTICS_T
	MMAL_PARAM_MEM_USAGE                      // Takes a MMAL_PARAMETER_MEM_USAGE_T
	MMAL_PARAM_BUFFER_FLAG_FILTER             // Takes a MMAL_PARAMETER_UINT32_T
	MMAL_PARAM_SEEK                           // Takes a MMAL_PARAMETER_SEEK_T
	MMAL_PARAM_POWERMON_ENABLE                // Takes a MMAL_PARAMETER_BOOLEAN_T
	MMAL_PARAM_LOGGING                        // Takes a MMAL_PARAMETER_LOGGING_T
	MMAL_PARAM_SYSTEM_TIME                    // Takes a MMAL_PARAMETER_UINT64_T
	MMAL_PARAM_NO_IMAGE_PADDING               // Takes a MMAL_PARAMETER_BOOLEAN_T
	MMAL_PARAM_LOCKSTEP_ENABLE                // Takes a MMAL_PARAMETER_BOOLEAN_T
)

////////////////////////////////////////////////////////////////////////////////
// GET PARAMETERS

func (this *port) SupportedEncodings() ([]mmal.EncodingType, error) {
	handle, err := mmal_port_param_alloc_get(this.handle, uint32(MMAL_PARAM_SUPPORTED_ENCODINGS), 0)
	if err != nil {
		return nil, err
	}
	defer mmal_port_param_free(handle)

	// Make an array of encoding types
	encodings := make([]mmal.EncodingType, 0)
	for _, encoding := range mmal_param_get_array_uint32(handle) {
		encodings = append(encodings, mmal.EncodingType(encoding))
	}

	return encodings, nil
}

func (this *port) ZeroCopy() (bool, error) {
	return mmal_port_param_get_bool(this.handle, uint32(MMAL_PARAM_ZERO_COPY))
}

func (this *port) PowerMonEnable() (bool, error) {
	return mmal_port_param_get_bool(this.handle, uint32(MMAL_PARAM_POWERMON_ENABLE))
}

func (this *port) NoImagePadding() (bool, error) {
	return mmal_port_param_get_bool(this.handle, uint32(MMAL_PARAM_NO_IMAGE_PADDING))
}

func (this *port) LockstepEnable() (bool, error) {
	return mmal_port_param_get_bool(this.handle, uint32(MMAL_PARAM_LOCKSTEP_ENABLE))
}

////////////////////////////////////////////////////////////////////////////////
// SET PARAMETERS

func (this *port) SetZeroCopy(value bool) error {
	return mmal_port_param_set_bool(this.handle, uint32(MMAL_PARAM_ZERO_COPY), value)
}

func (this *port) SetPowerMonEnable(value bool) error {
	return mmal_port_param_set_bool(this.handle, uint32(MMAL_PARAM_POWERMON_ENABLE), value)
}

func (this *port) SetNoImagePadding(value bool) error {
	return mmal_port_param_set_bool(this.handle, uint32(MMAL_PARAM_NO_IMAGE_PADDING), value)
}

func (this *port) SetLockstepEnable(value bool) error {
	return mmal_port_param_set_bool(this.handle, uint32(MMAL_PARAM_LOCKSTEP_ENABLE), value)
}
