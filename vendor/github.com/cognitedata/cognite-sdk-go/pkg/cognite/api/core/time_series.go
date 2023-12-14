package core

import (
	"encoding/json"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto_error "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// TimeSeries is a manager that is used to query
// timeseries and datapoints in CDF
type TimeSeries struct {
	apiClient *api.Client
}

// NewTimeSeries creates a TimeSerie manager that is used to query
// timeseries and datapoints in CDF
func NewTimeSeries(apiClient *api.Client) *TimeSeries {
	return &TimeSeries{
		apiClient: apiClient,
	}
}

// List timeseries
func (timeSeriesManager *TimeSeries) List() (dto.TimeSerieList, error) {
	body, err := timeSeriesManager.apiClient.Get("timeseries")
	if err != nil {
		return nil, err
	}

	var response = new(dto.TimeSerieListResponse)
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() TimeSeries")
	}

	return response.Items, nil
}

// Create timeseries
func (timeSeriesManager *TimeSeries) Create(timeseries dto.TimeSerieList) (dto.TimeSerieList, error) {
	if len(timeseries) == 0 {
		return nil, errors.New("Cannot create with empty list")
	}
	createTimeSeries := timeseries.ConvertToCreateTimeSeries()
	jsonBytes, err := json.Marshal(createTimeSeries)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() TimeSeries")
	}
	body, err := timeSeriesManager.apiClient.Post("timeseries", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.TimeSerieListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() TimeSeries")
	}

	return response.Items, nil
}

// Update timeseries
func (timeSeriesManager *TimeSeries) Update(timeseries dto.TimeSerieList) (dto.TimeSerieList, error) {
	if len(timeseries) == 0 {
		return nil, errors.New("Cannot update with empty list")
	}
	updateTimeSeries := timeseries.ConvertToUpdateTimeSeries()
	jsonBytes, err := json.Marshal(updateTimeSeries)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Update() TimeSeries")
	}
	body, err := timeSeriesManager.apiClient.Post("timeseries/update", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.TimeSerieListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Update() TimeSeries")
	}

	return response.Items, nil
}

// Retrieve timeseries
func (timeSeriesManager *TimeSeries) Retrieve(timeseries dto.TimeSerieIDList) (dto.TimeSerieList, error) {
	retrieveTimeSeries := dto.RetrieveTimeSeries{
		Items: timeseries,
	}
	jsonBytes, err := json.Marshal(retrieveTimeSeries)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in retrieve TimeSeries")
	}

	body, err := timeSeriesManager.apiClient.Post("timeseries/byids", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.TimeSerieListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Retrieve() TimeSeries")
	}
	return response.Items, nil
}

// Search timeseries
func (timeSeriesManager *TimeSeries) Search(filter dto.TimeSerieFilter, search dto.TimeSerieSearch, limit int) (dto.TimeSerieList, error) {
	searchWrapper := dto.TimeSerieSearchWrapper{
		Filter: filter,
		Search: search,
		Limit:  limit,
	}
	jsonBytes, err := json.Marshal(searchWrapper)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Search TimeSeries")
	}

	body, err := timeSeriesManager.apiClient.Post("timeseries/search", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.TimeSerieListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Search() TimeSeries")
	}
	return response.Items, nil
}

