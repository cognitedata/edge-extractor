package core

type EventListResponse struct {
	Items      EventList
	NextCursor string
}

type EventAggregateResponse struct {
	Items []struct {
		Count uint64
	}
}

// Event objects store complex information about multiple assets over a time period.
// For example, an event can describe two hours of maintenance on a water pump and
// some associated pipes, or a future time window where the pump is scheduled for inspection.
// This is in contrast with data points in time series that store single pieces of
// information about one asset at specific points in time (e.g., temperature measurements).
type Event struct {
	ID              uint64            `json:"id"`
	ExternalID      string            `json:"externalId"`
	StartTime       int64             `json:"startTime"`
	EndTime         int64             `json:"endTime"`
	Type            string            `json:"type"`
	Subtype         string            `json:"subtype"`
	Description     string            `json:"description"`
	Metadata        map[string]string `json:"metadata"`
	AssetsIds       []uint64          `json:"assetsIds"`
	Source          string            `json:"source"`
	CreatedTime     int64             `json:"createdTime"`
	LastUpdatedTime int64             `json:"lastUpdatedTime"`
}

type EventList []Event

func (eventList *EventList) ConvertToIDList() EventIDList {
	var eventIDList EventIDList
	for _, evt := range *eventList {
		eventID := EventID{ID: evt.ID}
		eventIDList = append(eventIDList, eventID)
	}
	return eventIDList
}

func (eventList *EventList) ConvertToExternalIDList() EventIDList {
	var eventIDList EventIDList
	for _, evt := range *eventList {
		eventID := EventID{ExternalID: evt.ExternalID}
		eventIDList = append(eventIDList, eventID)
	}
	return eventIDList
}

func (eventList *EventList) ConvertToCreateEvents() CreateEvents {
	var createEventList CreateEventList
	for _, evt := range *eventList {
		createEvent := CreateEvent{
			ExternalID:  evt.ExternalID,
			Type:        evt.Type,
			Subtype:     evt.Subtype,
			Description: evt.Description,
			Metadata:    evt.Metadata,
			Source:      evt.Source,
		}
		createEventList = append(createEventList, createEvent)
	}
	return CreateEvents{
		Items: createEventList,
	}
}

type EventID struct {
	ID         uint64 `json:"id,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}

type EventIDList []EventID

type CreateEvent struct {
	ExternalID  string            `json:"externalId,omitempty"`
	StartTime   int64             `json:"startTime,omitempty"`
	EndTime     int64             `json:"endTime,omitempty"`
	Type        string            `json:"type"`
	Subtype     string            `json:"subtype,omitempty"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	AssetsIds   []uint64          `json:"assetsIds,omitempty"`
	Source      string            `json:"source,omitempty"`
}

type CreateEventList []CreateEvent

type CreateEvents struct {
	Items CreateEventList `json:"items"`
}

type RetrieveEvents struct {
	Items EventIDList `json:"items"`
}

type UpdateEvents struct {
	Items EventList `json:"items"`
}

type DeleteEvents struct {
	Items EventIDList `json:"items"`
}

type EventFilterRequest struct {
	Filter EventFilter `json:"filter"`
	Limit  int         `json:"limit,omitempty"`
	Cursor string      `json:"cursor,omitempty"`
}

type EventFilter struct {
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedTime      int64             `json:"createdTime,omitempty"`
	LastUpdatedTime  int64             `json:"lastUpdatedTime,omitempty"`
	ExternalIDPrefix string            `json:"externalIdPrefix,omitempty"`
}
