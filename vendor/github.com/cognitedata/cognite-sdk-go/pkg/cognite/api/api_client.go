package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"time"

	dto_error "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client holds the HTTP client for the SDK
type Client struct {
	APIBaseURL  string
	AppName     string
	Auth        CogniteAuth
	client      *http.Client
	observers   []ClientObserver
	retryConfig *RetryConfig
}

// RetryConfig holds the retry configuration
type RetryConfig struct {
	MaxAttemptsPerRequest int
	InitialRetryDelay     time.Duration
	MaxRetryDelay         time.Duration
	BackoffMultiplier     float64
	RetryCodes            map[int]bool
}

// ClientObserver can be attached to a Client in order to inspect the requests/responses
type ClientObserver interface {
	OnCompletedRequest(req *http.Request, reqBody []byte, resp *http.Response, respBody []byte)
}

// ConfigTransport sets the transport for the api client
// Transport can be useful to add trusted proxy certificates.
func ConfigTransport(transport http.RoundTripper) func(*Client) {
	return func(apiClient *Client) {
		apiClient.client = &http.Client{Transport: transport}
	}
}

// ConfigRetry sets the retry configuration for the api client
func ConfigRetry(retryConfig RetryConfig) func(*Client) {
	return func(apiClient *Client) {
		apiClient.retryConfig = &retryConfig
	}
}

// NewClient is the constructor for Client
func NewClient(apiBaseURL string, appName string, auth CogniteAuth, options ...func(*Client)) *Client {
	apiClient := new(Client)
	apiClient.APIBaseURL = apiBaseURL
	apiClient.AppName = appName
	apiClient.Auth = auth
	apiClient.client = &http.Client{Transport: nil}
	apiClient.retryConfig = &RetryConfig{MaxAttemptsPerRequest: 1}

	for _, option := range options {
		option(apiClient)
	}

	return apiClient
}

// AddObserver adds the ClientObserver to the list of observers
// not thread safe
func (apiClient *Client) AddObserver(observer ClientObserver) {
	apiClient.observers = append(apiClient.observers, observer)
}

// SetHeaders sets the HTTP headers
func (apiClient *Client) SetHeaders(req *http.Request) {
	req.Header.Add("x-cdp-sdk", "go-sdk-v0.1")
	req.Header.Add("x-cdp-app", apiClient.AppName)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
}

// handleResponseCodes handles all response codes.
func handleResponseCodes(req *http.Request, resp *http.Response) (*dto_error.APIError, error) {
	if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
		return nil, nil
	} else if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error reading body with statuscode %d", resp.StatusCode))
		}
		var response = new(dto_error.APIErrorResponse)
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error unmarshal error body with statuscode %d", resp.StatusCode))
		}
		apiError := response.Error
		apiError.RequestID = resp.Header.Get("X-Request-Id")
		apiError.RequestURL = fmt.Sprintf("%s%s", req.Host, req.URL.Path)
		apiError.RequestMethod = req.Method
		logrus.Debugf("COGNITE-SDK: %s to %s resulted error code %d, '%s'. Request Id %s", apiError.RequestMethod, apiError.RequestURL, apiError.Code, apiError.Message, apiError.RequestID)
		return &apiError, nil
	} else {
		return nil, errors.New(resp.Status)
	}
}

func (apiClient *Client) makeRequest(httpType, path string, jsonBytes []byte, params url.Values) (*http.Request, error) {
	url := apiClient.APIBaseURL + path
	logrus.Debugf("COGNITE-SDK: %s %s", httpType, url)

	var req *http.Request
	var err error
	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
		logrus.Debugf("COGNITE-SDK: url = %s", url)
	}
	switch httpType {
	case "POST":
		logrus.Debugf("COGNITE-SDK: POST body %s", string(jsonBytes))
		req, err = http.NewRequest(httpType, url, bytes.NewBuffer(jsonBytes))
	case "PUT":
		logrus.Debugf("COGNITE-SDK: PUT body %s", string(jsonBytes))
		req, err = http.NewRequest(httpType, url, bytes.NewBuffer(jsonBytes))
	default:
		req, err = http.NewRequest(httpType, url, nil)
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error creating request")
	}
	apiClient.SetHeaders(req)
	apiClient.Auth.ConfigureAuth(req)

	return req, nil
}

func (apiClient *Client) handleRequest(httpType, path string, jsonBytes []byte, params url.Values) ([]byte, error) {

	var req *http.Request
	var resp *http.Response
	retriesLeft := apiClient.retryConfig.MaxAttemptsPerRequest
	retryDelay := apiClient.retryConfig.InitialRetryDelay

	for {
		var err error
		req, err = apiClient.makeRequest(httpType, path, jsonBytes, params)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error %s %s", httpType, path))
		}
		retriesLeft--
		resp, err = apiClient.client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error %s %s", httpType, path))
		}
		_, shouldRetryCode := apiClient.retryConfig.RetryCodes[resp.StatusCode]
		if !shouldRetryCode || retriesLeft <= 0 {
			break
		}
		logrus.Infof("Sleeping %v before retry", retryDelay)
		time.Sleep(retryDelay)
		retryDelay = time.Duration(math.Min(float64(retryDelay)*apiClient.retryConfig.BackoffMultiplier, float64(apiClient.retryConfig.MaxRetryDelay)))
	}

	apiError, err := handleResponseCodes(req, resp)
	if apiError != nil {
		return nil, apiError
	}
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error reading body for %s %s", httpType, path))
	}

	for _, o := range apiClient.observers {
		o.OnCompletedRequest(req, jsonBytes, resp, respBody)
	}

	return respBody, nil
}

// Get is a HTTP GET call that handles the http request as well as error handling of
// response codes.
func (apiClient *Client) Get(path string) ([]byte, error) {
	return apiClient.handleRequest("GET", path, nil, nil)
}

// GetWithParams is a HTTP GET call with params that handles the http request as well as error handling of
// response codes.
func (apiClient *Client) GetWithParams(path string, params url.Values) ([]byte, error) {
	return apiClient.handleRequest("GET", path, nil, params)
}

// Post is a HTTP POST call that handles the http request as well as error handling of
// response codes.
func (apiClient *Client) Post(path string, jsonBytes []byte) ([]byte, error) {
	return apiClient.handleRequest("POST", path, jsonBytes, nil)
}

// PostWithParams is a HTTP POST call with params that handles the http request as well as error handling of
// response codes.
func (apiClient *Client) PostWithParams(path string, jsonBytes []byte, params url.Values) ([]byte, error) {
	return apiClient.handleRequest("POST", path, jsonBytes, params)
}

// Put is a HTTP PUT call that handles the http request as well as error handling of
// response codes.
func (apiClient *Client) Put(path string, jsonBytes []byte) ([]byte, error) {
	return apiClient.handleRequest("PUT", path, jsonBytes, nil)
}
