package fontcache

import (
	"embed"
	"fmt"
	"io/fs"
	"sync"

	// Packages
	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
)

//go:embed *
var fontfs embed.FS

type fontcache struct {
	sync.Mutex
	fs fs.FS
}

func NewFontCache() draw2d.FontCache {
	return &fontcache{
		fs: fontfs,
	}
}

func (f *fontcache) Load(data draw2d.FontData) (*truetype.Font, error) {
	f.Lock()
	defer f.Unlock()

	fontFilename := fmt.Sprintf("%s/%s-%s%s.ttf", fontToName(data), fontToName(data), fontToWeight(data), fontToAccent(data))
	font, err := f.fs.(embed.FS).ReadFile(fontFilename)
	if err != nil {
		return nil, err
	}

	// Load font using truetype
	tt, err := truetype.Parse(font)
	if err != nil {
		return nil, err
	}

	// Return success
	return tt, nil
}

func (f *fontcache) Store(data draw2d.FontData, font *truetype.Font) {
	// Not implemented
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func fontToName(data draw2d.FontData) string {
	switch data.Family {
	case draw2d.FontFamilySerif:
		return data.Name + "Serif"
	case draw2d.FontFamilyMono:
		return data.Name + "Mono"
	default:
		return data.Name + "Sans"
	}
}

func fontToWeight(data draw2d.FontData) string {
	switch data.Style {
	case draw2d.FontStyleBold:
		return "Bold"
	case draw2d.FontStyleItalic:
		return ""
	case draw2d.FontStyleNormal:
		return "Regular"
	default:
		return ""
	}
}

func fontToAccent(data draw2d.FontData) string {
	switch data.Style {
	case draw2d.FontStyleItalic:
		return "Italic"
	default:
		return ""
	}
}
