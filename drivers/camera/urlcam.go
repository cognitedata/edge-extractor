package camera

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type UrlCameraDriver struct {
	httpClient http.Client
}

func NewUrlCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &UrlCameraDriver{httpClient: httpClient}
}

func (cam *UrlCameraDriver) ExtractImage(address, username, password string) (*Image, error) {

	resp, err := cam.httpClient.Get(address)
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

func (cam *UrlCameraDriver) ExtractMetadata(address, username, password string) ([]byte, error) {
	return nil, nil
}

func (cam *UrlCameraDriver) Ping(address string) bool {
	return true
}
