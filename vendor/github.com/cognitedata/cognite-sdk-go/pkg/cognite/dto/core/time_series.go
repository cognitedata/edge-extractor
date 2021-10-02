package core

type TimeSerieListResponse struct {
	Items      TimeSerieList
	NextCursor string
}

type TimeSerie struct {
	ID          uint64            `json:"id"`
	ExternalID  string            `json:"externalId"`
	Name        string            `json:"name"`
	IsString    bool              `json:"isString"`
	Metadata    map[string]string `json:"metadata"`
	Unit        string            `json:"unit"`
	AssetID     uint64            `json:"assetId"`
	IsStep      bool              `json:"isStep"`
	Description string            `json:"description"`
	//SecurityCategories []SecurityCategory `json:"securityCategories"`
	CreatedTime     int64 `json:"createdTime"`
	LastUpdatedTime int64 `json:"lastUpdatedTime"`
}

func (ts *TimeSerie) ConvertToCreateTimeSerie() CreateTimeSerie {
	return CreateTimeSerie{
		ExternalID:  ts.ExternalID,
		Name:        ts.Name,
		IsString:    ts.IsString,
		Metadata:    ts.Metadata,
		Unit:        ts.Unit,
		AssetID:     ts.AssetID,
		IsStep:      ts.IsStep,
		Description: ts.Description,
	}
}

func (ts *TimeSerie) ConvertToUpdateTimeSerie() UpdateTimeSerie {
	return UpdateTimeSerie{
		ID:         ts.ID,
		ExternalID: ts.ExternalID,
		Update: &UpdateTimeSerieAttributes{
			ExternalID:  &UpdateString{Set: ts.ExternalID, SetNull: ts.ExternalID == ""},
			Name:        &UpdateString{Set: ts.Name, SetNull: ts.Name == ""},
			Unit:        &UpdateString{Set: ts.Unit, SetNull: ts.Unit == ""},
			Metadata:    &UpdateMap{Set: ts.Metadata, SetNull: ts.Metadata == nil},
			AssetID:     &UpdateUint64{Set: ts.AssetID, SetNull: ts.AssetID == 0},
			Description: &UpdateString{Set: ts.Description, SetNull: ts.Description == ""},
		},
	}
}

type TimeSerieList []TimeSerie

func (timeSerieList *TimeSerieList) ConvertToIDList() TimeSerieIDList {
	var timeSerieIDList TimeSerieIDList
	for _, evt := range *timeSerieList {
		timeSerieID := TimeSerieID{ID: evt.ID}
		timeSerieIDList = append(timeSerieIDList, timeSerieID)
	}
	return timeSerieIDList
}

func (timeSerieList *TimeSerieList) ConvertToExternalIDList() TimeSerieIDList {
	var timeSerieIDList TimeSerieIDList
	for _, evt := range *timeSerieList {
		timeSerieID := TimeSerieID{ExternalID: evt.ExternalID}
		timeSerieIDList = append(timeSerieIDList, timeSerieID)
	}
	return timeSerieIDList
}

func (timeSerieList *TimeSerieList) ConvertToCreateTimeSeries() CreateTimeSeries {
	var createTimeSerieList CreateTimeSerieList
	for _, ts := range *timeSerieList {
		createTimeSerie := ts.ConvertToCreateTimeSerie()
		createTimeSerieList = append(createTimeSerieList, createTimeSerie)
	}
	return CreateTimeSeries{
		Items: createTimeSerieList,
	}
}

func (timeSerieList *TimeSerieList) ConvertToUpdateTimeSeries() UpdateTimeSeries {
	var updateTimeSerieList UpdateTimeSerieList
	for _, ts := range *timeSerieList {
		updateTimeSerie := ts.ConvertToUpdateTimeSerie()
		updateTimeSerieList = append(updateTimeSerieList, updateTimeSerie)
	}
	return UpdateTimeSeries{
		Items: updateTimeSerieList,
	}
}

