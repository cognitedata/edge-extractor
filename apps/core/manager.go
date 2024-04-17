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

func (am *AppManager) startConfigHandler() {
	log.Info("Starting processing loop using remote configurations")
	configQueue := am.ConfigObserver.SubscribeToAppsConfigUpdates()
	go func() {
		for configAction := range configQueue {
			log.Infof("Received new application config.Restarting apps")
			am.StopApps()
			am.LoadAppsFromRawConfig(configAction.Config)
			log.Info("Apps restarted")
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

func (am *AppManager) StartApps() {
	go am.startConfigHandler()
}

func (am *AppManager) StopApps() {
	for _, app := range am.Apps {
		app.Stop()
	}
}
