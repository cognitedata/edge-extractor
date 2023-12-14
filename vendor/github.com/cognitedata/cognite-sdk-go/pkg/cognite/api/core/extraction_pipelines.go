package core

import (
	"encoding/json"
	"net/url"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/pkg/errors"
)

// ExtractionPipelines
type ExtractionPipelines struct {
	apiClient *api.Client
}

// New creates a ExtractionPipelines manager that is used to query ExtractionPipelines in CDF
func NewExtractionPipelines(apiClient *api.Client) *ExtractionPipelines {
	return &ExtractionPipelines{
		apiClient: apiClient,
	}
}

// CreateExtractionRuns registers extraction run in CDF
// https://api.cognitedata.com/api/v1/projects/{project}/extpipes/runs
func (extp *ExtractionPipelines) CreateExtractionRuns(extractionRunsList core.CreateExtractonRunsList) error {

	extractionRuns := core.CreateExtractonRuns{Items: extractionRunsList}

	jsonBytes, err := json.Marshal(extractionRuns)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in CreateExtractionRuns()")
	}
	body, err := extp.apiClient.Post("extpipes/runs", jsonBytes)
	if err != nil {
		return err
	}

	var response = new(dto.AssetListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal struct in Create() Assets")
	}
	return nil
}

func (extp *ExtractionPipelines) GetRemoteConfig(extractorExternalId string) (*core.ConfigRessponse, error) {
	params :=  url.Values{}
	params.Add("externalId", extractorExternalId)
	body, err := extp.apiClient.GetWithParams("extpipes/config" , params)
	var response = new(core.ConfigRessponse)
	if err != nil {
		return response, err
	}
	
	
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, errors.Wrap(err, "Unable to unmarshal struct in GetRemoteConfig()")
	}
	return response, nil
}