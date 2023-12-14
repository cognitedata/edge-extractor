package iam

import (
	"encoding/json"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/iam"
	"github.com/pkg/errors"
)

type SecurityCategories struct {
	apiClient *api.Client
}

func NewSecurityCategories(apiClient *api.Client) *SecurityCategories {
	return &SecurityCategories{
		apiClient: apiClient,
	}
}

// List securitycategories
func (securitycategoriesManager *SecurityCategories) List() (dto.SecurityCategoryList, error) {
	body, err := securitycategoriesManager.apiClient.Get("securitycategories")
	if err != nil {
		return nil, err
	}
	var response = new(dto.SecurityCategoryListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() SecurityCategories")
	}
	return response.Items, nil
}

// Create securitycategories
func (securitycategoriesManager *SecurityCategories) Create(securitycategories dto.SecurityCategoryList) (dto.SecurityCategoryList, error) {
	createSecurityCategories := securitycategories.ConvertToCreateSecurityCategories()
	jsonBytes, err := json.Marshal(createSecurityCategories)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() SecurityCategories")
	}
	body, err := securitycategoriesManager.apiClient.Post("securitycategories", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.SecurityCategoryListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() SecurityCategories")
	}
	return response.Items, nil
}

// Delete securitycategories
func (securitycategoriesManager *SecurityCategories) Delete(securitycategories dto.SecurityCategoryIDList) error {
	securityCategoryIDs := dto.SecurityCategoryIDs{Items: securitycategories}
	jsonBytes, err := json.Marshal(securityCategoryIDs)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() SecurityCategories")
	}
	_, err = securitycategoriesManager.apiClient.Post("securitycategories/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
