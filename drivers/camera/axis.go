package camera

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type AxisCameraDriver struct {
	httpClient http.Client
}

func NewAxisCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &AxisCameraDriver{httpClient: httpClient}
}

func (cam *AxisCameraDriver) ExtractImage(address, username, password string) (*Image, error) {

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)

	resp, err := cam.httpClient.Do(req)
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

	if contentType != "image/jpeg" {
		log.Errorf("Incompatable content type %s from camera API", contentType)
		return nil, fmt.Errorf("incompatible content type %s", contentType)
	}

	img := Image{Body: body, Format: "image/jpeg"}

	return &img, nil
}

func (cam *AxisCameraDriver) Ping(address string) bool {
	return true
}
