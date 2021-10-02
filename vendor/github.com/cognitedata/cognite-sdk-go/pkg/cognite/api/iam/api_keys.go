package iam

import (
	"encoding/json"
	"net/url"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/iam"
	"github.com/pkg/errors"
)

type APIKeys struct {
	apiClient *api.Client
}

func NewAPIKeys(apiClient *api.Client) *APIKeys {
	return &APIKeys{
		apiClient: apiClient,
	}
}

func (apikeysManager *APIKeys) ListAll() (dto.APIKeyList, error) {
	params := url.Values{}
	params.Add("all", "true")
	return apikeysManager.List(params)
}

// List apikeys
func (apikeysManager *APIKeys) List(params url.Values) (dto.APIKeyList, error) {
	body, err := apikeysManager.apiClient.GetWithParams("apikeys", params)
	if err != nil {
		return nil, err
	}
	var response = new(dto.APIKeyListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() APIKeys")
	}
	return response.Items, nil
}

// Create apikeys
func (apikeysManager *APIKeys) Create(apikeys dto.APIKeyList) (dto.APIKeyList, error) {
	createAPIKeys := apikeys.ConvertToCreateAPIKeys()
	jsonBytes, err := json.Marshal(createAPIKeys)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() APIKeys")
	}
	body, err := apikeysManager.apiClient.Post("apikeys", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.APIKeyListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() APIKeys")
	}
	return response.Items, nil
}

// Delete apikeys
func (apikeysManager *APIKeys) Delete(apikeys dto.APIKeyIDList) error {
	apikeyIDs := dto.APIKeyIDs{Items: apikeys}
	jsonBytes, err := json.Marshal(apikeyIDs)

	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() APIKeys")
	}
	_, err = apikeysManager.apiClient.Post("apikeys/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
