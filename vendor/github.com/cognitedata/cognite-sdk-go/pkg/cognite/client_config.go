package cognite

import (
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BaseUrl     string
	Project     string
	AppName     string
	LogLevel    string
	CogniteAuth api.CogniteAuth
}

type baseEnvConfig struct {
	BASE_URL  string `required:"true"`
	PROJECT   string `required:"true"`
	LOG_LEVEL string `default:"INFO"`
}

type apiKeyEnvConfig struct {
	API_KEY string `required:"true"`
}

type oidcEnvConfig struct {
	OIDC_CLIENT_SECRET string   `required:"true"`
	OIDC_CLIENT_ID     string   `required:"true"`
	OIDC_SCOPES        []string `required:"true"`
	OIDC_TOKEN_URL     string   `required:"true"`
}

func processEnvConfig(cnf interface{}) error {
	if err := envconfig.Process("COGNITE", cnf); err != nil {
		return err
	}
	return nil
}

func NewConfigFromEnvironment(appName string, auth api.CogniteAuth) (*Config, error) {
	var cnf baseEnvConfig
	if err := processEnvConfig(&cnf); err != nil {
		return nil, err
	}
	return &Config{
		cnf.BASE_URL,
		cnf.PROJECT,
		appName,
		cnf.LOG_LEVEL,
		auth,
	}, nil
}

func NewConfigFromEnvironmentWithApiKeyAuth(appName string) (*Config, error) {
	var cnf apiKeyEnvConfig
	if err := processEnvConfig(&cnf); err != nil {
		return nil, err
	}
	return NewConfigFromEnvironment(appName, api.NewApiKeyAuth(cnf.API_KEY))
}

func NewConfigFromEnvironmentWithOidcAuth(appName string) (*Config, error) {
	var cnf oidcEnvConfig
	if err := processEnvConfig(&cnf); err != nil {
		return nil, err
	}
	auth := api.NewOidcClientCredsAuth(
		cnf.OIDC_TOKEN_URL,
		cnf.OIDC_CLIENT_ID,
		cnf.OIDC_CLIENT_SECRET,
		cnf.OIDC_SCOPES,
	)
	return NewConfigFromEnvironment(appName, auth)
}
