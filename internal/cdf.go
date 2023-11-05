package internal

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite"
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	log "github.com/sirupsen/logrus"
)

type CdfClient struct {
	client    *cognite.Client
	dataSetId int
}

func NewCdfClient(projectName, cdfCluster, clientID, clientSecret string, scopes []string, azureTenantId, tokenUrl string, datasetId int) *CdfClient {
	if tokenUrl == "" {
		tokenUrl = "https://login.microsoftonline.com/" + azureTenantId + "/oauth2/v2.0/token"
	}

	auth := api.NewOidcClientCredsAuthWithParams(tokenUrl, clientID, clientSecret, scopes, nil)

	config := cognite.Config{
		LogLevel:    log.GetLevel().String(),
		Project:     projectName,
		BaseUrl:     "https://" + cdfCluster + ".cognitedata.com",
		AppName:     "edge-extractor",
		CogniteAuth: auth,
	}

	client := cognite.NewClient(&config)

	cdf := CdfClient{client: client, dataSetId: datasetId}

	return &cdf

}

func (co *CdfClient) Client() *cognite.Client {
	return co.client
}

func (co *CdfClient) UploadFile(filePath, externalId, name, mimeType string, assetId uint64) error {

	fileMetadata := core.CreateFileMetadata{ExternalId: externalId, Name: name, MimeType: mimeType, DataSetId: co.dataSetId}
	if assetId != 0 {
		fileMetadata.AssetIds = []uint64{assetId}
	}

	uploadUrl, err := co.client.Files.Create(fileMetadata)
	if err != nil {
		log.Error("Upload error : ", err.Error())
		return err
	}
	log.Debug("Uploading file using URL:", uploadUrl)
	return co.BasicUploadFileBody(filePath, name, mimeType, uploadUrl.UploadUrl)
}

func (co *CdfClient) UploadInMemoryFile(body []byte, externalId, name, mimeType string, assetId uint64) error {

	fileMetadata := core.CreateFileMetadata{ExternalId: externalId, Name: name, MimeType: mimeType, DataSetId: co.dataSetId}
	if assetId != 0 {
		fileMetadata.AssetIds = []uint64{assetId}
	}

	uploadUrl, err := co.client.Files.Create(fileMetadata)
	if err != nil {
		log.Error("Upload error : ", err.Error())
		return err
	}
	log.Debug("Uploading file using URL:", uploadUrl)
	return co.UploadInMemoryBody(body, name, mimeType, uploadUrl.UploadUrl)
}

// UploadMultipartFileBody currently not supported by CDF
func (co *CdfClient) UploadMultipartFileBody(filePath, fileName, mimeType, uploadUrl string) error {
	log.Debug("Uploading file")
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	pReader, pWriter := io.Pipe()
	mWriter := multipart.NewWriter(pWriter)

	go func() {
		log.Debug("Copying content ")
		defer pWriter.Close()
		defer mWriter.Close()
		defer file.Close()

		part, err := mWriter.CreateFormFile("fileName", fileName)
		if err != nil {
			log.Error("Error form file ", err.Error())
			return
		}

		if _, err = io.Copy(part, file); err != nil {
			log.Error("Error pipe  ", err.Error())
			return
		}
		log.Debug("Copy is done")
	}()

	req, err := http.NewRequest("PUT", uploadUrl, pReader)
	req.Header.Set("Content-Type", mWriter.FormDataContentType())
	req.Header.Set("Content-Length", mWriter.FormDataContentType())

	if err != nil {
		return err
	}

	hClient := &http.Client{}
	log.Debug("Sending HTTP request")
	resp, err := hClient.Do(req)

	if err != nil {
		return err
	}
	log.Debug("Http response status code ", resp.Status)
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	log.Debug("Response :", string(body))
	if resp.Body != nil {
		resp.Body.Close()
	}

	return nil

}

func (co *CdfClient) BasicUploadFileBody(filePath, fileName, mimeType, uploadUrl string) error {
	log.Debug("Uploading file")
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	stat, _ := file.Stat()
	fileSize := stat.Size()

	log.Debug("File size :", fileSize)

	defer file.Close()

	req, err := http.NewRequest("PUT", uploadUrl, file)
	req.ContentLength = fileSize
	req.Header.Set("Content-Type", mimeType)
	if err != nil {
		return err
	}

	hClient := &http.Client{}
	resp, err := hClient.Do(req)

	if err != nil {
		return err
	}
	log.Debug("Http response status code ", resp.Status)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	log.Debug("Response :", string(body))
	if resp.Body != nil {
		resp.Body.Close()
	}

	return nil

}

func (co *CdfClient) UploadInMemoryBody(body []byte, fileName, mimeType, uploadUrl string) error {
	log.Debug("Uploading file")
	buf := bytes.NewReader(body)

	req, err := http.NewRequest("PUT", uploadUrl, buf)
	req.Header.Set("Content-Type", mimeType)
	if err != nil {
		return err
	}

	hClient := &http.Client{}
	resp, err := hClient.Do(req)

	if err != nil {
		return err
	}
	log.Debug("Http response status code ", resp.Status)

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	log.Debug("Response :", string(respBody))
	if resp.Body != nil {
		resp.Body.Close()
	}

	return nil

}

// CompareAssets compares 2 assets and returs true if they are equal
func (co *CdfClient) CompareAssets(asset1, asset2 core.Asset) bool {

	if asset1.ID != asset2.ID {
		return false
	}
	if asset1.ExternalID != asset2.ExternalID {
		return false
	}

	if asset1.Name != asset2.Name {
		return false
	}

	if len(asset1.Metadata) != len(asset2.Metadata) {
		return false
	}

	for i := range asset1.Metadata {
		if asset1.Metadata[i] != asset2.Metadata[i] {
			return false
		}
	}
	return true

}
