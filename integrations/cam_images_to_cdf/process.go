package cam_images_to_cdf

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/connectors/inputs"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)


type CameraImagesToCdf struct {
	internal.Processor
	cameraConfigs []CameraConfig
	secretManager *internal.SecretManager
}

func NewCameraImagesToCdf(cogClient *internal.CdfClient, extractorMonitoringID string, remoteConfigSource string) *CameraImagesToCdf {
	ingr := &CameraImagesToCdf{
		internal.Processor{cogClient: cogClient, globalCamPollingInterval: time.Second * 30, stateTracker: internal.NewStateTracker(), extractorID: extractoMonitoringID}
	}
	ingr.Processor.ConfigObserver = internal.NewCdfConfigObserver(extractorMonitoringID, cogClient, remoteConfigSource)
	return ingr
}

func (intgr *CameraImagesToCdf) SetLocalConfig(localConfig CameraImagesToCdfConfig) {
	intgr.cameraConfigs = localConfig.Cameras
}

func (intgr *CameraImagesToCdf) SetSecretManager(secretManager *internal.SecretManager) {
	intgr.secretManager = secretManager
	
}

func (intgr *CameraImagesToCdf) Start() error {
	intgr.isStarted = true
	if intgr.cameraConfigs != nil {
		log.Info("Starting processing loop using local configurations")
		for _, camera := range intgr.cameraConfigs {
			go intgr.startSingleCameraProcessorLoop(camera)
		}

	} else {
		log.Info("Starting processing loop using remote configurations")
		filter := core.AssetFilter{Metadata: map[string]string{"cog_class": "camera", "extractor_id": intgr.extractorID}}
		actionQueue := intgr.configObserver.Start(filter, 60*time.Second)
		go func() {
			for configAction := range actionQueue {
				log.Debugf("New config action event . Action ID = %d ", configAction.Name)
				config := configAction.Config.(CameraConfig)
				switch configAction.Name {
				case internal.RestartProcessorAction:
					// this means metada has been changed .
					if config.State == "enabled" {
						go intgr.restartProcessor(config)
					} else {
						log.Info("Camera has been disabled , sending STOP signal to processor")
						go intgr.stopProcessor(configAction.ProcId)
					}

				case internal.StartProcessorLoopAction:
					if config.State == "enabled" {
						log.Info("Camera has been enabled , sending START signal to processor")
						go intgr.startSingleCameraProcessorLoop(config)
					} else {
						log.Info("Camera is disabled , operation skipped")
					}

				case internal.StopProcessorAction:
					go intgr.stopProcessor(configAction.ProcId)
				default:
					log.Infof("Unknown cofig action %+v", configAction)
				}
			}
		}()
	}
	go intgr.startSelfMonitoring()
	return nil
}

func (intgr *CameraImagesToCdf) restartProcessor(camera CameraConfig) {
	intgr.stateTracker.SetProcessorTargetState(camera.ID, internal.ProcessorStateStopped)
	if intgr.stateTracker.WaitForProcessorTargetState(camera.ID, time.Second*120) {
		log.Infof("Processor %d has been stopped", camera.ID)
		intgr.startSingleCameraProcessorLoop(camera)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", camera.ID)
	}

}

