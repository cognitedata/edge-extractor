package iam

type APIKeyListResponse struct {
	Items      APIKeyList
	NextCursor string
}

type APIKey struct {
	ID               uint64 `json:"id"`
	ServiceAccountID uint64 `json:"serviceAccountId"`
	Status           string `json:"status"`
	CreatedTime      int64  `json:"createdTime"`
	Value            string `json:"value"`
}

type APIKeyList []APIKey

func (apiKeyList *APIKeyList) ConvertToIDList() APIKeyIDList {
	var apiKeyIDList APIKeyIDList
	for _, ak := range *apiKeyList {
		apiKeyIDList = append(apiKeyIDList, ak.ID)
	}
	return apiKeyIDList
}

func (apiKeyList *APIKeyList) ConvertToCreateAPIKeys() CreateAPIKeys {
	var createAPIKeyList CreateAPIKeyList
	for _, a := range *apiKeyList {
		createAPIKey := CreateAPIKey{
			ServiceAccountID: a.ServiceAccountID,
		}
		createAPIKeyList = append(createAPIKeyList, createAPIKey)
	}
	return CreateAPIKeys{
		Items: createAPIKeyList,
	}
}

func (apiKeyList *APIKeyList) ConvertToItemsList() APIKeyIDs {
	return APIKeyIDs{
		Items: apiKeyList.ConvertToIDList(),
	}
}

type CreateAPIKey struct {
	ServiceAccountID uint64 `json:"serviceAccountId"`
}

type CreateAPIKeyList []CreateAPIKey

type CreateAPIKeys struct {
	Items CreateAPIKeyList `json:"items"`
}

type APIKeyIDList []uint64

type APIKeyIDs struct {
	Items APIKeyIDList `json:"items"`
}
