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
	MaxParallelWorkers  int      // Maximum number of parallel workers
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

// ConfigureFromRaw parses the raw configuration data and configures the CameraEventBasedCaptureApp.
// It unmarshals the configRaw into a CameraEventBasedCaptureAppConfig struct and sets the app's config.
// If the MaxParallelWorkers field is not specified in the configuration, it defaults to 1.
// Returns an error if there is an issue with unmarshaling the configuration.
func (app *CameraEventBasedCaptureApp) ConfigureFromRaw(configRaw json.RawMessage) error {
	var config CameraEventBasedCaptureAppConfig
	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return err
	}
	if config.MaxParallelWorkers == 0 {
		config.MaxParallelWorkers = 1
	}
	app.config = config
	return nil
}

// GetDependencies returns the dependencies required by the CameraEventBasedCaptureApp.
// It returns an instance of AppDependencies, which contains a list of integrations.
func (app *CameraEventBasedCaptureApp) GetDependencies() AppDependencies {
	return AppDependencies{
		Integrations: []string{"ip_cams_to_cdf"},
	}
}

func (app *CameraEventBasedCaptureApp) Configure(config CameraEventBasedCaptureAppConfig) {
	app.config = config
}

// ConfigureIntegration sets the integration for the CameraEventBasedCaptureApp.
// It takes an interface{} as a parameter and assigns it to the app's integration field.
// The integration parameter should be of type *ip_cams_to_cdf.CameraImagesToCdf.
func (app *CameraEventBasedCaptureApp) ConfigureIntegration(integration interface{}) {
	app.integration = integration.(*ip_cams_to_cdf.CameraImagesToCdf)
}

// Start starts the CameraEventBasedCaptureApp.
// It subscribes to the event stream from the specified topics and processes incoming events.
// Each new event will overwrite the previous one, and the processing timer will be reset.
// If the app is not already running, it will start the image capture loop in a separate goroutine.
// This function returns an error if there was a problem starting the app.
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
			go app.startImageCaptureLoop()
		}
		app.mux.Unlock()
	}
	app.log.Info("CameraEventBasedCaptureApp stopped")
	return nil
}

// startImageCaptureLoop is a method of the CameraEventBasedCaptureApp struct that starts the image capture loop.
// It captures images for the duration specified in CaptureDurationSec or until the next event occurs, or until the app is stopped.
// The method runs each image capture task in a separate worker, up to the maximum number of parallel workers specified in MaxParallelWorkers.
// It updates the total elapsed processing time and logs the progress.
// Once the capture duration is reached, the method finishes and logs the total elapsed time.
func (app *CameraEventBasedCaptureApp) startImageCaptureLoop() {
	app.isLoopRunning = true
	defer func() {
		app.mux.Lock()
		app.isLoopRunning = false
		app.mux.Unlock()
	}()
	app.log.Debug("Starting image capture loop")
	// capture images for the duration of CaptureDurationSec or until next event
	// or until the app is stopped
	for {
		metadata := map[string]string{
			"eventCorrelationId": strconv.FormatInt(app.lastEvent.Timestamp, 10),
		}
		for _, cameraID := range app.config.ListOfTargetCameras {
			if app.activeWorkers >= app.config.MaxParallelWorkers {
				app.log.Warnf("Max parallel workers reached. Waiting...")
				for app.activeWorkers >= app.config.MaxParallelWorkers {
					time.Sleep(500 * time.Millisecond)
				}
				app.log.Info("Resuming...")
			}
			// Running each image capture task in a separate worker
			go func(id uint64) {
				app.activeWorkers++
				for i := 0; i < 3; i++ {
					err := app.integration.ExecuteProcessorRunByCameraID(id, metadata)
					if err == nil {
						if i > 0 {
							app.log.Infof("Successfully captured and uploaded image from camera %d after %d retries", id, i)
						}
						break
					}
					app.log.Infof("Failed to capture and upload image from camera %d. Retrying...", id)
				}
				app.activeWorkers--
			}(cameraID)
		}
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
}

func (app *CameraEventBasedCaptureApp) Stop() error {
	app.integration.GetEventBus().Close(app.config.TriggerTopics...)
	return nil
}
