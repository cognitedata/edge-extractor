package integrations

import (
	"fmt"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/connectors/inputs"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)

type CameraImagesToCdfConfig struct {
}

type CameraImagesToCdf struct {
	cogClient                *internal.CdfClient
	isStarted                bool
	cameraAssets             core.AssetList
	stateTracker             *internal.StateTracker
	globalCamPollingInterval time.Duration
	successCounter           uint64
	failureCounter           uint64
	extractorID              string
}

func NewCameraImagesToCdf(cogClient *internal.CdfClient, extractoMonitoringID string) *CameraImagesToCdf {
	return &CameraImagesToCdf{cogClient: cogClient, globalCamPollingInterval: time.Second * 20, stateTracker: internal.NewStateTracker(), extractorID: extractoMonitoringID}
}

// Introduce reload command.

func (intgr *CameraImagesToCdf) Start() error {
	intgr.isStarted = true

	go intgr.startCdfConfigPolling()
	go intgr.startSelfMonitoring()
	return nil
}

func (intgr *CameraImagesToCdf) Stop() {
	intgr.isStarted = false
}

func (intgr *CameraImagesToCdf) startCdfConfigPolling() {
	for {
		intgr.ReloadRemoteConfigs()
		time.Sleep(time.Second * 60)
		if !intgr.isStarted {
			break
		}
	}
}

// LoadConfigs load both static and dynamic configs
func (intgr *CameraImagesToCdf) ReloadRemoteConfigs() error {
	log.Debug("Reloading remote config")
	filter := core.AssetFilter{Metadata: map[string]string{"cog_class": "camera", "state": "enabled", "extractor_id": intgr.extractorID}}

	remoteAssetList, err := intgr.cogClient.Client().Assets.Filter(filter, 1000)
	if err != nil {
		return err
	}
	// comparing existing assets with assets in cdf , reloading processor if there is a difference
	localAssetList := intgr.cameraAssets

	// 1. asset is not present in master list - new asset has been added in CDF. Action - start new processor
	// 2. asset is not present in slave list - asset has been deleted . Action - stop processor
	// 3. asset present in both lists but metadata is defferent - asset has been updated . Action reload processor
	// 4. assets are equal - Do nothing

	var updatedRemoteAsset core.Asset
	var isUpdated bool

	for i := range localAssetList {
		isEqual := false
		isPresentRemotely := false
		for i2 := range remoteAssetList {
			if localAssetList[i].ID == remoteAssetList[i2].ID {
				isPresentRemotely = true
				if intgr.cogClient.CompareAssets(localAssetList[i], remoteAssetList[i2]) {
					isEqual = true
				} else {
					updatedRemoteAsset = remoteAssetList[i2]
				}
				break
			}
		}

		if !isPresentRemotely {
			log.Infof("Remote change detected. Removing processor %d ", localAssetList[i].ID)
			isUpdated = true
			go intgr.stopProcessor(localAssetList[i].ID)

		} else if isPresentRemotely && !isEqual {
			log.Infof("Remote change detected. Restarting and updating processor %d", localAssetList[i].ID)
			// Reload processor
			isUpdated = true
			go intgr.restartProcessor(updatedRemoteAsset)
		}
	}

	// Comparing remote assets with local , starting processors for all new remote assets
	for i3 := range remoteAssetList {
		isPresent := false
		for i4 := range localAssetList {
			if localAssetList[i4].ID == remoteAssetList[i3].ID {
				isPresent = true
				break
			}
		}
		if !isPresent {
			isUpdated = true
			log.Infof("Remote change detected. Starting new processor %d for camera %s", remoteAssetList[i3].ID, remoteAssetList[i3].Name)
			go intgr.startProcessor(remoteAssetList[i3])
		}

	}
	if isUpdated {
		intgr.cameraAssets = remoteAssetList
	}

	return nil
}

func (intgr *CameraImagesToCdf) restartProcessor(asset core.Asset) {
	intgr.stateTracker.SetProcessorTargetState(asset.ID, internal.ProcessorStateStopped)
	if intgr.stateTracker.WaitForProcessorTargetState(asset.ID, time.Second*120) {
		log.Infof("Processor %d has been stopped", asset.ID)
		intgr.startProcessor(asset)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", asset.ID)
	}

}

func (intgr *CameraImagesToCdf) stopProcessor(procId uint64) {
	log.Infof("Sending stop signal to processor %d ", procId)
	intgr.stateTracker.SetProcessorTargetState(procId, internal.ProcessorStateStopped)
	if intgr.stateTracker.WaitForProcessorTargetState(procId, time.Second*120) {
		log.Infof("Processor %d has been stopped", procId)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", procId)
	}
}

func (intgr *CameraImagesToCdf) reportRunStatus(camExternalID, status, msg string) {
	exRun := core.CreateExtractionRun{ExternalID: intgr.extractorID, Status: status, Message: msg}

	intgr.cogClient.Client().ExtractionPipelines.CreateExtractionRuns(core.CreateExtractonRunsList{exRun})
}

func (intgr *CameraImagesToCdf) startSelfMonitoring() {
	for {
		if intgr.successCounter > 0 && intgr.failureCounter == 0 {
			intgr.reportRunStatus("", core.ExtractionRunStatusSuccess, "all cameras operational")
		} else if intgr.successCounter > 0 && intgr.failureCounter > 0 {
			intgr.reportRunStatus("", core.ExtractionRunStatusSuccess, "some cameras not operational")
		} else {
			intgr.reportRunStatus("", core.ExtractionRunStatusSeen, "")
		}
		intgr.successCounter = 0
		intgr.failureCounter = 0
		time.Sleep(time.Second * 60)
	}
}

// Start camera processor.
func (intgr *CameraImagesToCdf) startProcessor(asset core.Asset) error {
	log.Infof("Starting camera processor %d", asset.ID)
	defer func() {
		intgr.stateTracker.SetProcessorCurrentState(asset.ID, internal.ProcessorStateStopped)
	}()

	intgr.stateTracker.SetProcessorCurrentState(asset.ID, internal.ProcessorStateStarting)
	intgr.stateTracker.SetProcessorTargetState(asset.ID, internal.ProcessorStateRunning)

	// TODO : Investigate memmory usage with many cameras and big images . Next step is to introduce image streaming through file system to avoid excessive RAM usag.
	model := asset.Metadata["cog_model"]
	address := asset.Metadata["uri"]
	username := asset.Metadata["username"]
	password := asset.Metadata["password"]
	mode := asset.Metadata["mode"]

	log.Infof(" Camera name = %s , model = %s , address = %s , username = %s , password = %s , mode = %s", asset.Name, model, address, username, password, mode)

	if model == "" || address == "" {
		log.Errorf("Processor can't be started for camera %s . Model or address aren't set.", asset.Name)
		return fmt.Errorf("empty asset model or address")
	}

	cam := inputs.NewIpCamera(model, address, "", username, password)
	if cam == nil {
		log.Error("Unsupported camera model")
		return fmt.Errorf("unsupported camera model")
	}
	intgr.stateTracker.SetProcessorCurrentState(asset.ID, internal.ProcessorStateRunning)
	for {
		// TODO : ADD extractor monitoring here.
		img, err := cam.ExtractImage()
		if err != nil {
			log.Debugf("Can't extract image from camera model %s  . Error : %s", model, err.Error())
			intgr.failureCounter++
			intgr.reportRunStatus(asset.ExternalID, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to extract img, err :%s", err.Error()))
			time.Sleep(time.Second * 60)
		} else {
			timeStamp := time.Now().Format(time.RFC3339)
			externalId := asset.Name + " " + timeStamp
			fileName := externalId + ".jpeg"
			err := intgr.cogClient.UploadInMemoryFile(img.Body, externalId, fileName, img.Format, asset.ID)
			if err != nil {
				log.Debug("Can't upload image. Error : ", err.Error())
				intgr.failureCounter++
				intgr.reportRunStatus(asset.ExternalID, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to upload img, err :%s", err.Error()))
				time.Sleep(time.Second * 60)
			} else {
				log.Debug("File uploaded to CDF successfully")
				intgr.successCounter++
			}
		}
		if !intgr.isStarted {
			break
		}
		time.Sleep(intgr.globalCamPollingInterval)
		st := intgr.stateTracker.GetProcessorState(asset.ID)
		if st == nil {
			break
		} else {
			if st.TargetState == internal.ProcessorStateStopped {
				break
			}
		}

		if mode == "camera+metadata" {
			bmeta, err := cam.ExtractMetadata()
			if err == nil {
				log.Info("Fetching Metadata from camera:")
				log.Info(string(bmeta))
				intgr.reportRunStatus("", core.ExtractionRunStatusSuccess, string(bmeta))
			} else {
				log.Info("Failed to extract metadata . Err :", err.Error())
			}
		}

	}
	log.Infof("Processor %d exited main loop ", asset.ID)
	return nil

}
