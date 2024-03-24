package lib

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/cognitedata/edge-extractor/drivers/camera"
	"github.com/cognitedata/edge-extractor/integrations/ip_cams_to_cdf"
	log "github.com/sirupsen/logrus"
)

// App listens for camera events (motion detection) and call image capture on one or multiple cameras
type CameraEventBasedCaptureAppConfig struct {
	TriggerTopics []string
	// List of camera IDs to capture images from
	ListOfTargetCameras []uint64 // List of camera IDs to capture images from
	CaptureDurationSec  int64    // For how long to capture images after the event
	DelayBetweenCapture float64  // Delay between image captures
}

type CameraEventBasedCaptureApp struct {
	integration                   *ip_cams_to_cdf.CameraImagesToCdf
	config                        CameraEventBasedCaptureAppConfig
	isLoopRunning                 bool
	lastEvent                     camera.CameraEvent
	totalElapsedProcessingTimeSec float64
	mux                           sync.Mutex
	log                           *log.Entry
	activeWorkers                 int
}

func NewCameraEventBasedCaptureApp() AppInstance {
	logger := log.WithField("app", "CameraEventBasedCaptureApp")
	return &CameraEventBasedCaptureApp{mux: sync.Mutex{}, log: logger}
}

func (app *CameraEventBasedCaptureApp) ConfigureFromRaw(configRaw json.RawMessage) error {
	var config CameraEventBasedCaptureAppConfig
	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return err
	}
	app.config = config
	return nil
}

func (app *CameraEventBasedCaptureApp) GetDependencies() AppDependencies {
	return AppDependencies{
		Integrations: []string{"ip_cams_to_cdf"},
	}
}

func (app *CameraEventBasedCaptureApp) Configure(integration *ip_cams_to_cdf.CameraImagesToCdf, config CameraEventBasedCaptureAppConfig) {
	app.integration = integration
	app.config = config
}

func (app *CameraEventBasedCaptureApp) ConfigureIntegration(integration interface{}) {
	app.integration = integration.(*ip_cams_to_cdf.CameraImagesToCdf)
}

func (app *CameraEventBasedCaptureApp) Start() error {
	app.log.Info("Starting CameraEventBasedCaptureApp")
	eventStream := app.integration.GetEventBus().Sub(app.config.TriggerTopics...)
	app.log.Info("CameraEventBasedCaptureApp subscribed to event stream from topics: ", app.config.TriggerTopics)
	for event := range eventStream {
		// Important: next event will overwrite the previous one. Next event can arrive before the previous one is processed
		app.log.Debugf("New event from topic: %s", event.Topic)
		app.mux.Lock()
		app.lastEvent = event
		app.totalElapsedProcessingTimeSec = 0 // reset the total elapsed processing timer
		if !app.isLoopRunning {
			app.log.Debugf("Starting image capture loop")
			go app.startImageCaptureLoop()
		}
		app.mux.Unlock()
		app.log.Debugf("Event processed from topic: %s", event.Topic)
	}
	app.log.Info("CameraEventBasedCaptureApp stopped")
	return nil
}
func (app *CameraEventBasedCaptureApp) startImageCaptureLoop() {
	app.isLoopRunning = true
	defer func() {
		app.mux.Lock()
		app.isLoopRunning = false
		app.mux.Unlock()
	}()
	app.log.Info("Starting image capture loop")
	// capture images for the duration of CaptureDurationSec or until next event
	// or until the app is stopped
	for {
		// startTime := time.Now()
		metadata := map[string]string{
			"eventCorrelationId": strconv.FormatInt(app.lastEvent.Timestamp, 10),
		}
		for _, cameraID := range app.config.ListOfTargetCameras {
			go func(id uint64) {
				app.activeWorkers++
				app.integration.ExecuteProcessorRunByCameraID(id, metadata)
				app.activeWorkers--
			}(cameraID)
		}
		// elapsedTime := time.Since(startTime)
		app.mux.Lock()
		app.totalElapsedProcessingTimeSec += app.config.DelayBetweenCapture
		app.mux.Unlock()
		if app.totalElapsedProcessingTimeSec >= float64(app.config.CaptureDurationSec) {
			break
		}
		app.log.Debugf("Total elapsed time: %d sec. Active workers %d", int(app.totalElapsedProcessingTimeSec), app.activeWorkers)
		// sleep for DelayBetweenCapture
		time.Sleep(time.Duration(app.config.DelayBetweenCapture) * time.Second)
	}
	app.log.Infof("Image capture loop finished.Total elapsed time: %d sec", int(app.totalElapsedProcessingTimeSec))
}

func (app *CameraEventBasedCaptureApp) Stop() error {
	app.integration.GetEventBus().Close(app.config.TriggerTopics...)
	return nil
}
