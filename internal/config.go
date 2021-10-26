package internal

type StaticConfig struct {
	ProjectName            string
	CdfCluster             string
	AdTenantId             string
	AuthTokenUrl           string
	ClientID               string
	Secret                 string
	Scopes                 []string
	CdfDatasetID           int
	ExtractionMonitoringID string

	EnabledIntegrations []string
	LogLevel            string
}
