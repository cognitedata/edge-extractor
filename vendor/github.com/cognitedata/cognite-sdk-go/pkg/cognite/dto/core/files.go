package core

type FileMetadata struct {
	ExternalId         string            `json:"externalId"`
	Name               string            `json:"name"`
	Directory          string            `json:"directory"`
	Source             string            `json:"source"`
	MimeType           string            `json:"mimeType"`
	Metadata           map[string]string `json:"metadata"`
	AssetIds           []uint64          `json:"assetIds"`
	DataSetId          int               `json:"dataSetId"`
	SourceCreatedTime  int64             `json:"sourceCreatedTime"`
	SourceModifiedTime int64             `json:"sourceModifiedTime"`
	SecurityCategories []int             `json:"securityCategories,omitempty"`
	ID                 uint64            `json:"id"`
	Uploaded           bool              `json:"uploaded"`
	UploadedTime       int64             `json:"uploadedTime"`
	CreatedTime        int64             `json:"createdTime"`
	LastUpdatedTime    int64             `json:"lastUpdatedTime"`
}

type FileMetadataWithUploadUrl struct {
	FileMetadata
	UploadUrl string `json:"uploadUrl"`
}

type FileMetadataList []FileMetadata

type CreateFileMetadata struct {
	ExternalId         string            `json:"externalId,omitempty"`
	Name               string            `json:"name"`
	Directory          string            `json:"directory,omitempty"`
	Source             string            `json:"source,omitempty"`
	MimeType           string            `json:"mimeType,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	AssetIds           []uint64          `json:"assetIds"`
	DataSetId          int               `json:"dataSetId,omitempty"`
	SourceCreatedTime  int64             `json:"sourceCreatedTime,omitempty"`
	SourceModifiedTime int64             `json:"sourceModifiedTime,omitempty"`
	SecurityCategories []int             `json:"securityCategories,omitempty"`
}

type FilesFilter struct {
	Name               string            `json:"name,omitempty"`
	DirectoryPrefix    string            `json:"directoryPrefix,omitempty"`
	MimeType           string            `json:"mimeType,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	AssetIds           []uint64          `json:"assetIds,omitempty"`
	AssetExternalIds   []string          `json:"assetExternalIds,omitempty"`
	RootAssetIds       IdentifierList    `json:"rootAssetIds,omitempty"`
	AssetSubtreeIds    IdentifierList    `json:"assetSubtreeIds,omitempty"`
	Source             string            `json:"source,omitempty"`
	CreatedTime        int64             `json:"createdTime,omitempty"`
	LastUpdatedTime    int64             `json:"lastUpdatedTime,omitempty"`
	UploadedTime       int64             `json:"uploadedTime,omitempty"`
	SourceCreatedTime  int64             `json:"sourceCreatedTime,omitempty"`
	SourceModifiedTime int64             `json:"sourceModifiedTime,omitempty"`
	ExternalIdPrefix   string            `json:"externalIdPrefix,omitempty"`
	Uploaded           *bool             `json:"uploaded,omitempty"`
}

func (fileMetadata FileMetadata) ToCreateFileMetadata() CreateFileMetadata {
	createFileMetadata := CreateFileMetadata{
		ExternalId:         fileMetadata.ExternalId,
		Name:               fileMetadata.Name,
		Directory:          fileMetadata.Directory,
		Source:             fileMetadata.Source,
		MimeType:           fileMetadata.MimeType,
		Metadata:           fileMetadata.Metadata,
		AssetIds:           fileMetadata.AssetIds,
		DataSetId:          fileMetadata.DataSetId,
		SourceCreatedTime:  fileMetadata.SourceCreatedTime,
		SourceModifiedTime: fileMetadata.SourceModifiedTime,
		SecurityCategories: fileMetadata.SecurityCategories,
	}
	return createFileMetadata
}
