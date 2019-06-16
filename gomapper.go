package gomapper

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
)

type Layer struct {
	name string
	img image.Image
}

type GoMap struct {
	llb LonLatBounds
	layers []Layer
	baselayer image.Image
}

func NewMapFromBounds(llb LonLatBounds) GoMap {
	gm := GoMap{
		llb: llb,
	}
	return gm
}

//The bottom middle of the marker image will be placed on the point
func (gm *GoMap) DrawMarkers(markers []LonLat, layerName string, markerImage image.Image) error {
	if gm.baselayer == nil {
		return fmt.Errorf("please set baselayer before drawing features")
	}

	img := image.NewRGBA(gm.baselayer.Bounds())

	markW, markH := markerImage.Bounds().Size().X, markerImage.Bounds().Size().Y

	for _, m := range markers {
		point := m.toPixels(gm.llb, gm.baselayer.Bounds())
		draw.Draw(img, image.Rect(point.X - (markW/2), point.Y - (markH), point.X + (markW/2), point.Y), markerImage, image.Point{0, 0}, draw.Src)
	}

	gm.layers = append(gm.layers, Layer{img: img, name: layerName})
	return nil
}

func (gm *GoMap) SetTileBaseLayer(zoom int, getter TileGetter) error {

	xyzg := gm.llb.ToXYZoomGrid(zoom)

	mapTiles := make([][]ImageTile, 0)

	//For each XYZoom in list
	for _, row := range xyzg {
		mtRow := make([]ImageTile, 0)
		for _, xyz := range row {
			rt, err := getter.getTile(xyz)
			if err != nil {
				return fmt.Errorf("failed to get tile: %v", err)
			}
			it, err := rt.toImage()
			if err != nil {
				return fmt.Errorf("toImage failed: %v", err)
			}
			mtRow = append(mtRow, it)
		}
		mapTiles = append(mapTiles, mtRow)
	}

	if len(mapTiles) == 0 || len(mapTiles[0]) == 0 {
		return fmt.Errorf("no tiles populated")
	}

	//concatenate into a single image

	//get number of tiles along each side
	nx, ny := len(xyzg[0]), len(xyzg)

	//get size of each tile //TODO: Assuming each tile is the same size (should be?)
	tw, th := mapTiles[0][0].img.Bounds().Size().X, mapTiles[0][0].img.Bounds().Size().Y

	//create the new base layer image
	bl := image.NewRGBA(image.Rect(0, 0, tw * nx, th * ny)) //TODO: What image type to use

	//draw each tile onto the image
	for y, row := range mapTiles {
		for x, it := range row {
			draw.Draw(bl, image.Rect(x * tw, y * th, (x+1) * tw, (y+1) * th), it.img, image.Point{0, 0}, draw.Src)
		}
	}

	//Trim baselayer to exact bounds
	tilesLLB := xyzg.ToLonLatBound() //The tiles cover a larger area than the requested latlong bounds, this finds the "outer" llb so we can trim
	trimRect := gm.llb.ToRectangle(tilesLLB, bl.Bounds())

	trimmed := image.NewRGBA(image.Rect(0, 0, trimRect.Dx(), trimRect.Dy()))
	draw.Draw(trimmed, trimmed.Bounds(), bl, image.Point{trimRect.Min.X, trimRect.Min.Y}, draw.Src)

	gm.baselayer = trimmed

	return nil
}

func (gm GoMap) toPNG() ([]byte, error) {

	out := image.NewRGBA(gm.baselayer.Bounds())

	draw.Draw(out, gm.baselayer.Bounds(), gm.baselayer, image.Point{0, 0}, draw.Src)

	for _, l := range gm.layers {
		draw.Draw(out, gm.baselayer.Bounds(), l.img, image.Point{0, 0}, draw.Over)
	}

	var b bytes.Buffer

	err := png.Encode(&b, out)
	if err != nil {
		return b.Bytes(), fmt.Errorf("png encoding failed: %v", err)
	}

	return b.Bytes(), nil
}