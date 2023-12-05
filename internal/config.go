package internal

var Key = ""

type IntegrationConfigs map[string]interface{}

type StaticConfig struct {
	ProjectName        string
	CdfCluster         string
	AdTenantId         string
	AuthTokenUrl       string
	ClientID           string
	Secret             string
	Scopes             []string
	CdfDatasetID       int
	ExtractorID        string
	RemoteConfigSource string

	EnabledIntegrations []string
	LogLevel            string
	LogDir              string

	LocalIntegrationConfig IntegrationConfigs
	IsEncrypted            bool
	Secrets                map[string]string // map of encrypted secrets
}

func (config *StaticConfig) EncryptSecrets() error {
	var err error
	config.Secret, err = EncryptString(Key, config.Secret)
	for k, v := range config.Secrets {
		config.Secrets[k], err = EncryptString(Key, v)
	}
	config.IsEncrypted = true
	return err
}

// func (config *StaticConfig) Decrypt() error {
// 	var err error
// 	if !config.IsEncrypted {
// 		return nil
// 	}
// 	config.Secret, err = DecryptString(Key, config.Secret)
// 	config.IsEncrypted = false
// 	return err
// }
