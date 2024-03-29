package core

const (
	ExtractionRunStatusSuccess = "success"
	ExtractionRunStatusFailure = "failure"
	ExtractionRunStatusSeen    = "seen"
)

type CreateExtractionRun struct {
	ExternalID  string `json:"externalId"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	CreatedTime int64  `json:"createdTime"`
}

type CreateExtractonRunsList []CreateExtractionRun

type CreateExtractonRuns struct {
	Items CreateExtractonRunsList `json:"items"`
}

type ConfigRessponse struct {
    ExternalId   string `json:"externalId"`
    Config       string `json:"config"`
    Revision     int    `json:"revision"`
    CreatedTime  int64  `json:"createdTime"`
    Description  string `json:"description"`
}
