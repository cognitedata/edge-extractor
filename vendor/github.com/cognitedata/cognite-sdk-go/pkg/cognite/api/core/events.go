package core

import (
	"encoding/json"
	"fmt"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/pkg/errors"
)

// Events is a manager that is used to query
// events in CDF
type Events struct {
	apiClient *api.Client
}

// NewEvents creates a new Events manager
func NewEvents(apiClient *api.Client) *Events {
	return &Events{
		apiClient: apiClient,
	}
}

// Aggregate events
func (eventsManager *Events) Aggregate(filter dto.EventFilter) (uint64, error) {
	eventFilter := dto.EventFilterRequest{
		Filter: filter,
	}
	jsonBytes, err := json.Marshal(eventFilter)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to marshal struct in Aggregate() Events")
	}

	body, err := eventsManager.apiClient.Post("events/aggregate", jsonBytes)
	if err != nil {
		return 0, err
	}
	var response = new(dto.EventAggregateResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to unmarshal struct in Aggregate() Events")
	}
	return response.Items[0].Count, nil
}

// Filter events
func (eventsManager *Events) Filter(filter dto.EventFilter, cursor string, limit int) (*dto.EventListResponse, error) {
	eventFilter := dto.EventFilterRequest{
		Filter: filter,
		Cursor: cursor,
		Limit:  limit,
	}
	jsonBytes, err := json.Marshal(eventFilter)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Filter() Events")
	}
	body, err := eventsManager.apiClient.Post("events/list", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.EventListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Filter() Events")
	}
	return response, nil
}

// Create creates multiple event objects.
// It is possible to post a maximum of 1000 events per request.
func (eventsManager *Events) Create(events dto.EventList) (dto.EventList, error) {
	createAssets := events.ConvertToCreateEvents()
	jsonBytes, err := json.Marshal(createAssets)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() Events")
	}
	body, err := eventsManager.apiClient.Post("events", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.EventListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() Events")
	}
	return response.Items, nil
}

// RetrieveByID retrieves an event given an ID
func (eventsManager *Events) RetrieveByID(ID uint64) (*dto.Event, error) {
	body, err := eventsManager.apiClient.Get(fmt.Sprintf("events/%d", ID))
	if err != nil {
		return nil, err
	}
	var response = new(dto.Event)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in RetrieveByID() Events")
	}
	return response, nil
}

// Retrieve retrieves an events
func (eventsManager *Events) Retrieve(events dto.EventIDList) (dto.EventList, error) {
	retrieveEvents := dto.RetrieveEvents{
		Items: events,
	}
	jsonBytes, err := json.Marshal(retrieveEvents)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Retrieve() Events")
	}

	body, err := eventsManager.apiClient.Post("events/byids", jsonBytes)
	if err != nil {
		return nil, err
	}
	var response = new(dto.EventListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Retrieve() Events")
	}
	return response.Items, nil
}

// Delete events
func (eventsManager *Events) Delete(events dto.EventIDList) error {
	deleteEvents := dto.DeleteEvents{
		Items: events,
	}
	jsonBytes, err := json.Marshal(deleteEvents)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Delete() Events")
	}
	_, err = eventsManager.apiClient.Post("events/delete", jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
