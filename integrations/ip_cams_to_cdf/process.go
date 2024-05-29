package ip_cams_to_cdf

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/connectors/inputs"
	"github.com/cognitedata/edge-extractor/drivers/camera"
	"github.com/cognitedata/edge-extractor/integrations"
	"github.com/cognitedata/edge-extractor/internal"
	"github.com/cskr/pubsub/v2"
	log "github.com/sirupsen/logrus"
)

type CameraImagesToCdf struct {
	integrations.BaseIntegration
	successCounter    uint64
	failureCounter    uint64
	cameraConfigs     []CameraConfig
	cameras           map[uint64]*inputs.IpCamera
	secretManager     *internal.SecretManager
	integrationConfig IntegrationConfig
	eventbus          *pubsub.PubSub[string, camera.CameraEvent]
}

func NewCameraImagesToCdf(cogClient *internal.CdfClient, extractorMonitoringID string, configObserver *internal.CdfConfigObserver, systemEventBus *pubsub.PubSub[string, internal.SystemEvent]) *CameraImagesToCdf {
	eventBus := pubsub.New[string, camera.CameraEvent](20)
	ingr := &CameraImagesToCdf{
		BaseIntegration: *integrations.NewIntegration("ip_cams_to_cdf", cogClient, extractorMonitoringID, configObserver),
		eventbus:        eventBus,
		cameras:         make(map[uint64]*inputs.IpCamera),
	}
	return ingr
}

func (intgr *CameraImagesToCdf) GetEventBus() *pubsub.PubSub[string, camera.CameraEvent] {
	return intgr.eventbus
}
func (intgr *CameraImagesToCdf) SetCameraConfig(localConfig IntegrationConfig) {
	intgr.cameraConfigs = localConfig.Cameras
}

func (intgr *CameraImagesToCdf) LoadConfigFromJson(config json.RawMessage) error {
	var localConfig IntegrationConfig
	err := json.Unmarshal(config, &localConfig)
	if err != nil {
		log.Error("Failed to unmarshal local config with error : ", err.Error())
		return err
	}
	if localConfig.RetryCount == 0 {
		localConfig.RetryCount = 3
	}
	if localConfig.RetryInterval == 0 {
		localConfig.RetryInterval = 10
	}
	intgr.cameraConfigs = localConfig.Cameras
	intgr.integrationConfig = localConfig
	intgr.BaseIntegration.DisableRunReporting(localConfig.DisableRunReporting)
	log.Info("Integration config has been loaded successfully. Cameras count = ", len(intgr.cameraConfigs))
	return nil
}

func (intgr *CameraImagesToCdf) SetSecretManager(secretManager *internal.SecretManager) {
	intgr.secretManager = secretManager
}

func (intgr *CameraImagesToCdf) Start() error {
	intgr.IsRunning = true
	if intgr.cameraConfigs != nil && len(intgr.cameraConfigs) > 0 {
		intgr.startAllProcessors()

	} else {
		log.Info("Starting processing loop using remote configurations")
		configQueue := intgr.BaseIntegration.ConfigObserver.SubscribeToIntegrationConfigUpdates(intgr.BaseIntegration.ID)
		go func() {
			isFirstRemoteConfig := true
			for configAction := range configQueue {
				// log.Debugf("Old config for ingration : %+v\n ", intgr.integrationConfig)
				// log.Debugf("New config for ingration : %+v\n ", config)
				log.Info("Config has been changed . Restarting processors")
				if !isFirstRemoteConfig {
					intgr.BaseIntegration.ReportRunStatus("", core.ExtractionRunStatusSuccess, "Config has been changed . Restarting processors")
					intgr.StopAndClean()
				}
				intgr.LoadConfigFromJson(configAction.Config)
				intgr.startAllProcessors()
				isFirstRemoteConfig = false
			}
		}()
	}
	go intgr.startSelfMonitoring()
	return nil
}

func (intgr *CameraImagesToCdf) startAllProcessors() {
	intgr.IsRunning = true
	log.Info("Starting all camera processors")
	for _, camera := range intgr.cameraConfigs {
		if camera.State == "enabled" {
			go intgr.startSingleCameraProcessorLoop(camera)
		} else {
			log.Infof("Camera %s is disabled , operation skipped", camera.Name)
		}
	}
	log.Info("All camera processors have been started")
}

