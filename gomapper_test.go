package gomapper

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestHttpTileGetter(t *testing.T) {

	work := []struct{
		urlT string
		x, y, zoom int
		shouldValidate bool
	} {
		{
			"https://static.geonet.org.nz/osm/v2/{z}/{x}/{y}.png",
			30, 19, 5,
			true,
		},
		{
			"https://static.geonet.org.nz/osm/v2/{x}/{y}.png",
			30, 19, 5,
			false,
		},
	}

	for _, w := range work {
		t.Run(w.urlT, func(t *testing.T) {

			htg, err := NewHttpTileGetter(w.urlT)
			if err != nil {
				t.Log(err)
				if !w.shouldValidate {
					return
				}
				t.Fatalf("NewHttpTileGetter returned error: %v", err)
			}

			rt, err := htg.getTile(XYzoom{w.x, w.y, w.zoom})
			if err != nil {
				t.Fatalf("getTile returned error: %v", err)
			}

			img, err := rt.toImage()
			if err != nil {
				t.Fatalf("toImage failed: %v", err)
			}

			t.Logf("received image type: %T", img.img)
		})
	}

}

func TestGetTileForLonLat(t *testing.T) {
	os.Mkdir("testout", 0755)

	work := []LonLat {
		{174.776230, -41.286461}, //Wellington
	}

	for i, w := range work {

		htg, err := NewHttpTileGetter("https://static.geonet.org.nz/osm/v2/{z}/{x}/{y}.png")
		if err != nil {
			t.Fatalf("NewHttpTileGetter returned error: %v", err)
		}

		rt, err := htg.getTile(w.toXYzoom(12))
		if err != nil {
			t.Fatalf("getTile returned error: %v", err)
		}

		err = ioutil.WriteFile(fmt.Sprintf("testout/TestGetTileForLonLat%d.%s", i, rt.format), rt.data, 0644)
		if err != nil {
			t.Fatalf("couldn't write file: %v", err)
		}
	}
}

func TestTileBaseLayerMap(t *testing.T) {
	os.Mkdir("testout", 0755)

	work := []struct {
		llb LonLatBounds
		zoom int
		markers []LonLat
	}{
		{
			LonLatBounds{LonLat{174.608878, -41.030715}, LonLat{175.181755, -41.357773}},
			12,
			[]LonLat{{174.758025, -41.308918}},
		},
		{
			LonLatBounds{LonLat{174.747011,-41.302995},LonLat{174.769205,-41.314784}},
			16,
			[]LonLat{{174.758025, -41.308918}},
		},
	}

	for i, w := range work {

		htg, err := NewHttpTileGetter("https://static.geonet.org.nz/osm/v2/{z}/{x}/{y}.png")
		if err != nil {
			t.Log(err)
			t.Fatalf("NewHttpTileGetter returned error: %v", err)
		}

		gm := NewMapFromBounds(w.llb)

		err = gm.SetTileBaseLayer(w.zoom, htg)
		if err != nil {
			t.Fatalf("SetTileBaseLayer returned error: %v", err)
		}

		tileImage, err := LoadMarkerImage("images/marker.png")
		if err != nil {
			t.Fatalf("LoadMarkerImage failed: %v", err)
		}

		err = gm.DrawMarkers(w.markers, "markers", tileImage)
		if err != nil {
			t.Fatalf("DrawMarkers failed: %v", err)
		}

		b, err := gm.toPNG()
		if err != nil {
			t.Fatalf("toPNG failed: %v", err)
		}

		err = ioutil.WriteFile(fmt.Sprintf("testout/TestTileBaseLayerMap%d.png", i), b, 0644)
		if err != nil {
			t.Fatalf("couldn't write file: %v", err)
		}
	}
}