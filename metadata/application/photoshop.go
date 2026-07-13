package application

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"mime"
	"regexp"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	imagemeta "github.com/mutablelogic/go-media/metadata/image"
	xmp "github.com/mutablelogic/go-media/pkg/xmp"
	psd "github.com/oov/psd"
)

const photoshopXMPResourceID = 0x0424

type meta struct {
	key   string
	value any
}

func (m meta) Key() string        { return m.key }
func (m meta) Bytes() []byte      { return nil }
func (m meta) Image() image.Image { return nil }
func (m meta) Any() any           { return m.value }

func (m meta) Value() string {
	return fmt.Sprint(m.value)
}

func init() {
	mime.AddExtensionType(".psd", "application/vnd.adobe.photoshop")
	mime.AddExtensionType(".psb", "application/vnd.adobe.photoshop")

	metadata.AddHandler(regexp.MustCompile(`^(?:application|image)/(?:vnd\.adobe\.photoshop|photoshop|x-photoshop)$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		cfg, _, err := psd.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		return photoshopMetadata(cfg, filter)
	}, "photoshop", "xmp")

	metadata.AddHandler(regexp.MustCompile(`^(?:application|image)/(?:vnd\.adobe\.photoshop|photoshop|x-photoshop)$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		if filter != "artwork:" && filter != "artwork:thumbnail" {
			return nil, nil
		}

		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		m, err := imagemeta.ExtractArtwork(data, "artwork:thumbnail")
		if err != nil {
			return nil, err
		}

		return []gomedia.Metadata{m}, nil
	}, "artwork")
}

func photoshopMetadata(cfg psd.Config, filter string) ([]gomedia.Metadata, error) {
	entries := map[string]gomedia.Metadata{
		"photoshop:Format":    meta{key: "photoshop:Format", value: photoshopFormat(cfg.Version)},
		"photoshop:Version":   meta{key: "photoshop:Version", value: cfg.Version},
		"photoshop:Width":     meta{key: "photoshop:Width", value: cfg.Rect.Dx()},
		"photoshop:Height":    meta{key: "photoshop:Height", value: cfg.Rect.Dy()},
		"photoshop:Channels":  meta{key: "photoshop:Channels", value: cfg.Channels},
		"photoshop:Depth":     meta{key: "photoshop:Depth", value: cfg.Depth},
		"photoshop:ColorMode": meta{key: "photoshop:ColorMode", value: cfg.ColorMode},
	}

	if res, ok := cfg.Res[photoshopXMPResourceID]; ok && len(res.Data) > 0 {
		if doc, err := xmp.Parse(res.Data); err == nil {
			for _, item := range doc.Items() {
				entries[item.Key()] = item
			}
		}
	}

	return metadata.FilterMetadata(entries, filter), nil
}

func photoshopFormat(version int) string {
	switch version {
	case 2:
		return "PSB"
	default:
		return "PSD"
	}
}
