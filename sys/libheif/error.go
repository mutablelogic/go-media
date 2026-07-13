package libheif

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libheif
#include <libheif/heif_error.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	HeifErrorCode    int
	HeifSuberrorCode int
)

type HeifError struct {
	Code    HeifErrorCode
	Subcode HeifSuberrorCode
	Message string
}

////////////////////////////////////////////////////////////////////////////////
// CONSTS - heif_error_code

const (
	HEIF_ERROR_OK                           HeifErrorCode = C.heif_error_Ok
	HEIF_ERROR_INPUT_DOES_NOT_EXIST         HeifErrorCode = C.heif_error_Input_does_not_exist
	HEIF_ERROR_INVALID_INPUT                HeifErrorCode = C.heif_error_Invalid_input
	HEIF_ERROR_UNSUPPORTED_FILETYPE         HeifErrorCode = C.heif_error_Unsupported_filetype
	HEIF_ERROR_UNSUPPORTED_FEATURE          HeifErrorCode = C.heif_error_Unsupported_feature
	HEIF_ERROR_USAGE_ERROR                  HeifErrorCode = C.heif_error_Usage_error
	HEIF_ERROR_MEMORY_ALLOCATION_ERROR      HeifErrorCode = C.heif_error_Memory_allocation_error
	HEIF_ERROR_DECODER_PLUGIN_ERROR         HeifErrorCode = C.heif_error_Decoder_plugin_error
	HEIF_ERROR_ENCODER_PLUGIN_ERROR         HeifErrorCode = C.heif_error_Encoder_plugin_error
	HEIF_ERROR_ENCODING_ERROR               HeifErrorCode = C.heif_error_Encoding_error
	HEIF_ERROR_COLOR_PROFILE_DOES_NOT_EXIST HeifErrorCode = C.heif_error_Color_profile_does_not_exist
	HEIF_ERROR_PLUGIN_LOADING_ERROR         HeifErrorCode = C.heif_error_Plugin_loading_error
	HEIF_ERROR_CANCELED                     HeifErrorCode = C.heif_error_Canceled
	HEIF_ERROR_END_OF_SEQUENCE              HeifErrorCode = C.heif_error_End_of_sequence
)

////////////////////////////////////////////////////////////////////////////////
// CONSTS - heif_suberror_code