// startSelfMonitoring run a status reporting look that periodically sends status reports to pipeline monitoring
func (intgr *CameraImagesToCdf) startSelfMonitoring() {
	for {
		if intgr.successCounter > 0 && intgr.failureCounter == 0 {
			intgr.ReportRunStatus("", core.ExtractionRunStatusSuccess, fmt.Sprintf("Processed %d images", intgr.successCounter))
		} else if intgr.successCounter > 0 && intgr.failureCounter > 0 {
			intgr.ReportRunStatus("", core.ExtractionRunStatusSuccess, fmt.Sprintf("Processed %d images, %d failures", intgr.successCounter, intgr.failureCounter))
		} else {
			intgr.ReportRunStatus("", core.ExtractionRunStatusSeen, "")
		}
		intgr.successCounter = 0
		intgr.failureCounter = 0
		time.Sleep(time.Second * 60)
	}
}

func (intgr *CameraImagesToCdf) restartProcessor(camera CameraConfig) {
	intgr.BaseIntegration.StateTracker.SetProcessorTargetState(camera.ID, internal.ProcessorStateStopped)
	if intgr.BaseIntegration.StateTracker.WaitForProcessorTargetState(camera.ID, time.Second*120) {
		log.Infof("Processor %d has been stopped", camera.ID)
		intgr.startSingleCameraProcessorLoop(camera)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", camera.ID)
	}

}

// startProcessor starts camera processor , the operation is blocking and must be started in its own goroute
func (intgr *CameraImagesToCdf) startSingleCameraProcessorLoop(cameraConfig CameraConfig) error {
	log.Infof("Starting camera processor %s", cameraConfig.Name)
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("startProcessor failed to start with error : ", stack)
		}
		intgr.BaseIntegration.StateTracker.SetProcessorCurrentState(cameraConfig.ID, internal.ProcessorStateStopped)
	}()

	intgr.BaseIntegration.StateTracker.SetProcessorCurrentState(cameraConfig.ID, internal.ProcessorStateStarting)
	intgr.BaseIntegration.StateTracker.SetProcessorTargetState(cameraConfig.ID, internal.ProcessorStateRunning)
	var pollingInterval time.Duration

	log.Infof("Non-default polling interval = %d", cameraConfig.PollingInterval)
	pollingInterval = time.Duration(cameraConfig.PollingInterval) * time.Second

	if pollingInterval == 0 {
		pollingInterval = 60 * time.Second
	}

	log.Infof("Camera name = %s, model = %s, address = %s, username = %s, mode = %s", cameraConfig.Name, cameraConfig.Model, cameraConfig.Address, cameraConfig.Username, cameraConfig.Mode)

	if cameraConfig.Model == "" || cameraConfig.Address == "" {
		log.Errorf("Processor can't be started for camera %s . Model or address aren't set.", cameraConfig.Name)
		return fmt.Errorf("empty asset model or address")
	}
	cam := inputs.NewIpCamera(cameraConfig.ID, cameraConfig.Name, cameraConfig.Model, cameraConfig.Address, "", cameraConfig.Username, intgr.secretManager.GetSecret(cameraConfig.Password))
	if cam == nil {
		log.Error("Unsupported camera model")
		return fmt.Errorf("unsupported camera model")
	}
	intgr.cameras[cameraConfig.ID] = cam
	intgr.BaseIntegration.StateTracker.SetProcessorCurrentState(cameraConfig.ID, internal.ProcessorStateRunning)

	cameraEventFilters := make([]camera.EventFilter, len(cameraConfig.EventFilters))
	for i, filter := range cameraConfig.EventFilters {
		cameraEventFilters[i] = camera.EventFilter(filter)
	}

	err := intgr.SyncCameraServiceDiscoveryManistsWithCdf(cam)
	if err != nil {
		log.Error("Failed to sync cameras manifests with CDF. Err:", err.Error())
	}

	if cameraConfig.EnableCameraEventStream {
		go intgr.StartSingleCameraEventsProcessingLoop(cameraConfig.ID, cameraConfig.Name, cam, cameraEventFilters)
	}
	if pollingInterval < 0 {
		log.Infof("Polling interval is negative, processor %d will not run", cameraConfig.ID)
		return nil
	}
	for {

		intgr.executeProcessorRun(cameraConfig, cam, nil)

		if !intgr.IsRunning {
			break
		}
		// TODO : Randomize delays to distribute load
		time.Sleep(pollingInterval)
		st := intgr.BaseIntegration.StateTracker.GetProcessorState(cameraConfig.ID)
		if st == nil {
			break
		} else {
			if st.TargetState == internal.ProcessorStateStopped {
				break
			}
		}

		if cameraConfig.Mode == "camera+metadata" {
			intgr.executeCameraMetadataProcessorRun(cameraConfig, cam)
		}
	}
	log.Infof("Processor %d exited main loop ", cameraConfig.ID)
	intgr.BaseIntegration.StateTracker.SetProcessorCurrentState(cameraConfig.ID, internal.ProcessorStateStopped)
	return nil
}

