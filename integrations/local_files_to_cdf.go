package integrations

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/connectors/inputs"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)

type LocalFilesToCdf struct {
	cogClient                *internal.CdfClient
	isStarted                bool
	stateTracker             *internal.StateTracker
	globalCamPollingInterval time.Duration
	successCounter           uint64
	failureCounter           uint64
	extractorID              string
	configObserver           *internal.CdfConfigObserver // remote config observer
	localConfig              []core.Asset                // local configuration
}

func NewLocalFilesToCdf(cogClient *internal.CdfClient, extractoMonitoringID string, remoteConfigSource string) *LocalFilesToCdf {
	ingr := &LocalFilesToCdf{cogClient: cogClient, globalCamPollingInterval: time.Second * 30, stateTracker: internal.NewStateTracker(), extractorID: extractoMonitoringID}
	ingr.configObserver = internal.NewCdfConfigObserver(extractoMonitoringID, cogClient, remoteConfigSource)
	return ingr
}

func (intgr *LocalFilesToCdf) SetLocalConfig(localConfig []core.Asset) {
	intgr.localConfig = localConfig
}

func (intgr *LocalFilesToCdf) Start() error {
	intgr.isStarted = true
	if intgr.localConfig != nil {
		log.Info("Starting LocalFilesToCdf processing loop using local configurations")
		for _, asset := range intgr.localConfig {
			go intgr.startSingleCameraProcessorLoop(asset)
		}
	} else {
		log.Info("Starting LocalFilesToCdf processing loop using remote configurations")
		filter := core.AssetFilter{Metadata: map[string]string{"cog_class": "camera", "extractor_id": intgr.extractorID}}
		actionQueue := intgr.configObserver.Start(filter, 60*time.Second)
		go func() {
			for configAction := range actionQueue {
				log.Debugf("New config action event . Action ID = %d ", configAction.Name)
				switch configAction.Name {
				case internal.RestartProcessorAction:
					// this means metada has been changed .
					if configAction.Asset.Metadata["state"] == "enabled" {
						go intgr.restartProcessor(configAction.Asset)
					} else {
						log.Info("Camera has been disabled , sending STOP signal to processor")
						go intgr.stopProcessor(configAction.ProcId)
					}

				case internal.StartProcessorLoopAction:
					if configAction.Asset.Metadata["state"] == "enabled" {
						go intgr.startSingleCameraProcessorLoop(configAction.Asset)
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

func (intgr *LocalFilesToCdf) Stop() {
	intgr.isStarted = false
}

func (intgr *LocalFilesToCdf) restartProcessor(asset core.Asset) {
	intgr.stateTracker.SetProcessorTargetState(asset.ID, internal.ProcessorStateStopped)
	if intgr.stateTracker.WaitForProcessorTargetState(asset.ID, time.Second*120) {
		log.Infof("Processor %d has been stopped", asset.ID)
		intgr.startSingleCameraProcessorLoop(asset)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", asset.ID)
	}

}

func (intgr *LocalFilesToCdf) stopProcessor(procId uint64) {
	log.Infof("Sending stop signal to processor %d ", procId)
	intgr.stateTracker.SetProcessorTargetState(procId, internal.ProcessorStateStopped)
	if intgr.stateTracker.WaitForProcessorTargetState(procId, time.Second*120) {
		log.Infof("Processor %d has been stopped", procId)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", procId)
	}
}

func (intgr *LocalFilesToCdf) reportRunStatus(camExternalID, status, msg string) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		log.Error(" Pipeliene monitoring failed to load configuration from CDF with error : ", stack)
	}
	exRun := core.CreateExtractionRun{ExternalID: intgr.extractorID, Status: status, Message: msg}
	intgr.cogClient.Client().ExtractionPipelines.CreateExtractionRuns(core.CreateExtractonRunsList{exRun})
}

// startSelfMonitoring run a status reporting look that periodically sends status reports to pipeline monitoring
func (intgr *LocalFilesToCdf) startSelfMonitoring() {
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

// startProcessor starts camera processor , the operation is blocking and must be started in its own goroute
func (intgr *LocalFilesToCdf) startSingleCameraProcessorLoop(asset core.Asset) error {
	log.Infof("Starting camera processor %d", asset.ID)
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("startProcessor failed to start with error : ", stack)
		}
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

	pollingIntervalTmp, err := strconv.Atoi(asset.Metadata["polling_interval"]) // polling interval in seconds
	var pollingInterval time.Duration

	if err == nil {
		log.Infof("Non-default polling interval = %d", pollingIntervalTmp)
		pollingInterval = time.Duration(pollingIntervalTmp) * time.Second
	}

	if pollingInterval < 1 && model != "fscam" {
		pollingInterval = intgr.globalCamPollingInterval
	}

	maxParallelRuns, _ := strconv.Atoi(asset.Metadata["max_parallel_runs"]) // polling interval in seconds
	if maxParallelRuns == 0 {
		maxParallelRuns = 1
	}

	log.Infof("Camera name = %s, model = %s, address = %s, username = %s, mode = %s, max parallel runs = %d", asset.Name, model, address, username, mode, maxParallelRuns)

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
	var concurentRunsCounter atomic.Int32
	concurentRunsCounter.Store(0)
	for {
		if !intgr.isStarted {
			break
		}
		runner := func() {
			intgr.executeProcessorRun(asset, cam)
			if mode == "camera+metadata" {
				intgr.executeCameraMetadataProcessorRun(asset, cam)
			}
			concurentRunsCounter.Add(-1)
		}
		if maxParallelRuns > 1 {
			if concurentRunsCounter.Load() >= int32(maxParallelRuns) {
				log.Debugf("Max parallel runs reached for camera %s . Waiting for previous runs to finish. Max runs = %d , counter = %d", asset.Name, maxParallelRuns, concurentRunsCounter.Load())
				time.Sleep(time.Second * 2)
				continue
			} else {
				// Concurent worker
				concurentRunsCounter.Add(1)
				go runner()
			}
		} else {
			// Single worker
			concurentRunsCounter.Add(1)
			runner()
		}
		// TODO : Randomize delays to distribute load
		if pollingInterval > 0 {
			time.Sleep(pollingInterval)
		}
		st := intgr.stateTracker.GetProcessorState(asset.ID)
		if st == nil {
			break
		} else {
			if st.TargetState == internal.ProcessorStateStopped {
				break
			}
		}

	}
	log.Infof("Processor %d exited main loop ", asset.ID)
	return nil
}

// executeProcessorRun
func (intgr *LocalFilesToCdf) executeProcessorRun(asset core.Asset, cam *inputs.IpCamera) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("executeProcessorRun crashed with error : ", stack)
			intgr.failureCounter++
			intgr.reportRunStatus(asset.ExternalID, core.ExtractionRunStatusFailure, fmt.Sprintf("executeProcessorRun crashed with error :%s", stack))
		}
	}()

	img, err := cam.ExtractImage()
	log.Debugf("Image extracted from camera ")
	if err != nil {
		log.Debugf("Can't extract image from camera  %s  . Error : %s", asset.Name, err.Error())
		intgr.failureCounter++
		intgr.reportRunStatus(asset.ExternalID, core.ExtractionRunStatusFailure, fmt.Sprintf("failed to extract img, err :%s", err.Error()))
		time.Sleep(time.Second * 60)
	} else {
		if img == nil {
			time.Sleep(time.Second * 10)
			return nil
		}

		timeStamp := time.Now().Format(time.RFC3339)

		externalId := asset.Name + "_" + timeStamp + "_" + img.ExternalId
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
			cam.Commit(img.TransactionId)
		}
	}
	return err
}

func (intgr *LocalFilesToCdf) executeCameraMetadataProcessorRun(asset core.Asset, cam *inputs.IpCamera) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("Camera metadata extraction crashed with error : ", stack)
			intgr.failureCounter++
			intgr.reportRunStatus(asset.ExternalID, core.ExtractionRunStatusFailure, fmt.Sprintf("camera metadata extraction crashed with error :%s", stack))
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
