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

func (cam *HikvisionCameraDriver) Ping(address string) bool {
	return true
}
