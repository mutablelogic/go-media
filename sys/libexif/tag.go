package libexif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-tag.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Tag          C.ExifTag
	IFD          C.ExifIfd
	SupportLevel C.ExifSupportLevel
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS - IFD

const (
	EXIF_IFD_0                IFD = C.EXIF_IFD_0
	EXIF_IFD_1                IFD = C.EXIF_IFD_1
	EXIF_IFD_EXIF             IFD = C.EXIF_IFD_EXIF
	EXIF_IFD_GPS              IFD = C.EXIF_IFD_GPS
	EXIF_IFD_INTEROPERABILITY IFD = C.EXIF_IFD_INTEROPERABILITY
	EXIF_IFD_COUNT            IFD = C.EXIF_IFD_COUNT
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS - SUPPORT LEVEL

const (
	EXIF_SUPPORT_LEVEL_UNKNOWN      SupportLevel = C.EXIF_SUPPORT_LEVEL_UNKNOWN
	EXIF_SUPPORT_LEVEL_NOT_RECORDED SupportLevel = C.EXIF_SUPPORT_LEVEL_NOT_RECORDED
	EXIF_SUPPORT_LEVEL_MANDATORY    SupportLevel = C.EXIF_SUPPORT_LEVEL_MANDATORY
	EXIF_SUPPORT_LEVEL_OPTIONAL     SupportLevel = C.EXIF_SUPPORT_LEVEL_OPTIONAL
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS - TAGS

const (
	EXIF_TAG_INTEROPERABILITY_INDEX                   Tag = C.EXIF_TAG_INTEROPERABILITY_INDEX
	EXIF_TAG_INTEROPERABILITY_VERSION                 Tag = C.EXIF_TAG_INTEROPERABILITY_VERSION
	EXIF_TAG_NEW_SUBFILE_TYPE                         Tag = C.EXIF_TAG_NEW_SUBFILE_TYPE
	EXIF_TAG_IMAGE_WIDTH                              Tag = C.EXIF_TAG_IMAGE_WIDTH
	EXIF_TAG_IMAGE_LENGTH                             Tag = C.EXIF_TAG_IMAGE_LENGTH
	EXIF_TAG_BITS_PER_SAMPLE                          Tag = C.EXIF_TAG_BITS_PER_SAMPLE
	EXIF_TAG_COMPRESSION                              Tag = C.EXIF_TAG_COMPRESSION
	EXIF_TAG_PHOTOMETRIC_INTERPRETATION               Tag = C.EXIF_TAG_PHOTOMETRIC_INTERPRETATION
	EXIF_TAG_FILL_ORDER                               Tag = C.EXIF_TAG_FILL_ORDER
	EXIF_TAG_DOCUMENT_NAME                            Tag = C.EXIF_TAG_DOCUMENT_NAME
	EXIF_TAG_IMAGE_DESCRIPTION                        Tag = C.EXIF_TAG_IMAGE_DESCRIPTION
	EXIF_TAG_MAKE                                     Tag = C.EXIF_TAG_MAKE
	EXIF_TAG_MODEL                                    Tag = C.EXIF_TAG_MODEL
	EXIF_TAG_STRIP_OFFSETS                            Tag = C.EXIF_TAG_STRIP_OFFSETS
	EXIF_TAG_ORIENTATION                              Tag = C.EXIF_TAG_ORIENTATION
	EXIF_TAG_SAMPLES_PER_PIXEL                        Tag = C.EXIF_TAG_SAMPLES_PER_PIXEL
	EXIF_TAG_ROWS_PER_STRIP                           Tag = C.EXIF_TAG_ROWS_PER_STRIP
	EXIF_TAG_STRIP_BYTE_COUNTS                        Tag = C.EXIF_TAG_STRIP_BYTE_COUNTS
	EXIF_TAG_X_RESOLUTION                             Tag = C.EXIF_TAG_X_RESOLUTION
	EXIF_TAG_Y_RESOLUTION                             Tag = C.EXIF_TAG_Y_RESOLUTION
	EXIF_TAG_PLANAR_CONFIGURATION                     Tag = C.EXIF_TAG_PLANAR_CONFIGURATION
	EXIF_TAG_RESOLUTION_UNIT                          Tag = C.EXIF_TAG_RESOLUTION_UNIT
	EXIF_TAG_TRANSFER_FUNCTION                        Tag = C.EXIF_TAG_TRANSFER_FUNCTION
	EXIF_TAG_SOFTWARE                                 Tag = C.EXIF_TAG_SOFTWARE
	EXIF_TAG_DATE_TIME                                Tag = C.EXIF_TAG_DATE_TIME
	EXIF_TAG_ARTIST                                   Tag = C.EXIF_TAG_ARTIST
	EXIF_TAG_WHITE_POINT                              Tag = C.EXIF_TAG_WHITE_POINT
	EXIF_TAG_PRIMARY_CHROMATICITIES                   Tag = C.EXIF_TAG_PRIMARY_CHROMATICITIES
	EXIF_TAG_SUB_IFDS                                 Tag = C.EXIF_TAG_SUB_IFDS
	EXIF_TAG_TRANSFER_RANGE                           Tag = C.EXIF_TAG_TRANSFER_RANGE
	EXIF_TAG_JPEG_PROC                                Tag = C.EXIF_TAG_JPEG_PROC
	EXIF_TAG_JPEG_INTERCHANGE_FORMAT                  Tag = C.EXIF_TAG_JPEG_INTERCHANGE_FORMAT
	EXIF_TAG_JPEG_INTERCHANGE_FORMAT_LENGTH           Tag = C.EXIF_TAG_JPEG_INTERCHANGE_FORMAT_LENGTH
	EXIF_TAG_YCBCR_COEFFICIENTS                       Tag = C.EXIF_TAG_YCBCR_COEFFICIENTS
	EXIF_TAG_YCBCR_SUB_SAMPLING                       Tag = C.EXIF_TAG_YCBCR_SUB_SAMPLING
	EXIF_TAG_YCBCR_POSITIONING                        Tag = C.EXIF_TAG_YCBCR_POSITIONING
	EXIF_TAG_REFERENCE_BLACK_WHITE                    Tag = C.EXIF_TAG_REFERENCE_BLACK_WHITE
	EXIF_TAG_XML_PACKET                               Tag = C.EXIF_TAG_XML_PACKET
	EXIF_TAG_RELATED_IMAGE_FILE_FORMAT                Tag = C.EXIF_TAG_RELATED_IMAGE_FILE_FORMAT
	EXIF_TAG_RELATED_IMAGE_WIDTH                      Tag = C.EXIF_TAG_RELATED_IMAGE_WIDTH
	EXIF_TAG_RELATED_IMAGE_LENGTH                     Tag = C.EXIF_TAG_RELATED_IMAGE_LENGTH
	EXIF_TAG_IMAGE_DEPTH                              Tag = C.EXIF_TAG_IMAGE_DEPTH
	EXIF_TAG_CFA_REPEAT_PATTERN_DIM                   Tag = C.EXIF_TAG_CFA_REPEAT_PATTERN_DIM
	EXIF_TAG_CFA_PATTERN                              Tag = C.EXIF_TAG_CFA_PATTERN
	EXIF_TAG_BATTERY_LEVEL                            Tag = C.EXIF_TAG_BATTERY_LEVEL
	EXIF_TAG_COPYRIGHT                                Tag = C.EXIF_TAG_COPYRIGHT
	EXIF_TAG_EXPOSURE_TIME                            Tag = C.EXIF_TAG_EXPOSURE_TIME
	EXIF_TAG_FNUMBER                                  Tag = C.EXIF_TAG_FNUMBER
	EXIF_TAG_IPTC_NAA                                 Tag = C.EXIF_TAG_IPTC_NAA
	EXIF_TAG_IMAGE_RESOURCES                          Tag = C.EXIF_TAG_IMAGE_RESOURCES
	EXIF_TAG_EXIF_IFD_POINTER                         Tag = C.EXIF_TAG_EXIF_IFD_POINTER
	EXIF_TAG_INTER_COLOR_PROFILE                      Tag = C.EXIF_TAG_INTER_COLOR_PROFILE
	EXIF_TAG_EXPOSURE_PROGRAM                         Tag = C.EXIF_TAG_EXPOSURE_PROGRAM
	EXIF_TAG_SPECTRAL_SENSITIVITY                     Tag = C.EXIF_TAG_SPECTRAL_SENSITIVITY
	EXIF_TAG_GPS_INFO_IFD_POINTER                     Tag = C.EXIF_TAG_GPS_INFO_IFD_POINTER
	EXIF_TAG_ISO_SPEED_RATINGS                        Tag = C.EXIF_TAG_ISO_SPEED_RATINGS
	EXIF_TAG_OECF                                     Tag = C.EXIF_TAG_OECF
	EXIF_TAG_TIME_ZONE_OFFSET                         Tag = C.EXIF_TAG_TIME_ZONE_OFFSET
	EXIF_TAG_SENSITIVITY_TYPE                         Tag = C.EXIF_TAG_SENSITIVITY_TYPE
	EXIF_TAG_STANDARD_OUTPUT_SENSITIVITY              Tag = C.EXIF_TAG_STANDARD_OUTPUT_SENSITIVITY
	EXIF_TAG_RECOMMENDED_EXPOSURE_INDEX               Tag = C.EXIF_TAG_RECOMMENDED_EXPOSURE_INDEX
	EXIF_TAG_ISO_SPEED                                Tag = C.EXIF_TAG_ISO_SPEED
	EXIF_TAG_ISO_SPEEDLatitudeYYY                     Tag = C.EXIF_TAG_ISO_SPEEDLatitudeYYY
	EXIF_TAG_ISO_SPEEDLatitudeZZZ                     Tag = C.EXIF_TAG_ISO_SPEEDLatitudeZZZ
	EXIF_TAG_EXIF_VERSION                             Tag = C.EXIF_TAG_EXIF_VERSION
	EXIF_TAG_DATE_TIME_ORIGINAL                       Tag = C.EXIF_TAG_DATE_TIME_ORIGINAL
	EXIF_TAG_DATE_TIME_DIGITIZED                      Tag = C.EXIF_TAG_DATE_TIME_DIGITIZED
	EXIF_TAG_OFFSET_TIME                              Tag = C.EXIF_TAG_OFFSET_TIME
	EXIF_TAG_OFFSET_TIME_ORIGINAL                     Tag = C.EXIF_TAG_OFFSET_TIME_ORIGINAL
	EXIF_TAG_OFFSET_TIME_DIGITIZED                    Tag = C.EXIF_TAG_OFFSET_TIME_DIGITIZED
	EXIF_TAG_COMPONENTS_CONFIGURATION                 Tag = C.EXIF_TAG_COMPONENTS_CONFIGURATION
	EXIF_TAG_COMPRESSED_BITS_PER_PIXEL                Tag = C.EXIF_TAG_COMPRESSED_BITS_PER_PIXEL
	EXIF_TAG_SHUTTER_SPEED_VALUE                      Tag = C.EXIF_TAG_SHUTTER_SPEED_VALUE
	EXIF_TAG_APERTURE_VALUE                           Tag = C.EXIF_TAG_APERTURE_VALUE
	EXIF_TAG_BRIGHTNESS_VALUE                         Tag = C.EXIF_TAG_BRIGHTNESS_VALUE
	EXIF_TAG_EXPOSURE_BIAS_VALUE                      Tag = C.EXIF_TAG_EXPOSURE_BIAS_VALUE
	EXIF_TAG_MAX_APERTURE_VALUE                       Tag = C.EXIF_TAG_MAX_APERTURE_VALUE
	EXIF_TAG_SUBJECT_DISTANCE                         Tag = C.EXIF_TAG_SUBJECT_DISTANCE
	EXIF_TAG_METERING_MODE                            Tag = C.EXIF_TAG_METERING_MODE
	EXIF_TAG_LIGHT_SOURCE                             Tag = C.EXIF_TAG_LIGHT_SOURCE
	EXIF_TAG_FLASH                                    Tag = C.EXIF_TAG_FLASH
	EXIF_TAG_FOCAL_LENGTH                             Tag = C.EXIF_TAG_FOCAL_LENGTH
	EXIF_TAG_SUBJECT_AREA                             Tag = C.EXIF_TAG_SUBJECT_AREA
	EXIF_TAG_TIFF_EP_STANDARD_ID                      Tag = C.EXIF_TAG_TIFF_EP_STANDARD_ID
	EXIF_TAG_MAKER_NOTE                               Tag = C.EXIF_TAG_MAKER_NOTE
	EXIF_TAG_USER_COMMENT                             Tag = C.EXIF_TAG_USER_COMMENT
	EXIF_TAG_SUB_SEC_TIME                             Tag = C.EXIF_TAG_SUB_SEC_TIME
	EXIF_TAG_SUB_SEC_TIME_ORIGINAL                    Tag = C.EXIF_TAG_SUB_SEC_TIME_ORIGINAL
	EXIF_TAG_SUB_SEC_TIME_DIGITIZED                   Tag = C.EXIF_TAG_SUB_SEC_TIME_DIGITIZED
	EXIF_TAG_XP_TITLE                                 Tag = C.EXIF_TAG_XP_TITLE
	EXIF_TAG_XP_COMMENT                               Tag = C.EXIF_TAG_XP_COMMENT
	EXIF_TAG_XP_AUTHOR                                Tag = C.EXIF_TAG_XP_AUTHOR
	EXIF_TAG_XP_KEYWORDS                              Tag = C.EXIF_TAG_XP_KEYWORDS
	EXIF_TAG_XP_SUBJECT                               Tag = C.EXIF_TAG_XP_SUBJECT
	EXIF_TAG_FLASH_PIX_VERSION                        Tag = C.EXIF_TAG_FLASH_PIX_VERSION
	EXIF_TAG_COLOR_SPACE                              Tag = C.EXIF_TAG_COLOR_SPACE
	EXIF_TAG_PIXEL_X_DIMENSION                        Tag = C.EXIF_TAG_PIXEL_X_DIMENSION
	EXIF_TAG_PIXEL_Y_DIMENSION                        Tag = C.EXIF_TAG_PIXEL_Y_DIMENSION
	EXIF_TAG_RELATED_SOUND_FILE                       Tag = C.EXIF_TAG_RELATED_SOUND_FILE
	EXIF_TAG_INTEROPERABILITY_IFD_POINTER             Tag = C.EXIF_TAG_INTEROPERABILITY_IFD_POINTER
	EXIF_TAG_FLASH_ENERGY                             Tag = C.EXIF_TAG_FLASH_ENERGY
	EXIF_TAG_SPATIAL_FREQUENCY_RESPONSE               Tag = C.EXIF_TAG_SPATIAL_FREQUENCY_RESPONSE
	EXIF_TAG_FOCAL_PLANE_X_RESOLUTION                 Tag = C.EXIF_TAG_FOCAL_PLANE_X_RESOLUTION
	EXIF_TAG_FOCAL_PLANE_Y_RESOLUTION                 Tag = C.EXIF_TAG_FOCAL_PLANE_Y_RESOLUTION
	EXIF_TAG_FOCAL_PLANE_RESOLUTION_UNIT              Tag = C.EXIF_TAG_FOCAL_PLANE_RESOLUTION_UNIT
	EXIF_TAG_SUBJECT_LOCATION                         Tag = C.EXIF_TAG_SUBJECT_LOCATION
	EXIF_TAG_EXPOSURE_INDEX                           Tag = C.EXIF_TAG_EXPOSURE_INDEX
	EXIF_TAG_SENSING_METHOD                           Tag = C.EXIF_TAG_SENSING_METHOD
	EXIF_TAG_FILE_SOURCE                              Tag = C.EXIF_TAG_FILE_SOURCE
	EXIF_TAG_SCENE_TYPE                               Tag = C.EXIF_TAG_SCENE_TYPE
	EXIF_TAG_NEW_CFA_PATTERN                          Tag = C.EXIF_TAG_NEW_CFA_PATTERN
	EXIF_TAG_CUSTOM_RENDERED                          Tag = C.EXIF_TAG_CUSTOM_RENDERED
	EXIF_TAG_EXPOSURE_MODE                            Tag = C.EXIF_TAG_EXPOSURE_MODE
	EXIF_TAG_WHITE_BALANCE                            Tag = C.EXIF_TAG_WHITE_BALANCE
	EXIF_TAG_DIGITAL_ZOOM_RATIO                       Tag = C.EXIF_TAG_DIGITAL_ZOOM_RATIO
	EXIF_TAG_FOCAL_LENGTH_IN_35MM_FILM                Tag = C.EXIF_TAG_FOCAL_LENGTH_IN_35MM_FILM
	EXIF_TAG_SCENE_CAPTURE_TYPE                       Tag = C.EXIF_TAG_SCENE_CAPTURE_TYPE
	EXIF_TAG_GAIN_CONTROL                             Tag = C.EXIF_TAG_GAIN_CONTROL
	EXIF_TAG_CONTRAST                                 Tag = C.EXIF_TAG_CONTRAST
	EXIF_TAG_SATURATION                               Tag = C.EXIF_TAG_SATURATION
	EXIF_TAG_SHARPNESS                                Tag = C.EXIF_TAG_SHARPNESS
	EXIF_TAG_DEVICE_SETTING_DESCRIPTION               Tag = C.EXIF_TAG_DEVICE_SETTING_DESCRIPTION
	EXIF_TAG_SUBJECT_DISTANCE_RANGE                   Tag = C.EXIF_TAG_SUBJECT_DISTANCE_RANGE
	EXIF_TAG_IMAGE_UNIQUE_ID                          Tag = C.EXIF_TAG_IMAGE_UNIQUE_ID
	EXIF_TAG_CAMERA_OWNER_NAME                        Tag = C.EXIF_TAG_CAMERA_OWNER_NAME
	EXIF_TAG_BODY_SERIAL_NUMBER                       Tag = C.EXIF_TAG_BODY_SERIAL_NUMBER
	EXIF_TAG_LENS_SPECIFICATION                       Tag = C.EXIF_TAG_LENS_SPECIFICATION
	EXIF_TAG_LENS_MAKE                                Tag = C.EXIF_TAG_LENS_MAKE
	EXIF_TAG_LENS_MODEL                               Tag = C.EXIF_TAG_LENS_MODEL
	EXIF_TAG_LENS_SERIAL_NUMBER                       Tag = C.EXIF_TAG_LENS_SERIAL_NUMBER
	EXIF_TAG_COMPOSITE_IMAGE                          Tag = C.EXIF_TAG_COMPOSITE_IMAGE
	EXIF_TAG_SOURCE_IMAGE_NUMBER_OF_COMPOSITE_IMAGE   Tag = C.EXIF_TAG_SOURCE_IMAGE_NUMBER_OF_COMPOSITE_IMAGE
	EXIF_TAG_SOURCE_EXPOSURE_TIMES_OF_COMPOSITE_IMAGE Tag = C.EXIF_TAG_SOURCE_EXPOSURE_TIMES_OF_COMPOSITE_IMAGE
	EXIF_TAG_GAMMA                                    Tag = C.EXIF_TAG_GAMMA
	EXIF_TAG_PRINT_IMAGE_MATCHING                     Tag = C.EXIF_TAG_PRINT_IMAGE_MATCHING
	EXIF_TAG_PADDING                                  Tag = C.EXIF_TAG_PADDING
)

// GPS tags share the same numeric space as regular tags but are only
// meaningful in the GPS IFD. Defined as #define in the header, not enum.
const (
	EXIF_TAG_GPS_VERSION_ID          Tag = 0x0000
	EXIF_TAG_GPS_LATITUDE_REF        Tag = 0x0001
	EXIF_TAG_GPS_LATITUDE            Tag = 0x0002
	EXIF_TAG_GPS_LONGITUDE_REF       Tag = 0x0003
	EXIF_TAG_GPS_LONGITUDE           Tag = 0x0004
	EXIF_TAG_GPS_ALTITUDE_REF        Tag = 0x0005
	EXIF_TAG_GPS_ALTITUDE            Tag = 0x0006
	EXIF_TAG_GPS_TIME_STAMP          Tag = 0x0007
	EXIF_TAG_GPS_SATELLITES          Tag = 0x0008
	EXIF_TAG_GPS_STATUS              Tag = 0x0009
	EXIF_TAG_GPS_MEASURE_MODE        Tag = 0x000a
	EXIF_TAG_GPS_DOP                 Tag = 0x000b
	EXIF_TAG_GPS_SPEED_REF           Tag = 0x000c
	EXIF_TAG_GPS_SPEED               Tag = 0x000d
	EXIF_TAG_GPS_TRACK_REF           Tag = 0x000e
	EXIF_TAG_GPS_TRACK               Tag = 0x000f
	EXIF_TAG_GPS_IMG_DIRECTION_REF   Tag = 0x0010
	EXIF_TAG_GPS_IMG_DIRECTION       Tag = 0x0011
	EXIF_TAG_GPS_MAP_DATUM           Tag = 0x0012
	EXIF_TAG_GPS_DEST_LATITUDE_REF   Tag = 0x0013
	EXIF_TAG_GPS_DEST_LATITUDE       Tag = 0x0014
	EXIF_TAG_GPS_DEST_LONGITUDE_REF  Tag = 0x0015
	EXIF_TAG_GPS_DEST_LONGITUDE      Tag = 0x0016
	EXIF_TAG_GPS_DEST_BEARING_REF    Tag = 0x0017
	EXIF_TAG_GPS_DEST_BEARING        Tag = 0x0018
	EXIF_TAG_GPS_DEST_DISTANCE_REF   Tag = 0x0019
	EXIF_TAG_GPS_DEST_DISTANCE       Tag = 0x001a
	EXIF_TAG_GPS_PROCESSING_METHOD   Tag = 0x001b
	EXIF_TAG_GPS_AREA_INFORMATION    Tag = 0x001c
	EXIF_TAG_GPS_DATE_STAMP          Tag = 0x001d
	EXIF_TAG_GPS_DIFFERENTIAL        Tag = 0x001e
	EXIF_TAG_GPS_H_POSITIONING_ERROR Tag = 0x001f
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - IFD

func Exif_ifd_get_name(ifd IFD) string {
	return C.GoString(C.exif_ifd_get_name(C.ExifIfd(ifd)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - TAG

func Exif_tag_from_name(name string) Tag {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return Tag(C.exif_tag_from_name(cname))
}

func Exif_tag_get_name_in_ifd(tag Tag, ifd IFD) string {
	return C.GoString(C.exif_tag_get_name_in_ifd(C.ExifTag(tag), C.ExifIfd(ifd)))
}

func Exif_tag_get_title_in_ifd(tag Tag, ifd IFD) string {
	return C.GoString(C.exif_tag_get_title_in_ifd(C.ExifTag(tag), C.ExifIfd(ifd)))
}

func Exif_tag_get_description_in_ifd(tag Tag, ifd IFD) string {
	return C.GoString(C.exif_tag_get_description_in_ifd(C.ExifTag(tag), C.ExifIfd(ifd)))
}

func Exif_tag_get_support_level_in_ifd(tag Tag, ifd IFD, dtype DataType) SupportLevel {
	return SupportLevel(C.exif_tag_get_support_level_in_ifd(C.ExifTag(tag), C.ExifIfd(ifd), C.ExifDataType(dtype)))
}

func Exif_tag_table_count() uint {
	return uint(C.exif_tag_table_count())
}

func Exif_tag_table_get_tag(n uint) Tag {
	return Tag(C.exif_tag_table_get_tag(C.uint(n)))
}

func Exif_tag_table_get_name(n uint) string {
	return C.GoString(C.exif_tag_table_get_name(C.uint(n)))
}
