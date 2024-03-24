package core

import (
	"encoding/json"

	"github.com/cognitedata/edge-extractor/apps/lib"
	log "github.com/sirupsen/logrus"
)

type AppConfiguration struct {
	InstanceID     string
	AppName        string
	Configurations json.RawMessage
}

type AppManager struct {
	Apps         map[string]lib.AppInstance
	Integrations map[string]interface{}
}

func NewAppManager() *AppManager {
	return &AppManager{
		Apps:         make(map[string]lib.AppInstance),
		Integrations: make(map[string]interface{}),
	}
}

func (am *AppManager) SetIntegration(name string, integration interface{}) {
	am.Integrations[name] = integration
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
