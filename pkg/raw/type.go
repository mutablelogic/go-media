package raw

import (
	"mime"
	"regexp"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// rawTypes maps RAW camera file extensions to their MIME types.
// All use the image/x-* convention; IANA has no official registrations for
// most proprietary RAW formats.
var rawTypes = map[string]string{
	// Adobe / generic
	".dng": "image/x-adobe-dng", // Canon, Leica, Ricoh, DJI, Google Pixel, …

	// Canon
	".cr2": "image/x-canon-cr2",
	".cr3": "image/x-canon-cr3",
	".crw": "image/x-canon-crw",

	// Nikon
	".nef": "image/x-nikon-nef",
	".nrw": "image/x-nikon-nrw",

	// Sony
	".arw": "image/x-sony-arw",
	".srf": "image/x-sony-srf",
	".sr2": "image/x-sony-sr2",

	// Olympus
	".orf": "image/x-olympus-orf",

	// Panasonic
	".rw2": "image/x-panasonic-rw2",

	// Fujifilm
	".raf": "image/x-fuji-raf",

	// Pentax / Ricoh
	".pef": "image/x-pentax-pef",

	// Leica
	".rwl": "image/x-leica-rwl",

	// Sigma
	".x3f": "image/x-sigma-x3f",

	// Samsung
	".srw": "image/x-samsung-srw",

	// Minolta / Konica Minolta
	".mrw": "image/x-minolta-mrw",

	// Epson
	".erf": "image/x-epson-erf",

	// Kodak
	".dcr": "image/x-kodak-dcr",
	".kdc": "image/x-kodak-kdc",

	// Mamiya
	".mef": "image/x-mamiya-mef",

	// Hasselblad
	".3fr": "image/x-hasselblad-3fr",
	".fff": "image/x-hasselblad-fff",

	// Phase One
	".iiq": "image/x-phaseone-iiq",

	// Leaf
	".mos": "image/x-leaf-mos",

	// Casio
	".bay": "image/x-casio-bay",

	// Generic / ambiguous (Contax, Panasonic, others)
	".raw": "image/x-raw",
}

var ContentTypes = regexp.MustCompile("^image/x-(adobe-dng|canon-cr2|canon-cr3|canon-crw|nikon-nef|nikon-nrw|sony-arw|sony-srf|sony-sr2|olympus-orf|panasonic-rw2|fuji-raf|pentax-pef|leica-rwl|sigma-x3f|samsung-srw|minolta-mrw|epson-erf|kodak-dcr|kodak-kdc|mamiya-mef|hasselblad-3fr|hasselblad-fff|phaseone-iiq|leaf-mos|casio-bay|raw)$")

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	for ext, typ := range rawTypes {
		_ = mime.AddExtensionType(ext, typ)
	}
}