// Split returns two time serie lists. First list is the time serie that exist in the given TimeSerieIDList
// The second is the rest
func (timeSerieList *TimeSerieList) Split(timeSerieIDList TimeSerieIDList) (TimeSerieList, TimeSerieList) {
	var inIDList TimeSerieList
	var notInIDList TimeSerieList
	for _, ts := range *timeSerieList {
		isInIDList := false
		for _, tsID := range timeSerieIDList {
			if (tsID.ExternalID != "" && tsID.ExternalID == ts.ExternalID) ||
				(tsID.ID != 0 && tsID.ID == ts.ID) {
				inIDList = append(inIDList, ts)
				isInIDList = true
			}
		}
		if !isInIDList {
			notInIDList = append(notInIDList, ts)
		}
	}
	return inIDList, notInIDList
}

func (timeSerieList *TimeSerieList) UniqueExternalId() TimeSerieList {
	externalIDs := make(map[string]bool)
	var uniqueList TimeSerieList
	for _, ts := range *timeSerieList {
		if _, value := externalIDs[ts.ExternalID]; !value {
			externalIDs[ts.ExternalID] = true
			uniqueList = append(uniqueList, ts)
		}
	}
	return uniqueList
}

type TimeSerieID struct {
	ID         uint64 `json:"id,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}

type TimeSerieIDList []TimeSerieID

type RetrieveTimeSeries struct {
	Items TimeSerieIDList `json:"items"`
}

type CreateTimeSerie struct {
	ExternalID  string            `json:"externalId,omitempty"`
	Name        string            `json:"name"`
	IsString    bool              `json:"isString,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	AssetID     uint64            `json:"assetId,omitempty"`
	IsStep      bool              `json:"isStep,omitempty"`
	Description string            `json:"description"`
	//SecurityCategories []SecurityCategory `json:"securityCategories"`
}

type CreateTimeSerieList []CreateTimeSerie

type CreateTimeSeries struct {
	Items CreateTimeSerieList `json:"items"`
}

type DeleteTimeSeries struct {
	Items TimeSerieIDList `json:"items"`
}

type TimeSerieSearchWrapper struct {
	Filter TimeSerieFilter `json:"filter,omitempty"`
	Search TimeSerieSearch `json:"search,omitempty"`
	Limit  int             `json:"limit,omitempty"`
}

type TimeSerieFilter struct {
	Name             string            `json:"name,omitempty"`
	Unit             string            `json:"unit,omitempty"`
	IsString         bool              `json:"isString,omitempty"`
	IsStep           bool              `json:"isStep,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	AssetIDList      []uint64          `json:"assetIdList,omitempty"`
	RootAssetIDs     []uint64          `json:"rootAssetIds,omitempty"`
	CreatedTime      int64             `json:"createdTime,omitempty"`
	LastUpdatedTime  int64             `json:"lastUpdatedTime,omitempty"`
	ExternalIDPrefix string            `json:"externalIdPrefix,omitempty"`
}

type TimeSerieSearch struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Query       string `json:"query,omitempty"`
}

type UpdateTimeSerieAttributes struct {
	ExternalID  *UpdateString `json:"externalId,omitempty"`
	Name        *UpdateString `json:"name,omitempty"`
	Unit        *UpdateString `json:"unit,omitempty"`
	Metadata    *UpdateMap    `json:"metadata,omitempty"`
	AssetID     *UpdateUint64 `json:"assetId,omitempty"`
	Description *UpdateString `json:"description,omitempty"`
}

type UpdateTimeSerie struct {
	ID         uint64                     `json:"id,omitempty"`
	ExternalID string                     `json:"externalId,omitempty"`
	Update     *UpdateTimeSerieAttributes `json:"update,omitempty"`
}

type UpdateTimeSerieList []UpdateTimeSerie

type UpdateTimeSeries struct {
	Items UpdateTimeSerieList `json:"items"`
}
