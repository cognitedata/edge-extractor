package core

type Database struct {
	Name string `json:"name"`
}

type DatabaseList []Database

type Table struct {
	Name string `json:"name"`
}

type TableList []Table

type CreateTable struct {
	Items TableList `json:"items"`
}

type DatabaseListResponse struct {
	Items      DatabaseList
	NextCursor string
}

type CreateTableResponse struct {
	Items TableList `json:"items"`
}

type Row struct {
	Key             string                 `json:"key"`
	Columns         map[string]interface{} `json:"columns"`
	LastUpdatedTime int64                  `json:"lastUpdatedTime"`
}

type RowList []Row

type RowKey string

type RowKeyList []RowKey

type RetrieveRows struct {
	Items RowKeyList `json:"items"`
}

type RowListResponse struct {
	Items      RowList
	NextCursor string
}

type InsertRows struct {
	Items RowList `json:"items"`
}

type DeleteDatabases struct {
	Items     DatabaseList `json:"items"`
	Recursive bool         `json:"recursive"`
}

func (tableList *TableList) ConvertToCreateTable() CreateTable {
	return CreateTable{
		Items: *tableList,
	}
}

func (rowList *RowList) ConvertToRowKeyList() RowKeyList {
	var rowKeyList RowKeyList
	for _, row := range *rowList {
		key := row.Key
		rowKeyList = append(rowKeyList, RowKey(key))
	}
	return rowKeyList
}

func (rowList *RowList) ConvertToInsertRows() InsertRows {
	return InsertRows{
		Items: *rowList,
	}
}

func (databaseList *DatabaseList) ConvertToDeleteDatabases(options ...func(*DeleteDatabases)) DeleteDatabases {
	deleteDatabases := DeleteDatabases{Items: *databaseList}
	for _, option := range options {
		option(&deleteDatabases)
	}
	return deleteDatabases
}

func DeleteDatabaseRecursive(b bool) func(*DeleteDatabases) {
	return func(deleteDatabases *DeleteDatabases) {
		deleteDatabases.Recursive = b
	}
}
