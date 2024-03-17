package ip_cams_to_cdf

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/connectors/inputs"
	"github.com/cognitedata/edge-extractor/drivers/camera"
	"github.com/cognitedata/edge-extractor/integrations"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)

type CameraImagesToCdf struct {
	integrations.BaseIntegration
	successCounter    uint64
	failureCounter    uint64
	cameraConfigs     []CameraConfig
	secretManager     *internal.SecretManager
	integrationConfig IntegrationConfig
}

func NewCameraImagesToCdf(cogClient *internal.CdfClient, extractorMonitoringID string, configObserver *internal.CdfConfigObserver) *CameraImagesToCdf {
	ingr := &CameraImagesToCdf{
		BaseIntegration: *integrations.NewIntegration("ip_cams_to_cdf", cogClient, extractorMonitoringID, configObserver),
	}
	return ingr
}

func (intgr *CameraImagesToCdf) SetConfig(localConfig IntegrationConfig) {
	intgr.cameraConfigs = localConfig.Cameras
}

func (intgr *CameraImagesToCdf) LoadConfigFromJson(config json.RawMessage) error {
	var localConfig IntegrationConfig
	err := json.Unmarshal(config, &localConfig)
	if err != nil {
		log.Error("Failed to unmarshal local config with error : ", err.Error())
		return err
	}
	intgr.cameraConfigs = localConfig.Cameras
	intgr.integrationConfig = localConfig
	log.Info("Local config has been loaded successfully. Cameras count = ", len(intgr.cameraConfigs))
	return nil
}

func (intgr *CameraImagesToCdf) SetSecretManager(secretManager *internal.SecretManager) {
	intgr.secretManager = secretManager
}

func (intgr *CameraImagesToCdf) Start() error {
	intgr.IsRunning = true
	if intgr.cameraConfigs != nil && len(intgr.cameraConfigs) > 0 {
		log.Info("Starting processing loop using local configurations")
		for _, camera := range intgr.cameraConfigs {
			if camera.State == "enabled" {
				go intgr.startSingleCameraProcessorLoop(camera)
			} else {
				log.Infof("Camera %s is disabled , operation skipped", camera.Name)
			}
		}

	} else {
		log.Info("Starting processing loop using remote configurations")
		configQueue := intgr.BaseIntegration.ConfigObserver.SubscribeToConfigUpdates(intgr.BaseIntegration.ID, &IntegrationConfig{})
		go func() {
			for configAction := range configQueue {
				config := configAction.Config.(*IntegrationConfig)
				// log.Debugf("Old config for ingration : %+v\n ", intgr.integrationConfig)
				// log.Debugf("New config for ingration : %+v\n ", config)

				if !intgr.integrationConfig.IsEqual(config) {
					log.Info("Config has been changed . Restarting processors")
					// Stopping all processors
					for _, camera := range intgr.cameraConfigs {
						intgr.BaseIntegration.StopProcessor(camera.ID)
					}
					intgr.integrationConfig = config.Clone()
					intgr.cameraConfigs = config.Cameras
					for _, camera := range intgr.cameraConfigs {
						if camera.State == "enabled" {
							go intgr.startSingleCameraProcessorLoop(camera)
						} else {
							log.Infof("Camera %s is disabled , operation skipped", camera.Name)
						}

					}
				}
			}
		}()
	}
	go intgr.startSelfMonitoring()
	return nil
}

// startSelfMonitoring run a status reporting look that periodically sends status reports to pipeline monitoring
func (intgr *CameraImagesToCdf) startSelfMonitoring() {
	for {
		if intgr.successCounter > 0 && intgr.failureCounter == 0 {
			intgr.ReportRunStatus("", core.ExtractionRunStatusSuccess, "all cameras operational")
		} else if intgr.successCounter > 0 && intgr.failureCounter > 0 {
			intgr.ReportRunStatus("", core.ExtractionRunStatusSuccess, "some cameras not operational")
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

	if pollingInterval < 1 {
		pollingInterval = 60 * time.Second
	}

	log.Infof("Camera name = %s, model = %s, address = %s, username = %s, mode = %s", cameraConfig.Name, cameraConfig.Model, cameraConfig.Address, cameraConfig.Username, cameraConfig.Mode)

	if cameraConfig.Model == "" || cameraConfig.Address == "" {
		log.Errorf("Processor can't be started for camera %s . Model or address aren't set.", cameraConfig.Name)
		return fmt.Errorf("empty asset model or address")
	}
	cam := inputs.NewIpCamera(cameraConfig.Model, cameraConfig.Address, "", cameraConfig.Username, intgr.secretManager.GetSecret(cameraConfig.Password))
	if cam == nil {
		log.Error("Unsupported camera model")
		return fmt.Errorf("unsupported camera model")
	}
	intgr.BaseIntegration.StateTracker.SetProcessorCurrentState(cameraConfig.ID, internal.ProcessorStateRunning)

	cameraEventFilters := make([]camera.EventFilter, len(cameraConfig.EventFilters))
	for i, filter := range cameraConfig.EventFilters {
		cameraEventFilters[i] = camera.EventFilter(filter)
	}

	if cameraConfig.EnableCameraEventStream {
		go intgr.StartSingleCameraEventsProcessingLoop(cameraConfig.Name, cam, cameraEventFilters)
	}
	for {

		intgr.executeProcessorRun(cameraConfig, cam)

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

func (intgr *CameraImagesToCdf) StartSingleCameraEventsProcessingLoop(name string, camera *inputs.IpCamera, eventFilters []camera.EventFilter) error {
	log.Infof("Starting camera events processor %s", name)
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("startProcessor failed to start with error : ", stack)
		}
	}()

	stream, err := camera.SubscribeToEventsStream(eventFilters)
	if err != nil {
		log.Errorf("Failed to subscribe to camera events stream. Error : %s", err.Error())
		return err
	}
	for event := range stream {
		log.Infof("Received event from camera %s : %s", name, event.Type)
		log.Debugf("Event data : %s", string(event.Data))
	}
	log.Infof("Camera events stream processor %s exited main loop ", name)
	return nil
}

// executeProcessorRun executes single processor run (full process) , the operation is blocking and must be started in its own goroute for low latency and high throughput
func (intgr *CameraImagesToCdf) executeProcessorRun(camera CameraConfig, cam *inputs.IpCamera) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("executeProcessorRun crashed with error : ", stack)
			intgr.failureCounter++
			intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("executeProcessorRun crashed with error :%s", stack))
		}
	}()

	img, err := cam.ExtractImage()
	if err != nil {
		log.Debugf("Can't extract image from camera  %s  . Error : %s", camera.Name, err.Error())
		intgr.failureCounter++
		intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to extract img, err :%s", err.Error()))
		time.Sleep(time.Second * 60)
	} else {
		if img == nil {
			time.Sleep(time.Second * 10)
			return nil
		}

		timeStamp := time.Now().Format(time.RFC3339)
		externalId := camera.Name + " " + timeStamp
		fileName := externalId + ".jpeg"
		err := intgr.BaseIntegration.CogClient.UploadInMemoryFile(img.Body, externalId, fileName, img.Format, camera.LinkedAssetID)
		if err != nil {
			log.Debug("Can't upload image. Error : ", err.Error())
			intgr.failureCounter++
			intgr.BaseIntegration.ReportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to upload img, err :%s", err.Error()))
			time.Sleep(time.Second * 60)
		} else {
			log.Debug("File uploaded to CDF successfully")
			intgr.successCounter++
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
