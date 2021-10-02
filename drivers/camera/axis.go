package camera

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
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

	img := Image{Body: body, Format: "image/jpeg"}

	return &img, nil
}

func (cam *AxisCameraDriver) Ping(address string) bool {
	return true
}
