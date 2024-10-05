package lib

import (
	"encoding/json"
	"github.com/cognitedata/edge-extractor/internal/llm"
	"strconv"
	"sync"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/drivers/camera"
	"github.com/cognitedata/edge-extractor/integrations/ip_cams_to_cdf"
	log "github.com/sirupsen/logrus"
)

type GenaiImageProcessingAppConfig struct {
	TriggerTopics []string
	// List of camera IDs to capture images from
	ListOfTargetCameras      []uint64 // List of camera IDs to capture images from
	CaptureDurationSec       int64    // For how long to capture images after the event
	DelayBetweenCapture      float64  // Delay between image captures
	MaxParallelWorkers       int      // Maximum number of parallel workers
	GcpProjectID             string
	GcpRegionLocation        string //
	ModelName                string // For example : gemini-1.5-flash
	Prompt                   string
	ImageCaptureStopKeywords []string
}

// GenaiImageProcessingApp listens for camera events (motion detection) and call image capture on one or multiple cameras
type GenaiImageProcessingApp struct {
	integration                   *ip_cams_to_cdf.CameraImagesToCdf
	config                        GenaiImageProcessingAppConfig
	isLoopRunning                 bool
	lastEvent                     camera.CameraEvent
	totalElapsedProcessingTimeSec float64
	mux                           sync.Mutex
	log                           *log.Entry
	activeWorkers                 int
	isEventProcessingCanceled     bool
	llmProvider                   *llm.GcpLlmProvider
}

func NewGenaiImageProcessingApp() AppInstance {
	logger := log.WithField("app", "GenaiImageProcessingApp")
	return &GenaiImageProcessingApp{mux: sync.Mutex{}, log: logger}
}

// ConfigureFromRaw parses the raw configuration data and configures the GenaiImageProcessingApp.
// It unmarshals the configRaw into a GenaiImageProcessingAppConfig struct and sets the app's config.
// If the MaxParallelWorkers field is not specified in the configuration, it defaults to 1.
// Returns an error if there is an issue with unmarshaling the configuration.
func (app *GenaiImageProcessingApp) ConfigureFromRaw(configRaw json.RawMessage) error {
	var config GenaiImageProcessingAppConfig
	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return err
	}
	if config.MaxParallelWorkers == 0 {
		config.MaxParallelWorkers = 1
	}
	app.config = config
	log.Infof("GenaiImageProcessingApp configured with: %+v", app.config)
	return nil
}

// GetDependencies returns the dependencies required by the GenaiImageProcessingApp.
// It returns an instance of AppDependencies, which contains a list of integrations.
func (app *GenaiImageProcessingApp) GetDependencies() AppDependencies {
	return AppDependencies{
		Integrations: []string{"ip_cams_to_cdf"},
	}
}

func (app *GenaiImageProcessingApp) Configure(config GenaiImageProcessingAppConfig) {
	app.config = config
}

// ConfigureIntegration sets the integration for the GenaiImageProcessingApp.
// It takes an interface{} as a parameter and assigns it to the app's integration field.
// The integration parameter should be of type *ip_cams_to_cdf.CameraImagesToCdf.
func (app *GenaiImageProcessingApp) ConfigureIntegration(integration interface{}) {
	app.integration = integration.(*ip_cams_to_cdf.CameraImagesToCdf)
}

// Start starts the GenaiImageProcessingApp.
// It subscribes to the event stream from the specified topics and processes incoming events.
// Each new event will overwrite the previous one, and the processing timer will be reset.
// If the app is not already running, it will start the image capture loop in a separate goroutine.
// This function returns an error if there was a problem starting the app.
func (app *GenaiImageProcessingApp) Start() error {

	app.llmProvider = llm.NewGcpLlmProvider(app.config.GcpProjectID, app.config.GcpRegionLocation, app.config.ModelName)
	err := app.llmProvider.Init()
	if err != nil {
		app.log.Errorf("Failed to initialize LLM provider: %v", err)
		return err
	}
	go app.startEventProcessingLoop()
	return nil
}

func (app *GenaiImageProcessingApp) startEventProcessingLoop() error {
	app.log.Info("Starting GenaiImageProcessingApp")
	eventStream := app.integration.GetEventBus().Sub(app.config.TriggerTopics...)
	app.log.Info("GenaiImageProcessingApp subscribed to event stream from topics: ", app.config.TriggerTopics)
	for event := range eventStream {
		// Important: next event will overwrite the previous one. Next event can arrive before the previous one is processed
		app.log.Debugf("New event from topic: %s", event.Topic)
		app.mux.Lock()
		app.lastEvent = event
		app.totalElapsedProcessingTimeSec = 0 // reset the total elapsed processing timer
		if !app.isLoopRunning {
			go app.startEventImageCaptureLoop()
		}
		app.mux.Unlock()
	}
	app.log.Info("GenaiImageProcessingApp stopped")
	return nil
}

