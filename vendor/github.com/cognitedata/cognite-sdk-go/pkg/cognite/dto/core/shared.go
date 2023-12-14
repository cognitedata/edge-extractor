package core

type Identifier struct {
	Id         uint64 `json:"id,omitempty"`
	ExternalId string `json:"externalId,omitempty"`
}

type IdentifierList []Identifier

type IdentifierItems struct {
	Items IdentifierList `json:"items"`
}

type Aggregate struct {
	Count int
}

func NewIdentifierListFromIds(ids ...uint64) IdentifierList {
	identifierList := IdentifierList{}
	for _, id := range ids {
		identifierList = append(identifierList, Identifier{Id: id})
	}
	return identifierList
}

func NewIdentifierListFromExternalIds(externalIds ...string) IdentifierList {
	identifierList := IdentifierList{}
	for _, id := range externalIds {
		identifierList = append(identifierList, Identifier{ExternalId: id})
	}
	return identifierList
}