// StartSingleCameraEventsProcessingLoop starts a loop to process events from a single camera.
// It subscribes to the events stream of the specified camera and publishes the events to the event bus and CDF.
// The loop continues until the IsRunning flag is set to false or an error occurs.
// Parameters:
//   - ID: The ID of the camera.
//   - name: The name of the camera.
//   - camera: The IP camera object.
//   - eventFilters: The filters to apply to the events stream.
//
// Returns:
//   - error: An error if the subscription to the events stream fails or if there is an error publishing the events to CDF.
func (intgr *CameraImagesToCdf) StartSingleCameraEventsProcessingLoop(ID uint64, name string, camera *inputs.IpCamera, eventFilters []camera.EventFilter) error {
	log.Infof("Starting camera events processor %s", name)
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("startProcessor failed to start with error : ", stack)
		}
	}()
	retryCount := 0
	for {
		stream, err := camera.SubscribeToEventsStream(eventFilters)
		if err != nil {
			retryCount++
			if retryCount > intgr.integrationConfig.RetryCount*10 {
				log.Errorf("Failed to subscribe to camera events stream %s after %d retries", name, intgr.integrationConfig.RetryCount)
				intgr.BaseIntegration.ReportRunStatus(name, core.ExtractionRunStatusFailure, fmt.Sprintf("extractor disconnected from event stream after %d retries. Camera name: %s ", intgr.integrationConfig.RetryCount, name))
				return err
			} else {
				log.Infof("Lost connection to camera %s event stream. Reconnecting ...", name)
				intgr.BaseIntegration.ReportRunStatus(name, core.ExtractionRunStatusFailure, fmt.Sprintf("Lost connection to camera %s event stream. Reconnecting ...", name))
				time.Sleep(time.Second * time.Duration(intgr.integrationConfig.RetryInterval))
				continue
			}
		}
		intgr.BaseIntegration.ReportRunStatus(name, core.ExtractionRunStatusSuccess, fmt.Sprintf("Connected to camera event stream.Camera name : %s", name))
		for event := range stream {
			log.Infof("Received event from camera %s : %s", name, event.Type)
			log.Debugf("Event data : %s", string(event.RawData))
			log.Debugf("Time from Axis WS stream: %d", event.Timestamp)
			topic := fmt.Sprintf("%d/%s", ID, event.Topic)
			intgr.eventbus.TryPub(event, topic)
			log.Debugf("Event published to event bus. Topic : %s", topic)
			corellationID := fmt.Sprintf("%d", event.Timestamp)
			cdfEvents := core.EventList{
				core.Event{
					StartTime:   event.Timestamp,
					EndTime:     event.Timestamp + 1,
					Type:        event.Type,
					Subtype:     event.CoreType,
					Description: "",
					Metadata:    map[string]string{"cameraName": name, "cameraId": fmt.Sprint(ID), "eventCorrelationId": corellationID, "topic": event.Topic, "rawData": string(event.RawData)},
					Source:      "edge-extractor:camera",
				},
			}
			_, err := intgr.CogClient.Client().Events.Create(cdfEvents)
			if err != nil {
				log.Errorf("Failed to publish event to CDF. Error : %s", err.Error())
				continue
			}
		}
		log.Infof("Camera events stream has been closed.Camera name : %s", name)
		if !intgr.IsRunning {
			log.Infof("Camera events processor %s has been stopped.Breaking stream retry loop.", name)
			break
		}
	}
	return nil
}

// ExecuteProcessorRunByCameraID executes the processor run for a specific camera ID.
// It retrieves the camera configuration and camera object based on the provided camera ID,
// and then calls the executeProcessorRun function to perform the actual processing.
// The metadata parameter is a map of additional information that can be passed to the processor.
// WARNING: This function is blocking and should be run in its own goroutine to avoid blocking the main application.
// Returns an error if any error occurs during the execution.
func (intgr *CameraImagesToCdf) ExecuteProcessorRunByCameraID(cameraID uint64, metadata map[string]string) error {
	camera := intgr.cameras[cameraID]
	cameraConfig := intgr.GetCameraConfigByID(cameraID)
	return intgr.executeProcessorRun(*cameraConfig, camera, metadata)
}

