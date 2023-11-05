package internal

import (
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
)

var Key = ""

type StaticConfig struct {
	ProjectName  string
	CdfCluster   string
	AdTenantId   string
	AuthTokenUrl string
	ClientID     string
	Secret       string
	Scopes       []string
	CdfDatasetID int
	ExtractorID  string

	EnabledIntegrations []string
	LogLevel            string
	LogDir              string

	LocalIntegrationConfig []core.Asset
	IsEncrypted            bool
}

func (config *StaticConfig) Encrypt() error {
	var err error
	config.Secret, err = EncryptString(Key, config.Secret)
	// iterate over LocalIntegrationConfig and encrypt all passwords
	for i, _ := range config.LocalIntegrationConfig {
		password, ok := config.LocalIntegrationConfig[i].Metadata["password"]
		if ok {
			config.LocalIntegrationConfig[i].Metadata["password"], err = EncryptString(Key, password)
			if err != nil {
				return err
			}
		}
	}
	config.IsEncrypted = true
	return err
}

func (config *StaticConfig) Decrypt() error {
	var err error
	if !config.IsEncrypted {
		return nil
	}
	config.Secret, err = DecryptString(Key, config.Secret)
	// iterate over LocalIntegrationConfig and decrypt all passwords
	for i, _ := range config.LocalIntegrationConfig {
		password, ok := config.LocalIntegrationConfig[i].Metadata["password"]
		if ok {
			config.LocalIntegrationConfig[i].Metadata["password"], err = DecryptString(Key, password)
			if err != nil {
				return err
			}
		}
	}
	config.IsEncrypted = false
	return err
}
