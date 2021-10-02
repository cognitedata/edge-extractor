package iam

import (
	"encoding/json"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/iam"
	"github.com/pkg/errors"
)

type Projects struct {
	apiClient *api.Client
}

func NewProjects(apiClient *api.Client) *Projects {
	return &Projects{
		apiClient: apiClient,
	}
}

func (projectsManager *Projects) Retrieve(projectName string) (*dto.Project, error) {
	body, err := projectsManager.apiClient.Get("/" + projectName)
	if err != nil {
		return nil, err
	}

	var response = new(dto.Project)
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Retrieve() Projects")
	}

	return response, nil
}

func (projectsManager *Projects) Update(projectName string, project *dto.Project) (*dto.Project, error) {
	jsonBytes, err := json.Marshal(project)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Update() Projects")
	}
	body, err := projectsManager.apiClient.Put("/"+projectName, jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.Project)
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Update() Projects")
	}

	return response, nil
}

// Create projects is only accessible to admins for creating new CDF tenants
func (projectsManager *Projects) Create(project *dto.Project) (*dto.ProjectCreation, error) {
	createProjectItems := project.ConvertToCreateProjectItems()

	jsonBytes, err := json.Marshal(createProjectItems)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() Projects")
	}
	body, err := projectsManager.apiClient.Post("", jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.ProjectCreationResponse)
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() Projects")
	}

	return &response.Data.Items[0], nil
}

// For adding a query at the end of a create POST request. Used by TF provider to specify no auto generating default groups and admin accounts
func (projectsManager *Projects) CreateWithQuery(project *dto.Project, query string) (*dto.ProjectCreation, error) {
	createProjectItems := project.ConvertToCreateProjectItems()

	jsonBytes, err := json.Marshal(createProjectItems)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Create() Projects")
	}

	body, err := projectsManager.apiClient.Post("?"+query, jsonBytes)
	if err != nil {
		return nil, err
	}

	var response = new(dto.ProjectCreationResponse)
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in Create() Projects")
	}

	return &response.Data.Items[0], nil
}
