package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/api"
	dto "github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/pkg/errors"
)

// Raw is a manager that is used to query raw in CDF
type Raw struct {
	apiClient *api.Client
}

// NewRaw creates a Raw manager that is used to query raw in CDF
func NewRaw(apiClient *api.Client) *Raw {
	return &Raw{
		apiClient: apiClient,
	}
}

// List Raw databases in CDF
func (rawManager *Raw) List(cursor string, limit int) (*dto.DatabaseListResponse, error) {
	params := url.Values{}
	if cursor != "" {
		params["cursor"] = []string{cursor}
	}
	params["limit"] = []string{strconv.Itoa(limit)}
	body, err := rawManager.apiClient.GetWithParams("raw/dbs", params)
	if err != nil {
		return nil, err
	}
	var response = new(dto.DatabaseListResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in List() Raw")
	}
	return response, nil
}

// DeleteDatabases deletes raw databases in CDF
func (rawManager *Raw) DeleteDatabases(databases dto.DatabaseList, options ...func(*dto.DeleteDatabases)) error {
	deleteDatabases := databases.ConvertToDeleteDatabases(options...)
	jsonBytes, err := json.Marshal(deleteDatabases)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in DeleteDatabases() Raw")
	}
	_, err = rawManager.apiClient.Post("raw/dbs/delete", jsonBytes)
	if err != nil {
		return errors.Wrap(err, "Unable to delete databases")
	}
	return nil
}

// CreateTables create tables
func (rawManager *Raw) CreateTables(dbName string, tables dto.TableList) (dto.TableList, error) {
	createTables := tables.ConvertToCreateTable()
	jsonBytes, err := json.Marshal(createTables)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal struct in Insert() Raw")
	}
	path := fmt.Sprintf("raw/dbs/%s/tables", dbName)
	params := url.Values{"ensureParent": {"true"}}
	body, err := rawManager.apiClient.PostWithParams(path, jsonBytes, params)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create tables")
	}
	var response = new(dto.CreateTableResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal struct in CreateTables() Raw")
	}
	return response.Items, nil
}

// Insert rows
func (rawManager *Raw) Insert(dbName string, tableName string, rows dto.RowList) error {
	insertRows := rows.ConvertToInsertRows()
	jsonBytes, err := json.Marshal(insertRows)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct in Insert() Raw")
	}
	path := fmt.Sprintf("raw/dbs/%s/tables/%s/rows", dbName, tableName)
	params := url.Values{"ensureParent": {"true"}}
	_, err = rawManager.apiClient.PostWithParams(path, jsonBytes, params)
	return err
}
