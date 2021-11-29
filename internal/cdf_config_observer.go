package internal

import (
	"runtime/debug"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	log "github.com/sirupsen/logrus"
)

type StartProcessor func(asset core.Asset) error
type RestartProcessor func(asset core.Asset)
type StartProcessorLoop func(asset core.Asset) error
type StopProcessor func(procId uint64)

type CdfConfigObserver struct {
	extractorID        string
	isStarted          bool
	cogClient          *CdfClient
	localAssetsList    core.AssetList
	restartProcessor   RestartProcessor
	startProcessorLoop StartProcessorLoop
	stopProcessor      StopProcessor
}

func NewCdfConfigObserver(extractorID string, cogClient *CdfClient, restartProcessor RestartProcessor, startProcessorLoop StartProcessorLoop, stopProcessor StopProcessor) *CdfConfigObserver {
	return &CdfConfigObserver{extractorID: extractorID, cogClient: cogClient, restartProcessor: restartProcessor, startProcessorLoop: startProcessorLoop, stopProcessor: stopProcessor}
}

func (intgr *CdfConfigObserver) StartCdfConfigPolling() {
	for {
		intgr.ReloadRemoteConfigs()
		time.Sleep(time.Second * 60)
		if !intgr.isStarted {
			break
		}
	}
}

// LoadConfigs load both static and dynamic configs
func (intgr *CdfConfigObserver) ReloadRemoteConfigs() error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error(" CameraImagesToCdf failed to load configuration from CDF with error : ", stack)
		}
	}()

	log.Debug("Reloading remote config")
	filter := core.AssetFilter{Metadata: map[string]string{"cog_class": "camera", "state": "enabled", "extractor_id": intgr.extractorID}}

	remoteAssetList, err := intgr.cogClient.Client().Assets.Filter(filter, 1000)
	if err != nil {
		return err
	}
	// comparing existing assets with assets in cdf , reloading processor if there is a difference

	// 1. asset is not present in master list - new asset has been added in CDF. Action - start new processor
	// 2. asset is not present in slave list - asset has been deleted . Action - stop processor
	// 3. asset present in both lists but metadata is defferent - asset has been updated . Action reload processor
	// 4. assets are equal - Do nothing

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
			go intgr.stopProcessor(intgr.localAssetsList[i].ID)

		} else if isPresentRemotely && !isEqual {
			log.Infof("Remote change detected. Restarting and updating processor %d", intgr.localAssetsList[i].ID)
			// Reload processor
			isUpdated = true
			go intgr.restartProcessor(updatedRemoteAsset)
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
			go intgr.startProcessorLoop(remoteAssetList[i3])
		}

	}
	if isUpdated {
		intgr.localAssetsList = remoteAssetList
	}

	return nil
}
