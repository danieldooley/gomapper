package gomapper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = time.Second * 5
)

/*
	Interface
 */

//Any TileGetter should probably be safe for parallel exec
type TileGetter interface {
	getTile(xyz XYzoom) (RawTile, error)
}


/*
	HTTP
 */

type HTTPTileGetter struct {
	urlTemplate string //Format of 'https://static1.geonet.org.nz/osm/v2/{z}/{x}/{y}.png' //TODO: Support multiple subdomains?
	httpClient http.Client
}

func NewHttpTileGetter(urlTemplate string) (HTTPTileGetter, error) {
	var out HTTPTileGetter

	//validate the URL as usable
	required := []string{"{z}", "{x}", "{y}"}
	for _, r := range required {
		if !strings.Contains(urlTemplate, r) {
			return out, fmt.Errorf("URL Template does not contain required token %v", r)
		}
	}

	out.httpClient = http.Client{Timeout:defaultHTTPTimeout}

	out.urlTemplate = urlTemplate

	return out, nil
}

func (tg HTTPTileGetter) getTile(xyz XYzoom) (RawTile, error) {
	out := RawTile{
		xyz: xyz,
	}

	xUrl := strings.Replace(tg.urlTemplate, "{x}", fmt.Sprintf("%d", xyz.X), -1)
	xyURL := strings.Replace(xUrl, "{y}", fmt.Sprintf("%d", xyz.Y), -1)
	xyzURL := strings.Replace(xyURL, "{z}", fmt.Sprintf("%d", xyz.Zoom), -1)

	res, err := tg.httpClient.Get(xyzURL)
	if err != nil {
		return out, fmt.Errorf("HTTP GET failed; %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return out, fmt.Errorf("non-200 status code from server: %v", res.Status)
	}

	out.data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return out, fmt.Errorf("failed to read HTTP response body: %v", err)
	}

	//try determine type using mime typing
	switch res.Header.Get("content-type") {
	case "image/png":
		out.format = "png"
	default:
		return out, fmt.Errorf("unknown mime type: %v", res.Header.Get("content-type"))
	}

	return out, nil
}