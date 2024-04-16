package internal

import (
	"encoding/json"
	"runtime/debug"
	"time"

	log "github.com/sirupsen/logrus"
)

// type StartProcessor func(asset core.Asset) error
// type RestartProcessor func(asset core.Asset)
// type StartProcessorLoop func(asset core.Asset) error
// type StopProcessor func(procId uint64)

type RemoteConfig map[string]json.RawMessage

// const StartProcessorAction = 1
const NewConfigAction = 1
const RestartProcessorAction = 2
const StopProcessorAction = 3
const StartProcessorLoopAction = 4

type CdfConfigObserver struct {
	extractorID            string
	isStarted              bool
	cogClient              *CdfClient
	remoteConfigSource     string // assets, ext_pipeline_config
	configUpdatesQueue     map[string]ConfigActionQueue
	appsConfigUpdatesQueue ConfigActionQueue
	secretManager          *SecretManager
	configRevision         int
}

type ConfigAction struct {
	Name     int
	Config   json.RawMessage
	ProcId   uint64
	Revision int
}

type ConfigActionQueue chan ConfigAction

func NewCdfConfigObserver(extractorID string, cogClient *CdfClient, remoteConfigSource string, secretManager *SecretManager) *CdfConfigObserver {
	return &CdfConfigObserver{extractorID: extractorID,
		cogClient:              cogClient,
		remoteConfigSource:     remoteConfigSource,
		configUpdatesQueue:     make(map[string]ConfigActionQueue),
		appsConfigUpdatesQueue: make(ConfigActionQueue),
		secretManager:          secretManager,
		configRevision:         -1,
	}
}

// Start starts observer process using provided asset filter and reload interval. The operation is non-blocking
func (intgr *CdfConfigObserver) Start(reloadInterval time.Duration) {
	log.Info("Starting CDF config observer, remote config source = ", intgr.remoteConfigSource)
	if reloadInterval == 0 {
		reloadInterval = 15 * time.Second
	}
	intgr.isStarted = true
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

}

// SubscribeToIntegrationConfigUpdates registers Integration in config observer and returns config action queue that Integration can use to receive config updates
// The queue has capacity of 5 items. If queue is full , the oldest item will be dropped. This is done to avoid blocking of config observer
// Config events aren't filtered and it's responsibility of Integration to do change detection
// name - name of Integration
// config - pointer to Integration config struct
func (intgr *CdfConfigObserver) SubscribeToIntegrationConfigUpdates(name string) ConfigActionQueue {
	intgr.configUpdatesQueue[name] = make(ConfigActionQueue, 5)
	return intgr.configUpdatesQueue[name]
}

func (intgr *CdfConfigObserver) SubscribeToAppsConfigUpdates() ConfigActionQueue {
	return intgr.appsConfigUpdatesQueue
}

func (intgr *CdfConfigObserver) Stop() {
	log.Info("Stopping CDF config observer")
	intgr.isStarted = false
}

// reloadRemoteConfigs loads config , sends config updates to all processors via config action queue
func (intgr *CdfConfigObserver) reloadRemoteConfigs() error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error(" CameraImagesToCdf failed to load configuration from CDF with error : ", stack)
		}
	}()

	log.Debug("Reloading remote config")

	if intgr.remoteConfigSource == "ext_pipeline_config" {
		remoteConfig, err := intgr.cogClient.Client().ExtractionPipelines.GetRemoteConfig(intgr.extractorID)
		if err != nil {
			return err
		}
		// Loading full static config from remote API. The config only expected to have integrations section and secrets section
		var remoteIntegrationsConfig StaticConfig
		err = json.Unmarshal([]byte(remoteConfig.Config), &remoteIntegrationsConfig)
		if err != nil {
			log.Error("Failed to unmarshal remote config with error : ", err)
			return err
		}

		if intgr.configRevision == remoteConfig.Revision {
			return nil
		} else {
			intgr.configRevision = remoteConfig.Revision
			log.Infof("New config revision has been loaded. Revision : %d", remoteConfig.Revision)
		}

		err = intgr.secretManager.LoadEncryptedSecrets(remoteIntegrationsConfig.Secrets)
		if err != nil {
			log.Error("Failed to load secrets with error : ", err)
		}

		for integrationNameFromRemote, rawConfig := range remoteIntegrationsConfig.Integrations {
			if queue, ok := intgr.configUpdatesQueue[integrationNameFromRemote]; ok {
				select {
				case queue <- ConfigAction{Name: NewConfigAction, Config: rawConfig, Revision: remoteConfig.Revision}:
				default:
					log.Warnf("Config action queue for processor %s is full", integrationNameFromRemote)
				}
			} else {
				log.Errorf("Processor %s is not registered in config registry", integrationNameFromRemote)
			}
		}

		if remoteIntegrationsConfig.Apps != nil {
			select {
			case intgr.appsConfigUpdatesQueue <- ConfigAction{Name: NewConfigAction, Config: remoteIntegrationsConfig.Apps, Revision: remoteConfig.Revision}:
			default:
				log.Warnf("Config action queue for app is full")
			}
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

	// intgr.configActionQueue <- ConfigAction{Name: RestartProcessorAction, Asset: updatedRemoteAsset}

	return nil
}
