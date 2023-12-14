package dto_error

import "fmt"

type APIErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code          int                      `json:"code"`
	Message       string                   `json:"message"`
	Missing       []map[string]interface{} `json:"missing"`
	Duplicated    []map[string]string      `json:"duplicated"`
	RequestID     string
	RequestURL    string
	RequestMethod string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API returned error '%s' with error code %d on url %s. Request ID: %s", e.Message, e.Code, e.RequestURL, e.RequestID)
}
