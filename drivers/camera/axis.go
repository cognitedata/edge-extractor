package camera

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type AxisCameraDriver struct {
	httpClient      http.Client
	digestTransport *dac.DigestTransport
}

func NewAxisCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &AxisCameraDriver{httpClient: httpClient}
}

func (cam *AxisCameraDriver) ExtractImage(address, username, password string) (*Image, error) {
	//"http://10.22.15.62/axis-cgi/jpg/image.cgi"
	address = address + "/axis-cgi/jpg/image.cgi"
	if cam.digestTransport == nil {
		t := dac.NewTransport(username, password)
		cam.digestTransport = &t
		cam.digestTransport.HTTPClient = &cam.httpClient
	}

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}
	// req.SetBasicAuth(username, password)

	// dump, _ := httputil.DumpRequestOut(req, false)
	// log.Debug("HTTP request debug :", string(dump))

	// resp, err := cam.httpClient.Do(req)
	// if err != nil {
	// 	return nil, err
	// }

	resp, err := cam.digestTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("camera api returned error code %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")

	if !strings.Contains(contentType, "image/jpeg") {
		log.Errorf("Incompatable content type %s from camera API", contentType)
		return nil, fmt.Errorf("incompatible content type %s", contentType)
	}

	img := Image{Body: body, Format: "image/jpeg"}

	return &img, nil
}

func (cam *AxisCameraDriver) ExtractMetadata(address, username, password string) ([]byte, error) {
	return nil, nil
}

func (cam *AxisCameraDriver) Ping(address string) bool {
	return true
}