// startEventImageCaptureLoop is a method of the GenaiImageProcessingApp struct that starts the image capture loop.
// It captures images for the duration specified in CaptureDurationSec or until the next event occurs, or until the app is stopped.
// The method runs each image capture task in a separate worker, up to the maximum number of parallel workers specified in MaxParallelWorkers.
// It updates the total elapsed processing time and logs the progress.
// Once the capture duration is reached, the method finishes and logs the total elapsed time.
func (app *GenaiImageProcessingApp) startEventImageCaptureLoop() {
	app.isLoopRunning = true
	defer func() {
		app.mux.Lock()
		app.isLoopRunning = false
		app.mux.Unlock()
	}()
	app.log.Info("Starting image capture loop...")
	// capture images for the duration of CaptureDurationSec or until next event
	// or until the app is stopped
	for {
		// imageSyncID is used to correlate images captured from different cameras
		imageSyncID := time.Now().UnixNano()
		for _, cameraID := range app.config.ListOfTargetCameras {
			if app.activeWorkers >= app.config.MaxParallelWorkers {
				app.log.Warnf("Max parallel workers reached. Waiting...")
				app.integration.BaseIntegration.ReportRunStatus("", core.ExtractionRunStatusFailure, "Max parallel workers reached for app GenaiImageProcessingApp. Waiting...")
				for app.activeWorkers >= app.config.MaxParallelWorkers {
					time.Sleep(500 * time.Millisecond)
				}
				app.log.Info("Resuming...")
			}
			// Running each image capture task in a separate worker
			go app.runWorker(cameraID, imageSyncID)
		}
		app.mux.Lock()
		app.totalElapsedProcessingTimeSec += app.config.DelayBetweenCapture
		app.mux.Unlock()
		if app.totalElapsedProcessingTimeSec >= float64(app.config.CaptureDurationSec) {
			break
		}
		delay := time.Duration(app.config.DelayBetweenCapture * 1000)
		app.log.Debugf("Total elapsed time: %d sec , delay %d milisec. Active workers %d", int(app.totalElapsedProcessingTimeSec), delay, app.activeWorkers)
		time.Sleep(delay * time.Millisecond)
	}
	app.log.Infof("Event processed in %d sec", int(app.totalElapsedProcessingTimeSec))
}

func (app *GenaiImageProcessingApp) runWorker(id uint64, imageSyncID int64) {
	app.activeWorkers++
	defer func() {
		app.activeWorkers--
	}()
	metadata := map[string]string{
		"eventCorrelationId": strconv.FormatInt(app.lastEvent.Timestamp, 10),
		"cameraId":           strconv.FormatUint(id, 10),
		"imageSyncId":        strconv.FormatInt(imageSyncID, 10),
	}

	camera := app.integration.GetCameraByID(id)
	if camera == nil {
		app.log.Errorf("Camera with id %d not found", id)
		return
	}
	cameraConfig := app.integration.GetCameraConfigByID(id)
	if cameraConfig == nil {
		app.log.Errorf("Camera config with id %d not found", id)
		return
	}

	image, err := camera.ExtractImage()
	if err != nil {
		app.log.Errorf("Failed to capture and upload image from camera %d: %v", id, err)
		return
	}

	llmResponse, err := app.llmProvider.ImageToText(image.Body, app.config.Prompt)

	if err != nil {
		app.log.Errorf("Failed to convert image to text: %v", err)
		metadata["llmAppSpecificResponse"] = ""
	} else {
		metadata["llmAppSpecificResponse"] = string(llmResponse.AppSpecificResponse)
		systemResponseText, err := json.Marshal(llmResponse.System)
		if err != nil {
			app.log.Errorf("Failed to marshal system response: %v", err)
		} else {
			metadata["llmSystemResponse"] = string(systemResponseText)
		}
	}

	log.Info("-------------Pipe detection results-------------")
	log.Infof("Event ID : %s", metadata["eventCorrelationId"])
	log.Infof("Camera ID : %s", metadata["cameraId"])
	log.Infof("Image Sync ID : %s", metadata["imageSyncId"])
	log.Info(metadata["llmAppSpecificResponse"])
	log.Info(metadata["llmSystemResponse"])

	//err = app.integration.PublishImageToCDF(*cameraConfig, image, metadata)
	//if err != nil {
	//	app.log.Errorf("Failed to publish image to CDF: %v", err)
	//	return
	//}
	log.Debug("Image processed and published to CDF successfully : %s")

}

func (app *GenaiImageProcessingApp) Stop() error {
	app.integration.GetEventBus().Close(app.config.TriggerTopics...)
	app.llmProvider.Close()
	return nil
}
