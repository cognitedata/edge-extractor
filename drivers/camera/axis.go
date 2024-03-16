package camera

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	edgedac "github.com/cognitedata/edge-extractor/internal/auth/digest"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type AxisCameraDriver struct {
	httpClient      http.Client
	digestTransport *dac.DigestTransport
}

type EventFilter struct {
	TopicFilter   string `json:"topicFilter,omitempty"`
	ContentFilter string `json:"contentFilter,omitempty"`
}

type EventParams struct {
	EventFilterList []EventFilter `json:"eventFilterList"`
}

type Event struct {
	APIVersion string      `json:"apiVersion"`
	Context    string      `json:"context"`
	Method     string      `json:"method"`
	Params     EventParams `json:"params"`
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

func (cam *AxisCameraDriver) ExtractMetadata(address, username, password string) ([]byte, error) {
	return nil, nil
}

func (cam *AxisCameraDriver) Ping(address string) bool {
	return true
}

func (cam *AxisCameraDriver) Commit(transactionId string) error {
	return nil
}

// Connect to Axis WebSocket API and subscribe to events from the camera , for example motion detection

func (cam *AxisCameraDriver) SubscribeToEventsStream(address, username, password string) (stream chan CameraEvent, err error) {
	// convert the address to a websocket address
	digestAddress := address + "/vapix/ws-data-stream?sources=events"
	digestRequest := edgedac.NewRequest(username, password, "GET", digestAddress, "")

	address = strings.Replace(address, "http", "ws", 1) + "/vapix/ws-data-stream?sources=events"

	log.Debug("Connecting to camera websocket at ", address)
	var authHeader string
	c, resp, err := websocket.DefaultDialer.Dial(address, nil)
	if resp.StatusCode == 401 {
		authHeader, err = digestRequest.GetNewDigestAuthHeaderFromResponse(resp)
		if err != nil {
			log.Error("Error getting new digest auth header:", err)
			return nil, err
		}
		header := http.Header{"Authorization": []string{authHeader}}
		log.Debug("Using auth header ", header)
		c, resp, err = websocket.DefaultDialer.Dial(address, header)
	}

	if err != nil {
		log.Error("Error connecting to camera websocket:", err)
		if resp != nil {
			log.Info("Response from camera websocket:", resp.Status)
			log.Info("Response headers from camera websocket:", resp.Header)
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Info("Response body from camera websocket:", string(bodyBytes))
		}
		return nil, err
	}

	messages := make(chan CameraEvent, 10)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Info("Recovered from panic:", r)
				log.Info("Disconnected from camera websocket.")
			}
			defer c.Close()
		}()
		log.Info("Connected to camera websocket and subscribed to events.")
		eventFilter := Event{
			APIVersion: "1.0",
			Context:    "edge-extractor event subscription",
			Method:     "events:configure",
			Params: EventParams{
				EventFilterList: []EventFilter{
					{
						TopicFilter: "tnsaxis:CameraApplicationPlatform/FenceGuardCamera1ProfileANY", //tnsaxis:CameraApplicationPlatform/FenceGuardCamera1ProfileANY
					},
					{
						TopicFilter: "tns1:Device/tnsaxis:IO/Port",
					},
				},
			},
		}
		log.Debugf("eventFilter: %+v\n", eventFilter)
		c.WriteJSON(eventFilter)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Info("Error from WS stream:", err)
				close(messages)
				break
			}
			select {
			case messages <- CameraEvent{Data: message, Type: fmt.Sprint(mt)}:
				// Message sent successfully
			default:
				// Channel is full, message not sent
				log.Info("Channel is full, message not sent")
			}
		}
		log.Info("Disconnected from camera websocket.")
	}()
	return messages, nil
}
