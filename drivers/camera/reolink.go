package camera

import (
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type ReolinkCameraDriver struct {
	httpClient http.Client
	address    string
	username   string
	password   string
}

func NewReolinkCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &ReolinkCameraDriver{httpClient: httpClient}
}

func (cam *ReolinkCameraDriver) Configure(address, username, password string) error {
	cam.address = address
	cam.username = username
	cam.password = password
	return nil
}

func (cam *ReolinkCameraDriver) ExtractImage() (*Image, error) {

	address := fmt.Sprintf("%s&user=%s&password=%s", cam.address, cam.username, cam.password)
	resp, err := cam.httpClient.Get(address)
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

func (cam *ReolinkCameraDriver) ExtractMetadata() ([]byte, error) {
	return nil, nil
}

func (cam *ReolinkCameraDriver) Ping(address string) bool {
	return true
}

func (cam *ReolinkCameraDriver) Commit(transactionId string) error {
	return nil
}

func (cam *ReolinkCameraDriver) SubscribeToEventsStream(eventFilters []EventFilter) (chan CameraEvent, error) {
	return nil, fmt.Errorf("reolink camera does not support event streams")
}
