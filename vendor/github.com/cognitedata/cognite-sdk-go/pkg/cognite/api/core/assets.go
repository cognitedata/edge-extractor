package core

import (
	"encoding/json"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/pkg/errors"
)

// Assets is a manager that is used to query
// assets in CDF
type Assets struct {
	apiClient *api.Client
}

// NewAssets creates a Assets manager that is used to query
// assets in CDF
func NewAssets(apiClient *api.Client) *Assets {
	return &Assets{
		apiClient: apiClient,
	}
}

// List assets
func (assetsManager *Assets) List() (dto.AssetList, error) {
	body, err := assetsManager.apiClient.Get("assets")
	if err != nil {
		return nil, err
	}
	var response = new(dto.AssetListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() Assets")
	}
	return response.Items, nil
}

// Filter assets
func (assetsManager *Assets) Filter(filter dto.AssetFilter, limit int) (dto.AssetList, error) {
	assetFilter := dto.AssetFilterWrapper{
		Filter: filter,
		Limit:  limit,
	}
	jsonBytes, err := json.Marshal(assetFilter)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Filter() Assets")
	}

	body, err := assetsManager.apiClient.Post("assets/list", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.AssetListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Filter() Assets")
	}
	return response.Items, nil
}

// Aggregate assets
func (assetsManager *Assets) Aggregate(filter dto.AssetFilter) (uint64, error) {
	assetFilter := dto.AssetFilterWrapper{
		Filter: filter,
	}
	jsonBytes, err := json.Marshal(assetFilter)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to marshal struct in Aggregate() Assets")
	}

	body, err := assetsManager.apiClient.Post("assets/aggregate", jsonBytes)
	if err != nil {
		return 0, err
	}
	var response = new(dto.AssetAggregateResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to unmarshal struct in Aggregate() Assets")
	}
	return response.Items[0].Count, nil
}

// Retrieve assets
func (assetsManager *Assets) Retrieve(assets dto.AssetIDList) (dto.AssetList, error) {
	retrieveAssets := dto.RetrieveAssets{
		Items: assets,
	}
	jsonBytes, err := json.Marshal(retrieveAssets)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Retrieve() Assets")
	}

	body, err := assetsManager.apiClient.Post("assets/byids", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.AssetListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Retrieve() Assets")
	}
	return response.Items, nil
}

// Create assets
func (assetsManager *Assets) Create(assets dto.AssetList) (dto.AssetList, error) {
	createAssets := assets.ConvertToCreateAssets()
	jsonBytes, err := json.Marshal(createAssets)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() Assets")
	}
	body, err := assetsManager.apiClient.Post("assets", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.AssetListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() Assets")
	}
	return response.Items, nil
}

// Update assets
func (assetsManager *Assets) Update(assets dto.AssetList) (dto.AssetList, error) {
	if len(assets) == 0 {
		return nil, errors.New("Cannot update with empty list")
	}
	updateAssets := assets.ConvertToUpdateAssets()
	jsonBytes, err := json.Marshal(updateAssets)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Update() Assets")
	}
	body, err := assetsManager.apiClient.Post("assets/update", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.AssetListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Update() Assets")
	}

	return response.Items, nil
}

// Delete assets
func (assetsManager *Assets) Delete(assets dto.AssetIDList, options ...func(*dto.DeleteAssets)) error {
	deleteAssets := assets.ConvertToDeleteAssetRequest(options...)
	jsonBytes, err := json.Marshal(deleteAssets)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() Assets")
	}
	_, err = assetsManager.apiClient.Post("assets/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

// RetrieveSubtree retrieves a asset subtree
func (assetsManager *Assets) RetrieveSubtree(rootID uint64, rootExternalID string) (dto.AssetList, error) {
	var filter dto.AssetFilter
	if rootExternalID == "" && rootID == 0 {
		return nil, errors.New("RootExternalID or RootID must be provided")
	}
	if rootID != 0 {
		filter = dto.AssetFilter{
			RootIDs: dto.AssetIDList{dto.AssetID{ID: rootID}},
		}
	}
	if rootExternalID != "" {
		filter = dto.AssetFilter{
			ExternalIDPrefix: rootExternalID,
		}
	}
	assetHierarchy, err := assetsManager.Filter(filter, 1000)
	if err != nil {
		return nil, err
	}
	return assetHierarchy, nil
}
