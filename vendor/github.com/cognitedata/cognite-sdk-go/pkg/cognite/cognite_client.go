package cognite

import (
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api/core"
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api/iam"
	"github.com/sirupsen/logrus"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
)

// Client represents a Client that is used to interact with
// Cognite Data Fusion (CDF) api
type Client struct {
	Config              *Config
	apiClient           *api.Client
	Assets              *core.Assets
	Events              *core.Events
	Files               *core.Files
	TimeSeries          *core.TimeSeries
	ExtractionPipelines *core.ExtractionPipelines
	APIKeys             *iam.APIKeys
	ServiceAccounts     *iam.ServiceAccounts
	SecurityCategories  *iam.SecurityCategories
	Groups              *iam.Groups
	Projects            *iam.Projects
	Raw                 *core.Raw
}

func setLogLevel(lvl string) {
	// LOG_LEVEL not set, let's default to debug
	if lvl == "" {
		lvl = "info"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	// set global log level
	logrus.SetLevel(ll)
}

// NewClient return a Cognite Client that is used to interact with
// Cognite Data Fusion (CDF) api.
func NewClient(config *Config, options ...func(*api.Client)) *Client {
	setLogLevel(config.LogLevel)

	apiBasePath := config.BaseUrl + "/api/v1/projects/" + config.Project + "/"
	apiClient := api.NewClient(apiBasePath, config.AppName, config.CogniteAuth, options...)

	projectAPIBasePath := config.BaseUrl + "/api/v1/projects"
	projectAPIClient := api.NewClient(projectAPIBasePath, config.AppName, config.CogniteAuth, options...)

	client := Client{
		Config:              config,
		apiClient:           apiClient,
		Assets:              core.NewAssets(apiClient),
		Events:              core.NewEvents(apiClient),
		Files:               core.NewFiles(apiClient),
		TimeSeries:          core.NewTimeSeries(apiClient),
		ExtractionPipelines: core.NewExtractionPipelines(apiClient),
		APIKeys:             iam.NewAPIKeys(apiClient),
		Groups:              iam.NewGroups(apiClient),
		SecurityCategories:  iam.NewSecurityCategories(apiClient),
		ServiceAccounts:     iam.NewServiceAccounts(apiClient),
		Projects:            iam.NewProjects(projectAPIClient),
		Raw:                 core.NewRaw(apiClient),
	}
	return &client
}

// SetProjectName updates the project name and the URL in every APIClient
func (client *Client) SetProjectName(projectName string) {
	logrus.Debugf("COGNITE-SDK: Setting project name from '%s' to '%s'", client.Config.Project, projectName)
	client.Config.Project = projectName
	client.apiClient.APIBaseURL = client.Config.BaseUrl + "/api/v1/projects/" + projectName + "/"
}

// AddObserver adds the api.ClientObserver to the list of api observers
func (client *Client) AddObserver(observer api.ClientObserver) {
	client.apiClient.AddObserver(observer)
}
