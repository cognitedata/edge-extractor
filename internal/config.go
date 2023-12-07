package internal

import (
	"encoding/json"
	"time"
)

var Key = ""

const ConfigSourceExtPipelines = "ext_pipeline_config"
const ConfigSourceLocal = "local"

type StaticConfig struct {
	ProjectName          string
	CdfCluster           string
	AdTenantId           string
	AuthTokenUrl         string
	ClientID             string
	Secret               string
	Scopes               []string
	CdfDatasetID         int
	ExtractorID          string
	RemoteConfigSource   string // local, ext_pipeline_config
	ConfigReloadInterval time.Duration
	EnabledIntegrations  []string
	LogLevel             string
	LogDir               string

	LocalIntegrationConfig map[string]json.RawMessage // map of integration configs (key is integration name, value is integration config)
	IsEncrypted            bool
	Secrets                map[string]string // map of encrypted secrets (key is secret name, value is encrypted secret)
}

func (config *StaticConfig) EncryptSecrets(key string) error {
	var err error
	config.Secret, err = EncryptString(key, config.Secret)
	for k, v := range config.Secrets {
		config.Secrets[k], err = EncryptString(key, v)
	}
	config.IsEncrypted = true
	return err
}

func (config *StaticConfig) DecryptSecrets(key string) error {
	var err error
	config.Secret, err = DecryptString(key, config.Secret)
	for k, v := range config.Secrets {
		config.Secrets[k], err = DecryptString(key, v)
	}
	config.IsEncrypted = false
	return err
}
