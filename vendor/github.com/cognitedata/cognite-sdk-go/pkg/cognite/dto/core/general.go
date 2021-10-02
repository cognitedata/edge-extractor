package core

type UpdateString struct {
	Set     string `json:"set,omitempty"`
	SetNull bool   `json:"setNull,omitempty"`
}
type UpdateUint64 struct {
	Set     uint64 `json:"set,omitempty"`
	SetNull bool   `json:"setNull,omitempty"`
}

type UpdateParentId struct {
	Set uint64 `json:"set,omitempty"`
}

func SetUpdateParentId(parentId uint64) *UpdateParentId {
	if parentId == 0 {
		return nil
	}
	return &UpdateParentId{
		Set: parentId,
	}
}

type UpdateParentExternalId struct {
	Set string `json:"set,omitempty"`
}

func SetUpdateParentExternalId(parentExternalId string) *UpdateParentExternalId {
	if parentExternalId == "" {
		return nil
	}
	return &UpdateParentExternalId{
		Set: parentExternalId,
	}
}

type UpdateMap struct {
	Set     map[string]string `json:"set,omitempty"`
	SetNull bool              `json:"setNull,omitempty"`
}

type UpdateArrayUint64 struct {
	Add    []uint64 `json:"add,omitempty"`
	Remove []uint64 `json:"remove,omitempty"`
	Set    []uint64 `json:"set,omitempty"`
}

type UpdateArrayString struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
	Set    []string `json:"set,omitempty"`
}