// Delete timeseries
func (timeSeriesManager *TimeSeries) Delete(timeSeries dto.TimeSerieIDList) error {
	deleteTimeSeries := dto.DeleteTimeSeries{
		Items: timeSeries,
	}
	jsonBytes, err := json.Marshal(deleteTimeSeries)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() TimeSeries")
	}
	_, err = timeSeriesManager.apiClient.Post("timeseries/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

// CreateOrRetrieve creates all time series in the list. If it already exists, it will be retrieved.
// Returns the merged result
func (timeSeriesManager *TimeSeries) CreateOrRetrieve(timeSeries dto.TimeSerieList) (dto.TimeSerieList, error) {
	var toCreate dto.TimeSerieList
	var toRetrieve dto.TimeSerieList
	var created dto.TimeSerieList
	var retrieved dto.TimeSerieList
	uniqueTimeSeries := timeSeries.UniqueExternalId()
	retrieved, err := timeSeriesManager.Retrieve(uniqueTimeSeries.ConvertToExternalIDList())
	if err != nil {
		if e, ok := err.(*dto_error.APIError); ok && e.Code == 400 {
			idsToCreate := e.Missing
			var timeSerieIDList dto.TimeSerieIDList
			_ = mapstructure.Decode(idsToCreate, &timeSerieIDList)
			toCreate, toRetrieve = uniqueTimeSeries.Split(timeSerieIDList)
		} else {
			return nil, err
		}
	} else {
		return retrieved, nil
	}

	//create timeSeriesList
	if len(toCreate) > 0 {
		created, err = timeSeriesManager.Create(toCreate)
		if err != nil {
			return nil, err
		}
	}
	if len(toRetrieve) > 0 {
		retrieved, err = timeSeriesManager.Retrieve(toRetrieve.ConvertToExternalIDList())
		if err != nil {
			return nil, err
		}
	}
	return append(retrieved, created...), nil
}

// CreateOrUpdate creates all time series in the list. If it already exists, it will be updated.
// Returns the merged result
func (timeSeriesManager *TimeSeries) CreateOrUpdate(timeSeries dto.TimeSerieList) (dto.TimeSerieList, error) {
	var toCreate dto.TimeSerieList
	var toUpdate dto.TimeSerieList
	var created dto.TimeSerieList
	var updated dto.TimeSerieList
	uniqueTimeSeries := timeSeries.UniqueExternalId()

	timeSeriesList := uniqueTimeSeries.ConvertToExternalIDList()
	_, err := timeSeriesManager.Retrieve(timeSeriesList)
	if err != nil {
		if e, ok := err.(*dto_error.APIError); ok && e.Code == 400 {
			idsToCreate := e.Missing
			var timeSerieIDList dto.TimeSerieIDList
			_ = mapstructure.Decode(idsToCreate, &timeSerieIDList)
			toCreate, toUpdate = uniqueTimeSeries.Split(timeSerieIDList)
		} else {
			return nil, err
		}
	} else {
		// every time serie exist, hence update all
		toUpdate = uniqueTimeSeries
	}
	//create timeSeriesList
	if len(toCreate) > 0 {
		created, err = timeSeriesManager.Create(toCreate)
		if err != nil {
			return nil, err
		}
	}
	//update timeSeriesList
	if len(toUpdate) > 0 {
		updated, err = timeSeriesManager.Update(toUpdate)
		if err != nil {
			return nil, err
		}
	}
	return append(updated, created...), nil
}

// InsertDatapoints inserts datapoints
func (timeSeriesManager *TimeSeries) InsertDatapoints(datapoints dto.DatapointList, timeSerieID uint64) error {
	createDatapoints := datapoints.ConvertToCreateDatapoints(timeSerieID)
	jsonBytes, err := json.Marshal(createDatapoints)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in InsertDatapoints() TimeSeries")
	}

	_, err = timeSeriesManager.apiClient.Post("timeseries/data", jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

// InsertDatapoints inserts datapoints
func (timeSeriesManager *TimeSeries) InsertDatapointsExt(datapoints dto.DatapointList, timeSerieIDExt string) error {
	createDatapoints := datapoints.ConvertToCreateDatapointsExt(timeSerieIDExt)
	jsonBytes, err := json.Marshal(createDatapoints)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in InsertDatapoints() TimeSeries")
	}

	_, err = timeSeriesManager.apiClient.Post("timeseries/data", jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveDatapoints retrieves datapoints
func (timeSeriesManager *TimeSeries) RetrieveDatapoints(datapointsFilter dto.DatapointsFilter) (dto.DatapointList, error) {
	jsonBytes, err := json.Marshal(datapointsFilter)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in InsertDatapoints() TimeSeries")
	}

	body, err := timeSeriesManager.apiClient.Post("timeseries/data/list", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.DatapointListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in InsertDatapoints() TimeSeries")
	}
	return response.Items[0].Datapoints, nil
}

// RetrieveLatestDatapoint retrieves latest datapoint
func (timeSeriesManager *TimeSeries) RetrieveLatestDatapoint(timeSerieID uint64) (dto.DatapointList, error) {
	retrieveLatestDatapoint := dto.RetrieveLatestDatapoint{
		Items: []dto.LatestDatapoint{
			{
				Before:      "now",
				TimeSerieID: timeSerieID,
			},
		},
	}
	jsonBytes, err := json.Marshal(retrieveLatestDatapoint)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in RetrieveLatestDatapoint() TimeSeries")
	}

	body, err := timeSeriesManager.apiClient.Post("timeseries/data/latest", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.DatapointListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in RetrieveLatestDatapoint() TimeSeries")
	}
	return response.Items[0].Datapoints, nil
}

// DeleteDatapoints deletes datapoints
func (timeSeriesManager *TimeSeries) DeleteDatapoints(timeSerieID uint64, inclusiveBegin int64, exclusiveEnd int64) error {
	deleteDatapoints := dto.DeleteDatapoints{
		Items: []dto.DeleteDatapointsQuery{
			{
				InclusiveBegin: inclusiveBegin,
				ExclusiveEnd:   exclusiveEnd,
				TimeSerieID:    timeSerieID,
			},
		},
	}
	jsonBytes, err := json.Marshal(deleteDatapoints)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in DeleteDatapoints() TimeSeries")
	}

	_, err = timeSeriesManager.apiClient.Post("timeseries/data/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
