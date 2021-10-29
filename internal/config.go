package internal

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
}
