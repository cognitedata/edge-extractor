package iam

import (
	"encoding/json"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/iam"
	"github.com/pkg/errors"
)

type ServiceAccounts struct {
	apiClient *api.Client
}

func NewServiceAccounts(apiClient *api.Client) *ServiceAccounts {
	return &ServiceAccounts{
		apiClient: apiClient,
	}
}

// List serviceaccounts
func (serviceaccountsManager *ServiceAccounts) List() (dto.ServiceAccountList, error) {
	body, err := serviceaccountsManager.apiClient.Get("serviceaccounts")
	if err != nil {
		return nil, err
	}
	var response = new(dto.ServiceAccountListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() ServiceAccounts")
	}
	return response.Items, nil
}

// Create serviceaccounts
func (serviceaccountsManager *ServiceAccounts) Create(serviceaccounts dto.ServiceAccountList) (dto.ServiceAccountList, error) {
	createServiceAccounts := serviceaccounts.ConvertToCreateServiceAccounts()
	jsonBytes, err := json.Marshal(createServiceAccounts)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() ServiceAccounts")
	}
	body, err := serviceaccountsManager.apiClient.Post("serviceaccounts", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.ServiceAccountListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() ServiceAccounts")
	}
	return response.Items, nil
}

// Delete serviceaccounts
func (serviceaccountsManager *ServiceAccounts) Delete(serviceaccounts dto.ServiceAccountIDList) error {
	serviceAccountIDs := dto.ServiceAccountIDs{Items: serviceaccounts}
	jsonBytes, err := json.Marshal(serviceAccountIDs)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() ServiceAccounts")
	}
	_, err = serviceaccountsManager.apiClient.Post("serviceaccounts/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
