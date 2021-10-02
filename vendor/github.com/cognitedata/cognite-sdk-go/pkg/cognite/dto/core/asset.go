package core

type AssetList []Asset

type AssetListResponse struct {
	Items      AssetList
	NextCursor string
}

type AssetAggregateResponse struct {
	Items []struct {
		Count uint64
	}
}

type Asset struct {
	ID               uint64            `json:"id"`
	Name             string            `json:"name"`
	ExternalID       string            `json:"externalId"`
	ParentID         uint64            `json:"parentId"`
	ParentExternalID string            `json:"parentExternalID"`
	Description      string            `json:"description"`
	Metadata         map[string]string `json:"metadata"`
	Source           string            `json:"source"`
	CreatedTime      int64             `json:"createdTime"`
	LastUpdatedTime  int64             `json:"lastUpdatedTime"`
	RootID           uint64            `json:"rootId"`
}

type AssetID struct {
	ID         uint64 `json:"id,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}

type AssetIDList []AssetID

type RetrieveAssets struct {
	Items AssetIDList `json:"items"`
}

type CreateAsset struct {
	ExternalID       string            `json:"externalId,omitempty"`
	Name             string            `json:"name"`
	ParentID         uint64            `json:"parentId,omitempty"`
	Description      string            `json:"description,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Source           string            `json:"source,omitempty"`
	ParentExternalID string            `json:"parentExternalId,omitempty"`
}

type CreateAssetList []CreateAsset

type CreateAssets struct {
	Items CreateAssetList `json:"items"`
}

type DeleteAssets struct {
	Items            AssetIDList `json:"items"`
	Recursive        bool        `json:"recursive"`
	IgnoreUnknownIDs bool        `json:"ignoreUnknownIds"`
}

type UpdateAssetAttributes struct {
	ExternalID       *UpdateString           `json:"externalId,omitempty"`
	Name             *UpdateString           `json:"name,omitempty"`
	Description      *UpdateString           `json:"description,omitempty"`
	Metadata         *UpdateMap              `json:"metadata,omitempty"`
	Source           *UpdateString           `json:"source,omitempty"`
	ParentID         *UpdateParentId         `json:"parentId,omitempty"`
	ParentExternalID *UpdateParentExternalId `json:"parentExternalId,omitempty"`
}

type UpdateAsset struct {
	ID         uint64                 `json:"id,omitempty"`
	ExternalID string                 `json:"externalId,omitempty"`
	Update     *UpdateAssetAttributes `json:"update,omitempty"`
}

type UpdateAssetList []UpdateAsset

type UpdateAssets struct {
	Items UpdateAssetList `json:"items"`
}

type AssetFilterWrapper struct {
	Filter AssetFilter `json:"filter"`
	Limit  int         `json:"limit,omitempty"`
	Cursor string      `json:"cursor,omitempty"`
}

type AssetFilter struct {
	Name             string            `json:"name,omitempty"`
	ParentIDs        []uint64          `json:"parentIds,omitempty"`
	RootIDs          AssetIDList       `json:"rootIds,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Source           string            `json:"source,omitempty"`
	CreatedTime      int64             `json:"createdTime,omitempty"`
	LastUpdatedTime  int64             `json:"lastUpdatedTime,omitempty"`
	ExternalIDPrefix string            `json:"externalIdPrefix,omitempty"`
	Root             *bool             `json:"root,omitempty"`
}

func (ts *Asset) ConvertToUpdateAsset() UpdateAsset {
	return UpdateAsset{
		ID:         ts.ID,
		ExternalID: ts.ExternalID,
		Update: &UpdateAssetAttributes{
			ExternalID:       &UpdateString{Set: ts.ExternalID, SetNull: ts.ExternalID == ""},
			Name:             &UpdateString{Set: ts.Name, SetNull: ts.Name == ""},
			Description:      &UpdateString{Set: ts.Description, SetNull: ts.Description == ""},
			Metadata:         &UpdateMap{Set: ts.Metadata, SetNull: ts.Metadata == nil},
			Source:           &UpdateString{Set: ts.Source, SetNull: ts.Source == ""},
			ParentID:         SetUpdateParentId(ts.ParentID),
			ParentExternalID: SetUpdateParentExternalId(ts.ParentExternalID),
		},
	}
}

func (assetList *AssetList) ConvertToIDList() AssetIDList {
	var assetIDList AssetIDList
	for _, a := range *assetList {
		assetID := AssetID{ID: a.ID}
		assetIDList = append(assetIDList, assetID)
	}
	return assetIDList
}

func (assetList *AssetList) ConvertToExternalIDList() AssetIDList {
	var assetIDList AssetIDList
	for _, a := range *assetList {
		assetID := AssetID{ExternalID: a.ExternalID}
		assetIDList = append(assetIDList, assetID)
	}
	return assetIDList
}

func (assetList *AssetList) ConvertToCreateAssets() CreateAssets {
	var createAssetList CreateAssetList
	for _, a := range *assetList {
		createAsset := CreateAsset{
			ExternalID:       a.ExternalID,
			Name:             a.Name,
			ParentID:         a.ParentID,
			Description:      a.Description,
			Metadata:         a.Metadata,
			Source:           a.Source,
			ParentExternalID: a.ParentExternalID,
		}
		createAssetList = append(createAssetList, createAsset)
	}
	return CreateAssets{
		Items: createAssetList,
	}
}

func (assetList *AssetList) ConvertToUpdateAssets() UpdateAssets {
	var updateAssetList UpdateAssetList
	for _, ts := range *assetList {
		updateAsset := ts.ConvertToUpdateAsset()
		updateAssetList = append(updateAssetList, updateAsset)
	}
	return UpdateAssets{
		Items: updateAssetList,
	}
}

func DeleteRecursive(b bool) func(*DeleteAssets) {
	return func(deleteAssets *DeleteAssets) {
		deleteAssets.Recursive = b
	}
}

func DeleteIgnoreUnknownIDs(b bool) func(*DeleteAssets) {
	return func(deleteAssets *DeleteAssets) {
		deleteAssets.IgnoreUnknownIDs = b
	}
}

func (assets *AssetIDList) ConvertToDeleteAssetRequest(options ...func(*DeleteAssets)) DeleteAssets {
	deleteAssets := DeleteAssets{Items: *assets}
	for _, option := range options {
		option(&deleteAssets)
	}
	return deleteAssets
}