const (
	HEIF_SUBERROR_UNSPECIFIED                              HeifSuberrorCode = C.heif_suberror_Unspecified
	HEIF_SUBERROR_END_OF_DATA                              HeifSuberrorCode = C.heif_suberror_End_of_data
	HEIF_SUBERROR_INVALID_BOX_SIZE                         HeifSuberrorCode = C.heif_suberror_Invalid_box_size
	HEIF_SUBERROR_NO_FTYP_BOX                              HeifSuberrorCode = C.heif_suberror_No_ftyp_box
	HEIF_SUBERROR_NO_IDAT_BOX                              HeifSuberrorCode = C.heif_suberror_No_idat_box
	HEIF_SUBERROR_NO_META_BOX                              HeifSuberrorCode = C.heif_suberror_No_meta_box
	HEIF_SUBERROR_NO_HDLR_BOX                              HeifSuberrorCode = C.heif_suberror_No_hdlr_box
	HEIF_SUBERROR_NO_HVCC_BOX                              HeifSuberrorCode = C.heif_suberror_No_hvcC_box
	HEIF_SUBERROR_NO_PITM_BOX                              HeifSuberrorCode = C.heif_suberror_No_pitm_box
	HEIF_SUBERROR_NO_IPCO_BOX                              HeifSuberrorCode = C.heif_suberror_No_ipco_box
	HEIF_SUBERROR_NO_IPMA_BOX                              HeifSuberrorCode = C.heif_suberror_No_ipma_box
	HEIF_SUBERROR_NO_ILOC_BOX                              HeifSuberrorCode = C.heif_suberror_No_iloc_box
	HEIF_SUBERROR_NO_IINF_BOX                              HeifSuberrorCode = C.heif_suberror_No_iinf_box
	HEIF_SUBERROR_NO_IPRP_BOX                              HeifSuberrorCode = C.heif_suberror_No_iprp_box
	HEIF_SUBERROR_NO_IREF_BOX                              HeifSuberrorCode = C.heif_suberror_No_iref_box
	HEIF_SUBERROR_NO_PICT_HANDLER                          HeifSuberrorCode = C.heif_suberror_No_pict_handler
	HEIF_SUBERROR_IPMA_BOX_REFERENCES_NONEXISTING_PROPERTY HeifSuberrorCode = C.heif_suberror_Ipma_box_references_nonexisting_property
	HEIF_SUBERROR_NO_PROPERTIES_ASSIGNED_TO_ITEM           HeifSuberrorCode = C.heif_suberror_No_properties_assigned_to_item
	HEIF_SUBERROR_NO_ITEM_DATA                             HeifSuberrorCode = C.heif_suberror_No_item_data
	HEIF_SUBERROR_INVALID_GRID_DATA                        HeifSuberrorCode = C.heif_suberror_Invalid_grid_data
	HEIF_SUBERROR_MISSING_GRID_IMAGES                      HeifSuberrorCode = C.heif_suberror_Missing_grid_images
	HEIF_SUBERROR_INVALID_CLEAN_APERTURE                   HeifSuberrorCode = C.heif_suberror_Invalid_clean_aperture
	HEIF_SUBERROR_INVALID_OVERLAY_DATA                     HeifSuberrorCode = C.heif_suberror_Invalid_overlay_data
	HEIF_SUBERROR_OVERLAY_IMAGE_OUTSIDE_OF_CANVAS          HeifSuberrorCode = C.heif_suberror_Overlay_image_outside_of_canvas
	HEIF_SUBERROR_AUXILIARY_IMAGE_TYPE_UNSPECIFIED         HeifSuberrorCode = C.heif_suberror_Auxiliary_image_type_unspecified
	HEIF_SUBERROR_NO_OR_INVALID_PRIMARY_ITEM               HeifSuberrorCode = C.heif_suberror_No_or_invalid_primary_item
	HEIF_SUBERROR_NO_INFE_BOX                              HeifSuberrorCode = C.heif_suberror_No_infe_box
	HEIF_SUBERROR_UNKNOWN_COLOR_PROFILE_TYPE               HeifSuberrorCode = C.heif_suberror_Unknown_color_profile_type
	HEIF_SUBERROR_WRONG_TILE_IMAGE_CHROMA_FORMAT           HeifSuberrorCode = C.heif_suberror_Wrong_tile_image_chroma_format
	HEIF_SUBERROR_INVALID_FRACTIONAL_NUMBER                HeifSuberrorCode = C.heif_suberror_Invalid_fractional_number
	HEIF_SUBERROR_INVALID_IMAGE_SIZE                       HeifSuberrorCode = C.heif_suberror_Invalid_image_size
	HEIF_SUBERROR_INVALID_PIXI_BOX                         HeifSuberrorCode = C.heif_suberror_Invalid_pixi_box
	HEIF_SUBERROR_NO_AV1C_BOX                              HeifSuberrorCode = C.heif_suberror_No_av1C_box
	HEIF_SUBERROR_WRONG_TILE_IMAGE_PIXEL_DEPTH             HeifSuberrorCode = C.heif_suberror_Wrong_tile_image_pixel_depth
	HEIF_SUBERROR_UNKNOWN_NCLX_COLOR_PRIMARIES             HeifSuberrorCode = C.heif_suberror_Unknown_NCLX_color_primaries
	HEIF_SUBERROR_UNKNOWN_NCLX_TRANSFER_CHARACTERISTICS    HeifSuberrorCode = C.heif_suberror_Unknown_NCLX_transfer_characteristics
	HEIF_SUBERROR_UNKNOWN_NCLX_MATRIX_COEFFICIENTS         HeifSuberrorCode = C.heif_suberror_Unknown_NCLX_matrix_coefficients
	HEIF_SUBERROR_INVALID_REGION_DATA                      HeifSuberrorCode = C.heif_suberror_Invalid_region_data
	HEIF_SUBERROR_NO_ISPE_PROPERTY                         HeifSuberrorCode = C.heif_suberror_No_ispe_property
	HEIF_SUBERROR_CAMERA_INTRINSIC_MATRIX_UNDEFINED        HeifSuberrorCode = C.heif_suberror_Camera_intrinsic_matrix_undefined
	HEIF_SUBERROR_CAMERA_EXTRINSIC_MATRIX_UNDEFINED        HeifSuberrorCode = C.heif_suberror_Camera_extrinsic_matrix_undefined
	HEIF_SUBERROR_INVALID_J2K_CODESTREAM                   HeifSuberrorCode = C.heif_suberror_Invalid_J2K_codestream
	HEIF_SUBERROR_NO_VVCC_BOX                              HeifSuberrorCode = C.heif_suberror_No_vvcC_box
	HEIF_SUBERROR_NO_ICBR_BOX                              HeifSuberrorCode = C.heif_suberror_No_icbr_box
	HEIF_SUBERROR_NO_AVCC_BOX                              HeifSuberrorCode = C.heif_suberror_No_avcC_box
	HEIF_SUBERROR_INVALID_MINI_BOX                         HeifSuberrorCode = C.heif_suberror_Invalid_mini_box
	HEIF_SUBERROR_DECOMPRESSION_INVALID_DATA               HeifSuberrorCode = C.heif_suberror_Decompression_invalid_data
	HEIF_SUBERROR_NO_MOOV_BOX                              HeifSuberrorCode = C.heif_suberror_No_moov_box
	HEIF_SUBERROR_NCLX_COLR_VUI_MISMATCH                   HeifSuberrorCode = C.heif_suberror_NCLX_colr_VUI_mismatch
	HEIF_SUBERROR_SECURITY_LIMIT_EXCEEDED                  HeifSuberrorCode = C.heif_suberror_Security_limit_exceeded
	HEIF_SUBERROR_COMPRESSION_INITIALISATION_ERROR         HeifSuberrorCode = C.heif_suberror_Compression_initialisation_error
	HEIF_SUBERROR_NONEXISTING_ITEM_REFERENCED              HeifSuberrorCode = C.heif_suberror_Nonexisting_item_referenced
	HEIF_SUBERROR_NULL_POINTER_ARGUMENT                    HeifSuberrorCode = C.heif_suberror_Null_pointer_argument
	HEIF_SUBERROR_NONEXISTING_IMAGE_CHANNEL_REFERENCED     HeifSuberrorCode = C.heif_suberror_Nonexisting_image_channel_referenced
	HEIF_SUBERROR_UNSUPPORTED_PLUGIN_VERSION               HeifSuberrorCode = C.heif_suberror_Unsupported_plugin_version
	HEIF_SUBERROR_UNSUPPORTED_WRITER_VERSION               HeifSuberrorCode = C.heif_suberror_Unsupported_writer_version
	HEIF_SUBERROR_UNSUPPORTED_PARAMETER                    HeifSuberrorCode = C.heif_suberror_Unsupported_parameter
	HEIF_SUBERROR_INVALID_PARAMETER_VALUE                  HeifSuberrorCode = C.heif_suberror_Invalid_parameter_value
	HEIF_SUBERROR_INVALID_PROPERTY                         HeifSuberrorCode = C.heif_suberror_Invalid_property
	HEIF_SUBERROR_ITEM_REFERENCE_CYCLE                     HeifSuberrorCode = C.heif_suberror_Item_reference_cycle
	HEIF_SUBERROR_UNSUPPORTED_CODEC                        HeifSuberrorCode = C.heif_suberror_Unsupported_codec
	HEIF_SUBERROR_UNSUPPORTED_IMAGE_TYPE                   HeifSuberrorCode = C.heif_suberror_Unsupported_image_type
	HEIF_SUBERROR_UNSUPPORTED_DATA_VERSION                 HeifSuberrorCode = C.heif_suberror_Unsupported_data_version
	HEIF_SUBERROR_UNSUPPORTED_COLOR_CONVERSION             HeifSuberrorCode = C.heif_suberror_Unsupported_color_conversion
	HEIF_SUBERROR_UNSUPPORTED_ITEM_CONSTRUCTION_METHOD     HeifSuberrorCode = C.heif_suberror_Unsupported_item_construction_method
	HEIF_SUBERROR_UNSUPPORTED_HEADER_COMPRESSION_METHOD    HeifSuberrorCode = C.heif_suberror_Unsupported_header_compression_method
	HEIF_SUBERROR_UNSUPPORTED_GENERIC_COMPRESSION_METHOD   HeifSuberrorCode = C.heif_suberror_Unsupported_generic_compression_method
	HEIF_SUBERROR_UNSUPPORTED_ESSENTIAL_PROPERTY           HeifSuberrorCode = C.heif_suberror_Unsupported_essential_property
	HEIF_SUBERROR_UNSUPPORTED_TRACK_TYPE                   HeifSuberrorCode = C.heif_suberror_Unsupported_track_type
	HEIF_SUBERROR_UNSUPPORTED_BIT_DEPTH                    HeifSuberrorCode = C.heif_suberror_Unsupported_bit_depth
	HEIF_SUBERROR_CANNOT_WRITE_OUTPUT_DATA                 HeifSuberrorCode = C.heif_suberror_Cannot_write_output_data
	HEIF_SUBERROR_ENCODER_INITIALIZATION                   HeifSuberrorCode = C.heif_suberror_Encoder_initialization
	HEIF_SUBERROR_ENCODER_ENCODING                         HeifSuberrorCode = C.heif_suberror_Encoder_encoding
	HEIF_SUBERROR_ENCODER_CLEANUP                          HeifSuberrorCode = C.heif_suberror_Encoder_cleanup
	HEIF_SUBERROR_TOO_MANY_REGIONS                         HeifSuberrorCode = C.heif_suberror_Too_many_regions
	HEIF_SUBERROR_PLUGIN_LOADING_ERROR                     HeifSuberrorCode = C.heif_suberror_Plugin_loading_error
	HEIF_SUBERROR_PLUGIN_IS_NOT_LOADED                     HeifSuberrorCode = C.heif_suberror_Plugin_is_not_loaded
	HEIF_SUBERROR_CANNOT_READ_PLUGIN_DIRECTORY             HeifSuberrorCode = C.heif_suberror_Cannot_read_plugin_directory
	HEIF_SUBERROR_NO_MATCHING_DECODER_INSTALLED            HeifSuberrorCode = C.heif_suberror_No_matching_decoder_installed
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - ERROR

func (e HeifError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("libheif error code=%d subcode=%d", e.Code, e.Subcode)
	}
	return e.Message
}

func fromCError(e C.heif_error) HeifError {
	return HeifError{
		Code:    HeifErrorCode(e.code),
		Subcode: HeifSuberrorCode(e.subcode),
		Message: C.GoString(e.message),
	}
}
