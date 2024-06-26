package camera

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type FlirAx8CameraDriver struct {
	httpClient http.Client
	address    string
	username   string
	password   string
}

// docs : https://flir.custhelp.com/app/answers/detail/a_id/3602/~/getting-started-using-rest-api-with-automation-cameras

// http://10.28.0.60/snapshot.jpg

func NewFlirAx8CameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &FlirAx8CameraDriver{httpClient: httpClient}
}

func (cam *FlirAx8CameraDriver) Configure(address, username, password string) error {
	cam.address = address
	cam.username = username
	cam.password = password
	return nil
}

func (cam *FlirAx8CameraDriver) ExtractImage() (*Image, error) {
	address := cam.address + "/snapshot.jpg"

	resp, err := cam.httpClient.Get(address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("camera api returned error code %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")

	if !strings.Contains(contentType, "image/jpeg") {
		log.Errorf("Incompatable content type %s from FlirAx8 camera API", contentType)
		return nil, fmt.Errorf("incompatible content type %s", contentType)
	}

	img := Image{Body: body, Format: "image/jpeg"}

	return &img, nil
}

func (cam *FlirAx8CameraDriver) ExtractMetadata() ([]byte, error) {
	address := cam.address + "/res.php"

	data := url.Values{
		"action": {"measurement"},
		"type":   {"spot"},
		"id":     {"1"},
	}

	resp, err := http.PostForm(address, data)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("camera metadata api returned error code %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (cam *FlirAx8CameraDriver) Ping(address string) bool {
	return true
}

func (cam *FlirAx8CameraDriver) Commit(transactionId string) error {
	return nil
}

func (cam *FlirAx8CameraDriver) SubscribeToEventsStream(eventFilters []EventFilter) (chan CameraEvent, error) {
	return nil, fmt.Errorf("FlirAx8 camera does not support event streaming")
}

func (cam *FlirAx8CameraDriver) Close() {
}

func (cam *FlirAx8CameraDriver) GetCameraCapabilitiesManifest(component string) ([]CameraCapabilitiesManifest, error) {
	return nil, nil
}
