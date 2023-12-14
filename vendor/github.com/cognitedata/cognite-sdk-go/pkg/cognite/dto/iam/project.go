package iam

type Project struct {
	Name           string          `json:"name"`
	URLName        string          `json:"urlName"`
	DefaultGroupID uint64          `json:"defaultGroupId,omitempty"`
	Authentication *Authentication `json:"authentication,omitempty"`
}

type CreateProject struct {
	Name           string          `json:"name"`
	URLName        string          `json:"urlName"`
	Authentication *Authentication `json:"authentication"`
}

type Authentication struct {
	Protocol             string                `json:"protocol"`
	ValidDomains         []string              `json:"validDomains,omitempty"`
	ApplicationDomains   []string              `json:"applicationDomains,omitempty"`
	AzureADConfiguration *AzureADConfiguration `json:"azureADConfiguration,omitempty"`
	OAuth2Configuration  *OAuth2Configuration  `json:"oAuth2Configuration,omitempty"`
}

type AzureADConfiguration struct {
	AppID         string `json:"appId,omitempty"`
	AppSecret     string `json:"appSecret,omitempty"`
	TenantID      string `json:"tenantId,omitempty"`
	AppResourceID string `json:"appResourceId,omitempty"`
}

type OAuth2Configuration struct {
	LoginURL     string `json:"loginUrl,omitempty"`
	LogoutURL    string `json:"logoutUrl,omitempty"`
	TokenURL     string `json:"tokenUrl,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

func (project *Project) ConvertToCreateProjectItems() CreateProjectItems {
	createProject := CreateProject{
		Name:           project.Name,
		URLName:        project.URLName,
		Authentication: project.Authentication,
	}
	createProjectItems := CreateProjectItems{
		Items: []CreateProject{createProject},
	}
	return createProjectItems
}

type CreateProjectItems struct {
	Items []CreateProject `json:"items"`
}

type ProjectCreation struct {
	Project  Project `json:"project"`
	AdminKey string  `json:"adminKey"`
}

type ProjectCreations struct {
	Items []ProjectCreation `json:"items"`
}

type ProjectCreationResponse struct {
	Data ProjectCreations `json:"data"`
}
