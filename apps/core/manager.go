package core

import (
	"encoding/json"

	"github.com/cognitedata/edge-extractor/apps/lib"
	"github.com/cognitedata/edge-extractor/internal"
	"github.com/cskr/pubsub/v2"
	log "github.com/sirupsen/logrus"
)

type AppConfiguration struct {
	InstanceID     string
	AppName        string
	Configurations json.RawMessage
}

type AppManager struct {
	Apps           map[string]lib.AppInstance
	Integrations   map[string]interface{}
	systemEventBus *pubsub.PubSub[string, internal.SystemEvent]
}

func NewAppManager(systemEventBus *pubsub.PubSub[string, internal.SystemEvent]) *AppManager {
	appManager := &AppManager{
		Apps:           make(map[string]lib.AppInstance),
		Integrations:   make(map[string]interface{}),
		systemEventBus: systemEventBus,
	}
	if systemEventBus != nil {
		go appManager.runSystemEventsHandler()
	}
	return appManager
}

func (am *AppManager) SetIntegration(name string, integration interface{}) {
	am.Integrations[name] = integration
}

func (am *AppManager) runSystemEventsHandler() {
	for event := range am.systemEventBus.Sub("system/configs_updated") {
		log.Infof("New system event received: %v", event.EventType)
		newConfig, ok := event.Payload.(json.RawMessage)
		if !ok {
			log.Errorf("Failed to convert payload to json.RawMessage")
			continue
		}
		am.StopApps()
		err := am.LoadAppsFromRawConfig(newConfig)
		if err != nil {
			log.Errorf("Failed to load apps from raw config: %v", err)
			continue
		}
		log.Infof("Apps reloaded successfully")
	}
}

func (am *AppManager) LoadAppsFromRawConfig(configs json.RawMessage) error {
	var appConfigs []AppConfiguration
	err := json.Unmarshal(configs, &appConfigs)
	if err != nil {
		return err
	}

	for _, appConfig := range appConfigs {
		appInstane := lib.NewAppInstance(appConfig.AppName)
		err := appInstane.ConfigureFromRaw(appConfig.Configurations)
		if err != nil {
			log.Errorf("Failed to configure app %s: %v", appConfig.AppName, err)
			continue
		}
		dependencies := appInstane.GetDependencies()
		for _, integrationName := range dependencies.Integrations {
			if integration, ok := am.Integrations[integrationName]; ok {
				appInstane.ConfigureIntegration(integration)
			} else {
				log.Errorf("Integration %s not configured for app %s", integrationName, appConfig.AppName)
			}
		}
		err = appInstane.Start()
		if err != nil {
			log.Errorf("Failed to start app %s: %v", appConfig.AppName, err)
			continue
		}
		am.Apps[appConfig.InstanceID] = appInstane
		log.Infof("App %s started", appConfig.AppName)
	}

	return nil
}

func (am *AppManager) StopApps() {
	for _, app := range am.Apps {
		app.Stop()
	}
}
