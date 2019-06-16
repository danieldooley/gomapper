package gomapper

import (
	"image"
	"math"
)

type LonLat struct {
	Lon, Lat float64
}

//From https://github.com/apeyroux/gosm/blob/master/gosm.go
//I'm pretty sure this assumes tiles to be in EPSG:3857 //TODO: Do we need other projections
func (ll LonLat) toXYzoom(zoom int) XYzoom {
	x := int(math.Floor((ll.Lon + 180.0) / 360.0 * (math.Exp2(float64(zoom)))))
	y := int(math.Floor((1.0 - math.Log(math.Tan(ll.Lat*math.Pi/180.0)+1.0/math.Cos(ll.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(zoom)))))
	return XYzoom{x, y, zoom}
}

//Worked this one out myself :D
func (ll LonLat) toPixels(llb LonLatBounds, r image.Rectangle) image.Point {
	lonSpan, latSpan := llb.BottomRight.Lon - llb.TopLeft.Lon, llb.BottomRight.Lat - llb.TopLeft.Lat
	lonScale, latScale := (ll.Lon - llb.TopLeft.Lon) / lonSpan, (ll.Lat - llb.TopLeft.Lat) / latSpan
	return image.Point{int(math.Round(float64(r.Size().X) * lonScale)), int(math.Round(float64(r.Size().Y) * latScale))}
}

type XYzoom struct {
	X, Y, Zoom int
}

//From https://github.com/apeyroux/gosm/blob/master/gosm.go
//I'm pretty sure this assumes tiles to be in EPSG:3857 //TODO: Do we need other projections
func (xyz XYzoom) toLonLat() LonLat {
	n := math.Pi - 2.0*math.Pi*float64(xyz.Y)/math.Exp2(float64(xyz.Zoom))
	lat := 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	lon := float64(xyz.X)/math.Exp2(float64(xyz.Zoom))*360.0 - 180.0
	return LonLat{lon, lat}
}

type LonLatBounds struct {
	TopLeft, BottomRight LonLat
}


//Adapted from https://github.com/apeyroux/gosm/blob/master/gosm.go
func (llb LonLatBounds) ToXYZoomGrid(zoom int) XYZoomGrid {
	out := make(XYZoomGrid, 0)

	topLeftXYZ := llb.TopLeft.toXYzoom(zoom)
	bottomRightXYZ := llb.BottomRight.toXYzoom(zoom)

	for y := topLeftXYZ.Y; y <= bottomRightXYZ.Y; y++ {
		row := make([]XYzoom, 0)
		for x := topLeftXYZ.X; x <= bottomRightXYZ.X; x++ {
			row = append(row, XYzoom{x, y, zoom})
		}
		out = append(out, row)
	}

	return out
}

func (llb LonLatBounds) ToRectangle(outerllb LonLatBounds, r image.Rectangle) image.Rectangle {
	min := llb.TopLeft.toPixels(outerllb, r)
	max := llb.BottomRight.toPixels(outerllb, r)
	return image.Rect(min.X, min.Y, max.X, max.Y)
}

//2d Array of XYzoom locations (indexed like: xyzg[y][x]
type XYZoomGrid [][]XYzoom

func (xyzg XYZoomGrid) ToLonLatBound() LonLatBounds {
	if len(xyzg) == 0 || len(xyzg[0]) == 0 {
		return LonLatBounds{}
	}

	TopLeftTile := xyzg[0][0]
	BottomRightTile := xyzg[len(xyzg)-1][len(xyzg[len(xyzg)-1])-1]

	return LonLatBounds{
		TopLeft: TopLeftTile.toLonLat(),
		//The LonLat calculated is the TopLeft (minimum) point of the tile, we want the bottom right so we just get the diagonally down/right tile
		BottomRight: XYzoom{X: BottomRightTile.X+1, Y:BottomRightTile.Y+1, Zoom:BottomRightTile.Zoom}.toLonLat(),
	}
}

