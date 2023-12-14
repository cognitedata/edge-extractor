package core

import (
	"encoding/json"
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/pkg/errors"
)

// Files is a manager that is used to query files in CDF
type Files struct {
	apiClient *api.Client
}

// New creates a Files manager that is used to query files in CDF
func NewFiles(apiClient *api.Client) *Files {
	return &Files{
		apiClient: apiClient,
	}
}

type fileFilterRequest struct {
	Filter dto.FilesFilter `json:"filter"`
	Limit  int             `json:"limit,omitempty"`
	Cursor string          `json:"cursor,omitempty"`
}

type fileMetadataListResponse struct {
	Items  dto.FileMetadataList `json:"items"`
	Cursor string               `json:"cursor"`
}

type filesAggregateResponse struct {
	Items []dto.Aggregate
}

// List file metadata
func (files *Files) Filter(filter dto.FilesFilter, limit int) (dto.FileMetadataList, error) {
	assetFilter := fileFilterRequest{
		Filter: filter,
		Limit:  limit,
	}
	jsonBytes, err := json.Marshal(assetFilter)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in files.Filter()")
	}

	body, err := files.apiClient.Post("files/list", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(fileMetadataListResponse)
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in files.Filter()")
	}
	return response.Items, nil
}

// Create file metadata
func (files *Files) Create(fileMetadata dto.CreateFileMetadata) (dto.FileMetadataWithUploadUrl, error) {
	jsonBytes, err := json.Marshal(fileMetadata)
	if err != nil {
		return dto.FileMetadataWithUploadUrl{}, errors.Wrap(err, "Unable to marshal struct in files.Create()")
	}
	body, err := files.apiClient.Post("files", jsonBytes)
	if err != nil {
		return dto.FileMetadataWithUploadUrl{}, err
	}

	var response = dto.FileMetadataWithUploadUrl{}
	if err := json.Unmarshal(body, &response); err != nil {
		return dto.FileMetadataWithUploadUrl{}, errors.Wrap(err, "Unable to unmarshal struct in files.Create()")
	}
	return response, nil
}

// Retrieve file metadata
func (files *Files) Retrieve(ids dto.IdentifierList) (dto.FileMetadataList, error) {
	identifiers := dto.IdentifierItems{
		Items: ids,
	}
	jsonBytes, err := json.Marshal(identifiers)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in files.Retrieve()")
	}

	body, err := files.apiClient.Post("files/byids", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(fileMetadataListResponse)
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in files.Retrieve()")
	}
	return response.Items, nil
}

// Delete file metadata
func (files *Files) Delete(ids dto.IdentifierList) error {
	identifiers := dto.IdentifierItems{
		Items: ids,
	}
	jsonBytes, err := json.Marshal(identifiers)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in files.Delete()")
	}

	body, err := files.apiClient.Post("files/delete", jsonBytes)
	if err != nil {
		return err
	}
	var response = new(fileMetadataListResponse)
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, "Unable to unmarshal struct in files.Delete()")
	}
	return nil
}

// Aggregate file metadata
func (files *Files) Aggregate(filter dto.FilesFilter) (dto.Aggregate, error) {
	req := fileFilterRequest{Filter: filter}
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return dto.Aggregate{}, errors.Wrap(err, "Unable to marshal struct in files.Aggregate()")
	}

	body, err := files.apiClient.Post("files/aggregate", jsonBytes)
	if err != nil {
		return dto.Aggregate{}, err
	}

	var response = new(filesAggregateResponse)
	if err := json.Unmarshal(body, &response); err != nil {
		return dto.Aggregate{}, errors.Wrap(err, "Unable to unmarshal struct in files.Aggregate()")
	}
	return response.Items[0], nil
}