func (intgr *CameraImagesToCdf) GetCameraConfigByID(cameraID uint64) *CameraConfig {
	for i := range intgr.cameraConfigs {
		if intgr.cameraConfigs[i].ID == cameraID {
			return &intgr.cameraConfigs[i]
		}
	}
	return nil
}

// executeProcessorRun executes single processor run (full process) , the operation is blocking and must be started in its own goroute for low latency and high throughput
func (intgr *CameraImagesToCdf) executeProcessorRun(camera CameraConfig, cam *inputs.IpCamera, metadata map[string]string) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("executeProcessorRun crashed with error : ", stack)
			intgr.failureCounter++
			intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("executeProcessorRun crashed with error :%s", stack))
		}
	}()

	img, err := cam.ExtractImage()
	if metadata != nil {
		metadata["capturedAt"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	if err != nil {
		log.Errorf("Can't extract image from camera  %s  . Error : %s", camera.Name, err.Error())
		intgr.failureCounter++
		intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to extract img, err :%s", err.Error()))
		time.Sleep(time.Second * 20)
	} else {
		if img == nil {
			time.Sleep(time.Second * 1)
			return nil
		}

		timeStamp := time.Now().Format("2006-01-02T15:04:05.999")
		externalId := fmt.Sprintf("%s_%d", camera.Name, time.Now().UnixNano())
		fileName := camera.Name + " " + timeStamp + ".jpeg"
		retryCount := 0
		for {
			err := intgr.BaseIntegration.CogClient.UploadInMemoryFile(img.Body, externalId, fileName, img.Format, camera.LinkedAssetID, metadata)
			if err != nil {
				if strings.Contains(err.Error(), "Duplicate external ids") {
					log.Info("Duplicate external ids error. Errror ignored. Error : ", err.Error())
					intgr.successCounter++
					break
				} else {
					log.Error("Failed to upload image to CDF. Error : ", err.Error())
				}
				intgr.failureCounter++
				intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to upload img, err :%s", err.Error()))
				retryCount++
				if !intgr.IsRunning || retryCount > intgr.integrationConfig.RetryCount {
					break
				}
				time.Sleep(time.Second * time.Duration(intgr.integrationConfig.RetryInterval*retryCount))
			} else {
				log.Debug("File uploaded to CDF successfully")
				intgr.successCounter++
				break
			}
		}
	}
	return err
}

func (intgr *CameraImagesToCdf) executeCameraMetadataProcessorRun(camera CameraConfig, cam *inputs.IpCamera) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("Camera metadata extraction crashed with error : ", stack)
			intgr.failureCounter++
			intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("camera metadata extraction crashed with error :%s", stack))
		}
	}()

	bmeta, err := cam.ExtractMetadata()
	if err == nil {
		log.Debug("Fetching Metadata from camera:")
		log.Debug(string(bmeta))
		// This is just a test.
		intgr.BaseIntegration.ReportRunStatus("", core.ExtractionRunStatusSuccess, string(bmeta))
	} else {
		log.Info("Failed to extract metadata . Err :", err.Error())
	}
	return err
}

func (intgr *CameraImagesToCdf) StopAndClean() error {
	intgr.IsRunning = false
	log.Info("Stopping all camera processors")
	for ID, camera := range intgr.cameras {
		camera.Close()
		intgr.BaseIntegration.StopProcessor(ID)
		delete(intgr.cameras, ID)
	}
	log.Info("All camera processors have been stopped")

	return nil
}

func (intgr *CameraImagesToCdf) SyncCameraServiceDiscoveryManistsWithCdf(camera *inputs.IpCamera) error {
	manifests, err := camera.GetServicesDiscoveryManifest("all")
	if err != nil {
		log.Errorf("Failed to get services discovery manifest . Error : %s", err.Error())
		return err
	}
	if manifests == nil {
		log.Debug("Services discovery manifest is empty")
		return nil
	}

	for _, manifest := range manifests {
		externalId := fmt.Sprintf("camera_%d_discovery_manifest", camera.ID)
		fileName := fmt.Sprintf("camera_%s_discovery_manifest_%s", camera.Name, manifest.Name)
		err := intgr.BaseIntegration.CogClient.UploadInMemoryFile(manifest.Body, externalId, fileName, "", 0, nil)
		if err != nil {
			log.Infof("Failed to upload services discovery manifest to CDF. Error : %s", err.Error())
		}
		log.Infof("Services discovery manifest %s has been uploaded to CDF", manifest.Name)
	}
	return nil
}
