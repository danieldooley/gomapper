package gomapper

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"path"
	"strings"
	"github.com/nfnt/resize"
)

const (
	commmonMarkerWidth = 25
	commonMarkerHeight = 41
)

/*
	Loads markerimage from file and restricts to common marker size
 */
func LoadMarkerImage(filepath string) (image.Image, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not load image from filepath '%v': %v", filepath, err)
	}

	baseSplit := strings.Split(path.Base(filepath), ".")
	extension := baseSplit[len(baseSplit)-1]

	var img image.Image

	switch extension {
	case "png":
		var buf bytes.Buffer
		_, err := buf.Write(b)
		if err != nil {
			return nil, fmt.Errorf("failed to buffer image data: %v", err)
		}
		img, err = png.Decode(&buf)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image as png: %v", err)
		}
	default:
		return nil, fmt.Errorf("cannot handle extension '%v'", extension)
	}

	return resize.Resize(commmonMarkerWidth, commonMarkerHeight, img, resize.NearestNeighbor), nil
}
