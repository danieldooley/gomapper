package gomapper

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
)

/*
	RawTile
 */

type RawTile struct {
	data []byte
	format string
	xyz XYzoom
}

func (rt RawTile) toImage() (ImageTile, error) {
	var out ImageTile
	switch rt.format {
	case "png":
		var b bytes.Buffer
		_, err := b.Write(rt.data)
		if err != nil {
			return out, fmt.Errorf("failed to buffer RawTile data: %v", err)
		}
		img, err := png.Decode(&b)
		if err != nil {
			return out, fmt.Errorf("failed to decode RawTile as png: %v", err)
		}
		return ImageTile{
			img: img,
			xyz: rt.xyz,
		}, nil
	default:
		return out, fmt.Errorf("toImage cannot handle format %v", rt.format)
	}
}

type ImageTile struct {
	img image.Image
	xyz XYzoom
}

