package camera

import (
	"encoding/json"
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
	address         string
	username        string
	password        string
	wsConnection    *websocket.Conn
}

// TopicFilter is the filter for the event topic.
// ContentFilter is the filter for the event content.
/*
Example of event filter for Axis camera:
 "EventFilters": [
            {
              "TopicFilter": "tnsaxis:CameraApplicationPlatform/FenceGuard/Camera1Profile1",
              "ContentFilter": "boolean(//SimpleItem[@Name=\"active\" and @Value=\"1\"])"
            }
*/
type AxisEventFilter struct {
	TopicFilter   string `json:"topicFilter,omitempty"`
	ContentFilter string `json:"contentFilter,omitempty"`
}

type AxisNotificationMessage struct {
	Source map[string]string `json:"source,omitempty"`
	Key    map[string]string `json:"key,omitempty"`
	Data   map[string]string `json:"data,omitempty"`
}
type AxisNotification struct {
	Topic     string                  `json:"topic,omitempty"`
	Timestamp int64                   `json:"timestamp,omitempty"` // Unix timestamp in milliseconds
	Message   AxisNotificationMessage `json:"message,omitempty"`
}

type AxisEventParams struct {
	EventFilterList []AxisEventFilter `json:"eventFilterList,omitempty"`
	Notification    AxisNotification  `json:"notification,omitempty"`
}

type AxisEvent struct {
	APIVersion string          `json:"apiVersion"`
	Context    string          `json:"context"`
	Method     string          `json:"method"`
	Params     AxisEventParams `json:"params"`
}

func NewAxisCameraDriver() Driver {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	return &AxisCameraDriver{httpClient: httpClient}
}

func (cam *AxisCameraDriver) Configure(address, username, password string) error {
	cam.address = address
	cam.username = username
	cam.password = password
	return nil
}

func (cam *AxisCameraDriver) ExtractImage() (*Image, error) {
	address := cam.address + "/axis-cgi/jpg/image.cgi"
	if cam.digestTransport == nil {
		t := dac.NewTransport(cam.username, cam.password)
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

func (cam *AxisCameraDriver) ExtractMetadata() ([]byte, error) {
	return nil, nil
}

func (cam *AxisCameraDriver) Ping(address string) bool {
	return true
}

func (cam *AxisCameraDriver) Commit(transactionId string) error {
	return nil
}

func (cam *AxisCameraDriver) GetServicesDiscoveryManifest(component string) ([]CameraServiceDiscoveryManifest, error) {
	address := cam.address + "/vapix/services"
	if cam.digestTransport == nil {
		t := dac.NewTransport(cam.username, cam.password)
		cam.digestTransport = &t
		cam.digestTransport.HTTPClient = &cam.httpClient
	}

	requestBody := `<?xml version="1.0" encoding="UTF-8"?>
	<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
	 <Header/>
	 <Body >
	  <GetEventInstances xmlns="http://www.axis.com/vapix/ws/event1"/>
	 </Body>
	</Envelope>`

	requestBodyIoReader := strings.NewReader(requestBody)

	req, err := http.NewRequest("POST", address, requestBodyIoReader)
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
	eventsManifest := CameraServiceDiscoveryManifest{
		Name:          "events.xml",
		Format:        "soap",
		ComponentName: "events",
		Body:          body,
	}
	manifests := []CameraServiceDiscoveryManifest{eventsManifest}
	return manifests, nil
}

// Connect to Axis WebSocket API and subscribe to events from the camera , for example motion detection
func (cam *AxisCameraDriver) SubscribeToEventsStream(eventFilters []EventFilter) (stream chan CameraEvent, err error) {
	// convert the address to a websocket address
	digestAddress := cam.address + "/vapix/ws-data-stream?sources=events"
	digestRequest := edgedac.NewRequest(cam.username, cam.password, "GET", digestAddress, "")
	address := strings.Replace(cam.address, "http", "ws", 1) + "/vapix/ws-data-stream?sources=events"

	log.Info("Connecting to Axis camera event stream over WS. Address : ", address)
	var authHeader string
	var resp *http.Response
	cam.wsConnection, resp, err = websocket.DefaultDialer.Dial(address, nil)
	if resp != nil && resp.StatusCode == 401 {
		authHeader, err = digestRequest.GetNewDigestAuthHeaderFromResponse(resp)
		if err != nil {
			log.Error("Error getting new digest auth header:", err)
			return nil, err
		}
		header := http.Header{"Authorization": []string{authHeader}}
		log.Debug("Using auth header ", header)
		cam.wsConnection, resp, err = websocket.DefaultDialer.Dial(address, header)
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

	axisEventFilterList := make([]AxisEventFilter, len(eventFilters))
	for i, filter := range eventFilters {
		axisEventFilterList[i] = AxisEventFilter(filter)
	}

	messages := make(chan CameraEvent, 10)
	go func() {
		cam.wsMessageHandler(axisEventFilterList, messages)
	}()
	return messages, nil
}

func (cam *AxisCameraDriver) wsMessageHandler(axisEventFilterList []AxisEventFilter, messages chan CameraEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Info("Recovered from panic:", r)
			log.Info("Disconnected from camera websocket.")
		}
		defer cam.wsConnection.Close()
	}()
	log.Info("Connected to camera websocket and subscribed to events.")
	eventFilter := AxisEvent{
		APIVersion: "1.0",
		Context:    "edge-extractor event subscription",
		Method:     "events:configure",
		Params: AxisEventParams{
			EventFilterList: axisEventFilterList,
		},
	}
	log.Debugf("eventFilter: %+v\n", eventFilter)
	cam.wsConnection.WriteJSON(eventFilter)
	for {
		_, message, err := cam.wsConnection.ReadMessage()
		if err != nil {
			log.Error("Error from WS stream:", err)
			close(messages)
			break
		}

		exisEvent := AxisEvent{}
		err = json.Unmarshal(message, &exisEvent)
		if err != nil {
			log.Info("Error parsing JSON from Axis WS stream:", err)
			continue
		}
		if exisEvent.Method == "events:configure" {
			continue
		}
		cameraEvent := CameraEvent{
			CoreType:  "notification",
			Type:      "CamMotionDetected",
			Source:    "cam:axis:name",
			Topic:     exisEvent.Params.Notification.Topic,
			Timestamp: exisEvent.Params.Notification.Timestamp,
			RawData:   message}
		select {
		case messages <- cameraEvent:

		default:

			log.Info("Channel is full, message not sent")
		}
	}
	log.Info("Disconnected from camera websocket.")
}

func (cam *AxisCameraDriver) Close() {
	log.Info("Stopping Axis camera driver")
	if cam.wsConnection != nil {
		cam.wsConnection.Close()
	}
}
