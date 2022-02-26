package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type QueryRow map[string]interface{}

type QueryResult []QueryRow

type DbConnection interface {
	Connect(string) error
	Ping() (bool, error)
	IsConnected() bool
	Query(exp string, params ...interface{}) (QueryResult, error)
	Insert(table string, args []map[string]interface{}, returnId bool) (QueryResult, error)
	UpdateQuery(table string, id int, args map[string]interface{}) (QueryResult, error)
	DeleteQuery(table string, id ...int) (QueryResult, error)
	GetCopy() (DbConnection, error)
}

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func (conf *DatabaseConfig) GetDatasource() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", conf.User, conf.Password, conf.Database)
}

func (q *QueryResult) ReadRows(rows *sql.Rows) error {
	Columns, err := rows.Columns()
	if err != nil {
		return err
	}
	Count := len(Columns)
	valuePtrs := make([]interface{}, Count)
	values := make([]interface{}, Count)
	counter := 0
	for rows.Next() {
		mapArrayUnit := QueryRow{}
		for i := range Columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)
		for i, col := range Columns {
			val := values[i]
			mapArrayUnit[col] = val
		}
		*q = append(*q, mapArrayUnit)
		counter++
	}
	return nil
}

func (q *QueryResult) Length() int {
	return len(*q)
}

func (q *QueryResult) Bytes() []byte {
	data, err := json.Marshal(&q)
	if err != nil {
		return []byte(err.Error())
	}
	return data
}
