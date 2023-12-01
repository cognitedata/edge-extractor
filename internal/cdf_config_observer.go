package internal

import (
	"encoding/json"
	"runtime/debug"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	log "github.com/sirupsen/logrus"
)

type StartProcessor func(asset core.Asset) error
type RestartProcessor func(asset core.Asset)
type StartProcessorLoop func(asset core.Asset) error
type StopProcessor func(procId uint64)

// const StartProcessorAction = 1
const RestartProcessorAction = 2
const StopProcessorAction = 3
const StartProcessorLoopAction = 4

// CdfConfigObserver is a component that is responsible for monitoring filtererd subset of CDF Assets for changes , if change being detected it restarts or stops linked processor .
// If observer detects that the asset is present remotely and not locally , it calls startProcessorLoop function
// If observer detects that the asset is not present remotelyu is calls stopProcessor function
// If observer detect that Asset name or external_id or metadata have been chagned,it calls restartProcessor function.
type CdfConfigObserver struct {
	extractorID        string
	isStarted          bool
	cogClient          *CdfClient
	localAssetsList    core.AssetList
	assetFilter        core.AssetFilter
	configActionQueue  ConfigActionQueue
	remoteConfigSource string // assets, ext_pipeline_config
}

type ConfigAction struct {
	Name   int
	Asset  core.Asset
	ProcId uint64
}

type ConfigActionQueue chan ConfigAction

func NewCdfConfigObserver(extractorID string, cogClient *CdfClient, remoteConfigSource string) *CdfConfigObserver {
	configActionQueue := make(chan ConfigAction, 5)
	return &CdfConfigObserver{extractorID: extractorID, cogClient: cogClient, configActionQueue: configActionQueue, remoteConfigSource: remoteConfigSource}
}

// Start starts observer process using provided asset filter and reload interval. The operation is non-blocking
func (intgr *CdfConfigObserver) Start(assetFilter core.AssetFilter, reloadInterval time.Duration) ConfigActionQueue {
	log.Info("Starting CDF config observer, remote config source = ", intgr.remoteConfigSource)
	if reloadInterval == 0 {
		reloadInterval = 60 * time.Second
	}
	intgr.isStarted = true
	intgr.assetFilter = assetFilter
	go func() {
		for {
			err := intgr.reloadRemoteConfigs()
			if err != nil {
				log.Error("Failed to reload remote configs with error : ", err)
			}
			time.Sleep(reloadInterval)
			if !intgr.isStarted {
				break
			}
		}
		log.Info("CDF config loop has been terminated")
	}()
	return intgr.configActionQueue

}

func (intgr *CdfConfigObserver) Stop() {
	log.Info("Stopping CDF config observer")
	intgr.isStarted = false
}

// reloadRemoteConfigs loads relote configuration from CDF , compares with local state and run actions (start/stop/restart) to match local state to target state .
func (intgr *CdfConfigObserver) reloadRemoteConfigs() error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error(" CameraImagesToCdf failed to load configuration from CDF with error : ", stack)
		}
	}()

	log.Debug("Reloading remote config")

	var remoteAssetList core.AssetList
	var err error

	if intgr.remoteConfigSource == "assets" {
		remoteAssetList, err = intgr.cogClient.Client().Assets.Filter(intgr.assetFilter, 1000)
		if err != nil {
			return err
		}
	} else if intgr.remoteConfigSource == "ext_pipeline_config" {
		remoteConfig, err := intgr.cogClient.Client().ExtractionPipelines.GetRemoteConfig(intgr.extractorID)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(remoteConfig.Config), &remoteAssetList)
		if err != nil {
			return err
		}

	} else {
		log.Error("Unknown remote config source")
		return nil
	}

	// comparing existing assets with assets in cdf , reloading processor if there is a difference

	// 1. asset is not present in master list - new asset has been added in CDF. Action - start new processor
	// 2. asset is not present in slave list - asset has been deleted . Action - stop processor
	// 3. asset present in both lists but metadata is defferent - asset has been updated . Action reload processor
	// 4. assets are equal - Do nothing

	log.Debug("Number of assets from CDF =", len(remoteAssetList))

	var updatedRemoteAsset core.Asset
	var isUpdated bool

	for i := range intgr.localAssetsList {
		isEqual := false
		isPresentRemotely := false
		for i2 := range remoteAssetList {
			if intgr.localAssetsList[i].ID == remoteAssetList[i2].ID {
				isPresentRemotely = true
				if intgr.cogClient.CompareAssets(intgr.localAssetsList[i], remoteAssetList[i2]) {
					isEqual = true
				} else {
					updatedRemoteAsset = remoteAssetList[i2]
				}
				break
			}
		}

		if !isPresentRemotely {
			log.Infof("Remote change detected. Removing processor %d ", intgr.localAssetsList[i].ID)
			isUpdated = true
			intgr.configActionQueue <- ConfigAction{Name: StopProcessorAction, ProcId: intgr.localAssetsList[i].ID}

		} else if isPresentRemotely && !isEqual {
			log.Infof("Remote change detected. Restarting and updating processor %d", intgr.localAssetsList[i].ID)
			// Reload processor
			isUpdated = true
			intgr.configActionQueue <- ConfigAction{Name: RestartProcessorAction, Asset: updatedRemoteAsset}
		}
	}

	// Comparing remote assets with local , starting processors for all new remote assets
	for i3 := range remoteAssetList {
		isPresent := false
		for i4 := range intgr.localAssetsList {
			if intgr.localAssetsList[i4].ID == remoteAssetList[i3].ID {
				isPresent = true
				break
			}
		}
		if !isPresent {
			isUpdated = true
			log.Infof("Remote change detected. Starting new processor %d for camera %s", remoteAssetList[i3].ID, remoteAssetList[i3].Name)
			// The service starts separate processor for each camera.
			intgr.configActionQueue <- ConfigAction{Name: StartProcessorLoopAction, Asset: remoteAssetList[i3]}
		}

	}
	if isUpdated {
		intgr.localAssetsList = remoteAssetList
	}

	return nil
}
