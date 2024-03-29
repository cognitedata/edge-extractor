package camera

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type HikvisionCameraDriver struct {
	httpClient      http.Client
	digestTransport *dac.DigestTransport
}

func NewHikvisionCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &HikvisionCameraDriver{httpClient: httpClient}
}

func (cam *HikvisionCameraDriver) ExtractImage(address, username, password string) (*Image, error) {
	// http://10.22.15.61/ISAPI/Streaming/channels/1/picture

	address = address + "/ISAPI/Streaming/channels/1/picture"

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
	body, err := io.ReadAll(resp.Body)

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

func (cam *HikvisionCameraDriver) ExtractMetadata(address, username, password string) ([]byte, error) {
	return nil, nil
}

func (cam *HikvisionCameraDriver) Ping(address string) bool {
	return true
}

func (cam *HikvisionCameraDriver) Commit(transactionId string) error {
	return nil
}
