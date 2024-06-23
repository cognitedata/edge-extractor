package core

import (
	"encoding/json"

	"github.com/cognitedata/edge-extractor/apps/lib"
	"github.com/cognitedata/edge-extractor/internal"
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
	ConfigObserver *internal.CdfConfigObserver
}

func NewAppManager(configObserver *internal.CdfConfigObserver) *AppManager {
	appManager := &AppManager{
		Apps:           make(map[string]lib.AppInstance),
		Integrations:   make(map[string]interface{}),
		ConfigObserver: configObserver,
	}
	return appManager
}

func (am *AppManager) SetIntegration(name string, integration interface{}) {
	am.Integrations[name] = integration
}

func (am *AppManager) StartConfigHandler() {
	log.Info("Starting processing loop using remote configurations")
	configQueue := am.ConfigObserver.SubscribeToAppsConfigUpdates()
	go func() {
		for configAction := range configQueue {
			log.Infof("Received new application config.Restarting apps")
			am.StopApps()
			err := am.LoadAppsFromRawConfig(configAction.Config)
			if err != nil {
				log.Errorf("Failed to load apps from config: %v", err)
				continue
			} else {
				am.StartApps()
				log.Info("Apps restarted")
			}
		}
	}()
}

func (am *AppManager) LoadAppsFromRawConfig(configs json.RawMessage) error {
	var appConfigs []AppConfiguration
	err := json.Unmarshal(configs, &appConfigs)
	if err != nil {
		return err
	}

	for _, appConfig := range appConfigs {
		appInstance := lib.NewAppInstance(appConfig.AppName)
		err := appInstance.ConfigureFromRaw(appConfig.Configurations)
		if err != nil {
			log.Errorf("Failed to configure app %s: %v", appConfig.AppName, err)
			continue
		}
		dependencies := appInstance.GetDependencies()
		for _, integrationName := range dependencies.Integrations {
			if integration, ok := am.Integrations[integrationName]; ok {
				appInstance.ConfigureIntegration(integration)
			} else {
				log.Errorf("Integration %s not configured for app %s", integrationName, appConfig.AppName)
			}
		}
		am.Apps[appConfig.InstanceID] = appInstance
		log.Infof("App %s loaded", appConfig.AppName)
	}

	return nil
}

func (am *AppManager) StartApps() {
	log.Info("Starting micro-apps")
	for _, app := range am.Apps {
		err := app.Start()
		if err != nil {
			log.Errorf("Failed to start micro-app: %v", err)
		}
	}
	log.Info("Micro-apps started")
}

func (am *AppManager) StopApps() {
	for _, app := range am.Apps {
		app.Stop()
	}
}
