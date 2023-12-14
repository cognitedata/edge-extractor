package iam

type GroupListResponse struct {
	Items      GroupList
	NextCursor string
}

type Group struct {
	ID           uint64       `json:"id"`
	SourceID     string       `json:"sourceId"`
	Capabilities Capabilities `json:"capabilities"`
	Name         string       `json:"name"`
	IsDeleted    bool         `json:"isDeleted"`
	DeletedTime  int64        `json:"deletedTime"`
}

type GroupList []Group

func (groupList *GroupList) ConvertToIDList() GroupIDList {
	var groupIDList []uint64
	for _, g := range *groupList {
		groupIDList = append(groupIDList, g.ID)
	}
	return groupIDList
}

func (groupList *GroupList) ConvertToCreateGroups() CreateGroups {
	var createGroupList CreateGroupList
	for _, g := range *groupList {
		createGroup := CreateGroup{
			Name:         g.Name,
			Capabilities: g.Capabilities,
			SourceID:     g.SourceID,
		}
		createGroupList = append(createGroupList, createGroup)
	}
	return CreateGroups{
		Items: createGroupList,
	}
}

func (groupList *GroupList) ConvertToDeleteList() GroupIDs {
	return GroupIDs{
		Items: groupList.ConvertToIDList(),
	}
}

type CreateGroup struct {
	SourceID     string       `json:"sourceId"`
	Capabilities Capabilities `json:"capabilities"`
	Name         string       `json:"name"`
}

type CreateGroupList []CreateGroup

type CreateGroups struct {
	Items CreateGroupList `json:"items"`
}

type GroupIDList []uint64

type GroupIDs struct {
	Items GroupIDList `json:"items"`
}

type Capability struct {
	GroupsACL             *ACL      `json:"groupsAcl,omitempty"`
	APIKeysACL            *ACL      `json:"apikeysAcl,omitempty"`
	SecurityCategoriesACL *ACL      `json:"securityCategoriesAcl,omitempty"`
	ServiceAccountsACL    *ACL      `json:"usersAcl,omitempty"`
	ProjectsACL           *ACL      `json:"projectsAcl,omitempty"`
	RawACL                *ACL      `json:"rawAcl,omitempty"`
	TimeSeriesACL         *ACL      `json:"timeSeriesAcl,omitempty"`
	FilesACL              *ACL      `json:"filesAcl,omitempty"`
	EventsACL             *ACL      `json:"eventsAcl,omitempty"`
	AssetsACL             *ACL      `json:"assetsAcl,omitempty"`
	DatasetsACL           *ACL      `json:"datasetsAcl,omitempty"`
	AllProjects           *struct{} `json:"allProjects"`
}

type Capabilities []Capability

type ACL struct {
	Actions []string `json:"actions,omitempty"`
	Scope   *Scope   `json:"scope,omitempty"`
}

type Scope struct {
	All              *struct{}         `json:"all"`
	CurrentUserScope map[string]string `json:"currentuserscope,omitempty"`
}
