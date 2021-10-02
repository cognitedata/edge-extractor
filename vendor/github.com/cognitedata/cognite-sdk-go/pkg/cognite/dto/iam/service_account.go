package iam

type ServiceAccountListResponse struct {
	Items      ServiceAccountList
	NextCursor string
}

type ServiceAccount struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Groups      []uint64 `json:"groups"`
	IsDeleted   bool     `json:"isDeleted"`
	DeletedTime int64    `json:"deletedTime"`
}

type ServiceAccountList []ServiceAccount

func (serviceAccountList *ServiceAccountList) ConvertToIDList() ServiceAccountIDList {
	var serviceAccountIDList ServiceAccountIDList
	for _, s := range *serviceAccountList {
		serviceAccountIDList = append(serviceAccountIDList, s.ID)
	}
	return serviceAccountIDList
}

func (serviceAccountList *ServiceAccountList) ConvertToCreateServiceAccounts() CreateServiceAccounts {
	var createServiceAccountList CreateServiceAccountList
	for _, s := range *serviceAccountList {
		createServiceAccount := CreateServiceAccount{
			Name:   s.Name,
			Groups: s.Groups,
		}
		createServiceAccountList = append(createServiceAccountList, createServiceAccount)
	}
	return CreateServiceAccounts{
		Items: createServiceAccountList,
	}
}

func (serviceAccountList *ServiceAccountList) ConvertToItemsList() ServiceAccountIDs {
	return ServiceAccountIDs{
		Items: serviceAccountList.ConvertToIDList(),
	}
}

type CreateServiceAccount struct {
	Name   string   `json:"name"`
	Groups []uint64 `json:"groups"`
}

type CreateServiceAccountList []CreateServiceAccount

type CreateServiceAccounts struct {
	Items CreateServiceAccountList `json:"items"`
}

type ServiceAccountIDList []uint64
type ServiceAccountIDs struct {
	Items ServiceAccountIDList `json:"items"`
}
