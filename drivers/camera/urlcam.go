package camera

import (
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type UrlCameraDriver struct {
	httpClient http.Client
	address    string
	username   string
	password   string
}

func NewUrlCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &UrlCameraDriver{httpClient: httpClient}
}

func (cam *UrlCameraDriver) Configure(address, username, password string) error {
	cam.address = address
	cam.username = username
	cam.password = password
	return nil
}

func (cam *UrlCameraDriver) ExtractImage() (*Image, error) {

	// resp, err := cam.httpClient.Get(address)

	req, err := http.NewRequest("GET", cam.address, nil)
	if err != nil {
		return nil, err
	}
	if cam.username != "" && cam.password != "" {
		req.SetBasicAuth(cam.username, cam.password)
	}
	resp, err := cam.httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("camera api returned error code %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)

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

func (cam *UrlCameraDriver) ExtractMetadata() ([]byte, error) {
	return nil, nil
}

func (cam *UrlCameraDriver) Ping(address string) bool {
	return true
}

func (cam *UrlCameraDriver) Commit(transactionId string) error {
	return nil
}

func (cam *UrlCameraDriver) SubscribeToEventsStream(eventFilters []EventFilter) (chan CameraEvent, error) {
	return nil, nil
}

func (cam *UrlCameraDriver) Close() {
}

func (cam *UrlCameraDriver) GetCameraCapabilitiesManifest(component string) ([]CameraCapabilitiesManifest, error) {
	return nil, nil
}
