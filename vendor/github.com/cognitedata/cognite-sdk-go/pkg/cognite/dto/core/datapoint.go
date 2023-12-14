package core

type Datapoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type DatapointList []Datapoint

func (datapoints *DatapointList) ConvertToCreateDatapoints(timeSerieID uint64) CreateDatapoints {
	createDatapointList := CreateDatapointsInTimeSerieList{
		CreateDatapointsInTimeSerie{
			Datapoints:  *datapoints,
			TimeSerieID: timeSerieID,
		},
	}
	return CreateDatapoints{
		Items: createDatapointList,
	}
}

func (datapoints *DatapointList) ConvertToCreateDatapointsExt(timeSerieIDExternal string) CreateDatapoints {
	createDatapointList := CreateDatapointsInTimeSerieList{
		CreateDatapointsInTimeSerie{
			Datapoints:          *datapoints,
			TimeSerieExternalID: timeSerieIDExternal,
		},
	}
	return CreateDatapoints{
		Items: createDatapointList,
	}
}

type DatapointsFilter struct {
	Items                DatapointsQuery `json:"items,omitempty"`
	Start                int64           `json:"start,omitempty"`
	End                  int64           `json:"end,omitempty"`
	Limit                uint32          `json:"limit,omitempty"`
	Aggregates           []string        `json:"aggregates,omitempty"`
	Granularity          string          `json:"granularity,omitempty"`
	IncludeOutsidePoints bool            `json:"includeOutsidePoints,omitempty"`
}

type DatapointsQuery struct {
	Start                int64    `json:"start,omitempty"`
	End                  int64    `json:"end,omitempty"`
	Limit                uint32   `json:"limit,omitempty"`
	Aggregates           []string `json:"aggregates,omitempty"`
	Granularity          string   `json:"granularity,omitempty"`
	IncludeOutsidePoints bool     `json:"includeOutsidePoints,omitempty"`
	TimeSerieID          uint64   `json:"id,omitempty"`
}

type DatapointListResponse struct {
	Items DatapointsListResponse `json:"items,omitempty"`
}

type DatapointsResponse struct {
	TimeSerieID         uint64        `json:"id,omitempty"`
	TimeSerieExternalID string        `json:"externalId,omitempty"`
	IsString            bool          `json:"isString,omitempty"`
	Datapoints          DatapointList `json:"datapoints,omitempty"`
}

type DatapointsListResponse []DatapointsResponse

type CreateDatapointsInTimeSerie struct {
	Datapoints          DatapointList `json:"datapoints,omitempty"`
	TimeSerieID         uint64        `json:"id,omitempty"`
	TimeSerieExternalID string        `json:"externalId,omitempty"`
}

type CreateDatapointsInTimeSerieList []CreateDatapointsInTimeSerie

type CreateDatapoints struct {
	Items CreateDatapointsInTimeSerieList `json:"items,omitempty"`
}

type RetrieveLatestDatapoint struct {
	Items []LatestDatapoint `json:"items,omitempty"`
}

type LatestDatapoint struct {
	Before              string `json:"before,omitempty"`
	TimeSerieID         uint64 `json:"id,omitempty"`
	TimeSerieExternalID string `json:"externalId,omitempty"`
}

type DeleteDatapoints struct {
	Items []DeleteDatapointsQuery `json:"items,omitempty"`
}

type DeleteDatapointsQuery struct {
	InclusiveBegin      int64  `json:"inclusiveBegin,omitempty"`
	ExclusiveEnd        int64  `json:"exclusiveEnd,omitempty"`
	TimeSerieID         uint64 `json:"id,omitempty"`
	TimeSerieExternalID string `json:"externalId,omitempty"`
}
