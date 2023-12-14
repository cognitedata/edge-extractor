package iam

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/iam"
	"github.com/pkg/errors"
)

type Groups struct {
	apiClient *api.Client
}

func NewGroups(apiClient *api.Client) *Groups {
	return &Groups{
		apiClient: apiClient,
	}
}

// List groups
func (groupsManager *Groups) List(params url.Values) (dto.GroupList, error) {
	body, err := groupsManager.apiClient.GetWithParams("groups", params)
	if err != nil {
		return nil, err
	}
	var response = new(dto.GroupListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() Groups")
	}
	return response.Items, nil
}

// Create groups
func (groupsManager *Groups) Create(groups dto.GroupList) (dto.GroupList, error) {
	createGroups := groups.ConvertToCreateGroups()
	jsonBytes, err := json.Marshal(createGroups)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() Groups")
	}
	body, err := groupsManager.apiClient.Post("groups", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.GroupListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() Groups")
	}
	return response.Items, nil
}

// Delete groups
func (groupsManager *Groups) Delete(groups dto.GroupIDs) error {
	jsonBytes, err := json.Marshal(groups)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() Groups")
	}
	_, err = groupsManager.apiClient.Post("groups/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

// ListServiceAccounts lists service accounts in groups
func (groupsManager *Groups) ListServiceAccounts(groupID uint64) (dto.ServiceAccountList, error) {
	body, err := groupsManager.apiClient.Get(fmt.Sprintf("groups/%d/serviceaccounts", groupID))
	if err != nil {
		return nil, err
	}
	var response = new(dto.ServiceAccountListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in ListServiceAccounts() Groups")
	}
	return response.Items, nil
}

// AddServiceAccounts adds service accounts to groups
func (groupsManager *Groups) AddServiceAccounts(groupID uint64, serviceaccounts dto.ServiceAccountIDList) error {
	serviceAccountIDs := dto.ServiceAccountIDs{Items: serviceaccounts}
	jsonBytes, err := json.Marshal(serviceAccountIDs)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in AddServiceAccounts() Groups")
	}
	_, err = groupsManager.apiClient.Post(fmt.Sprintf("groups/%d/serviceaccounts", groupID), jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

// RemoveServiceAccounts removes service accounts from groups
func (groupsManager *Groups) RemoveServiceAccounts(groupID uint64, serviceaccounts dto.ServiceAccountIDList) error {
	serviceAccountIDs := dto.ServiceAccountIDs{Items: serviceaccounts}
	jsonBytes, err := json.Marshal(serviceAccountIDs)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in RemoveServiceAccounts() Groups")
	}
	_, err = groupsManager.apiClient.Post(fmt.Sprintf("groups/%d/serviceaccounts/remove", groupID), jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