// startProcessor starts camera processor , the operation is blocking and must be started in its own goroute
func (intgr *CameraImagesToCdf) startSingleCameraProcessorLoop(camera CameraConfig) error {
	log.Infof("Starting camera processor %s", camera.Name)
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("startProcessor failed to start with error : ", stack)
		}
		intgr.stateTracker.SetProcessorCurrentState(camera.ID, internal.ProcessorStateStopped)
	}()

	intgr.stateTracker.SetProcessorCurrentState(camera.ID, internal.ProcessorStateStarting)
	intgr.stateTracker.SetProcessorTargetState(camera.ID, internal.ProcessorStateRunning)

	// TODO : Investigate memmory usage with many cameras and big images . Next step is to introduce image streaming through file system to avoid excessive RAM usag.
	// model := asset.Metadata["cog_model"]
	// address := asset.Metadata["uri"]
	// username := asset.Metadata["username"]
	// password := asset.Metadata["password"]
	// mode := asset.Metadata["mode"]

	// pollingIntervalTmp, err := strconv.Atoi(asset.Metadata["polling_interval"]) // polling interval in seconds
	var pollingInterval time.Duration

	log.Infof("Non-default polling interval = %d", camera.PollingInterval)
	pollingInterval = time.Duration(camera.PollingInterval) * time.Second

	if pollingInterval < 1 {
		pollingInterval = intgr.globalCamPollingInterval
	}

	log.Infof("Camera name = %s, model = %s, address = %s, username = %s, mode = %s", camera.Name, camera.Model, camera.Address, camera.Username, camera.Mode)

	if camera.Model == "" || camera.Address == "" {
		log.Errorf("Processor can't be started for camera %s . Model or address aren't set.", camera.Name)
		return fmt.Errorf("empty asset model or address")
	}

	cam := inputs.NewIpCamera(camera.Model, camera.Address, "", camera.Username, camera.Password)
	if cam == nil {
		log.Error("Unsupported camera model")
		return fmt.Errorf("unsupported camera model")
	}
	intgr.stateTracker.SetProcessorCurrentState(camera.ID, internal.ProcessorStateRunning)
	for {

		intgr.executeProcessorRun(camera, cam)

		if !intgr.isStarted {
			break
		}
		// TODO : Randomize delays to distribute load
		time.Sleep(pollingInterval)
		st := intgr.stateTracker.GetProcessorState(camera.ID)
		if st == nil {
			break
		} else {
			if st.TargetState == internal.ProcessorStateStopped {
				break
			}
		}

		if camera.Mode == "camera+metadata" {
			intgr.executeCameraMetadataProcessorRun(camera, cam)
		}
	}
	log.Infof("Processor %d exited main loop ", camera.ID)
	return nil
}

// executeProcessorRun executes single processor run (full process) , the operation is blocking and must be started in its own goroute
func (intgr *CameraImagesToCdf) executeProcessorRun(camera CameraConfig, cam *inputs.IpCamera) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("executeProcessorRun crashed with error : ", stack)
			intgr.failureCounter++
			intgr.reportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("executeProcessorRun crashed with error :%s", stack))
		}
	}()

	img, err := cam.ExtractImage()
	if err != nil {
		log.Debugf("Can't extract image from camera  %s  . Error : %s", camera.Name, err.Error())
		intgr.failureCounter++
		intgr.reportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to extract img, err :%s", err.Error()))
		time.Sleep(time.Second * 60)
	} else {
		if img == nil {
			time.Sleep(time.Second * 10)
			return nil
		}

		timeStamp := time.Now().Format(time.RFC3339)
		externalId := camera.Name + " " + timeStamp
		fileName := externalId + ".jpeg"
		err := intgr.cogClient.UploadInMemoryFile(img.Body, externalId, fileName, img.Format, camera.ID)
		if err != nil {
			log.Debug("Can't upload image. Error : ", err.Error())
			intgr.failureCounter++
			intgr.reportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to upload img, err :%s", err.Error()))
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
			intgr.reportRunStatus(camera.Name, core.ExtractionRunStatusFailure, fmt.Sprintf("camera metadata extraction crashed with error :%s", stack))
		}
	}()

	bmeta, err := cam.ExtractMetadata()
	if err == nil {
		log.Debug("Fetching Metadata from camera:")
		log.Debug(string(bmeta))
		// This is just a test.
		intgr.reportRunStatus("", core.ExtractionRunStatusSuccess, string(bmeta))
	} else {
		log.Info("Failed to extract metadata . Err :", err.Error())
	}
	return err
}
