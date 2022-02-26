package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Badgain/platform/config"
	"github.com/Badgain/platform/database/models"
	"github.com/Badgain/platform/logger"

	_ "github.com/lib/pq"
)

type PgConnection struct {
	conn      *sql.DB
	connected bool
	logger    *logger.Logger
	config    string
}

func (dbc *PgConnection) Connect(databaseConfig string) error {
	dbc.config = databaseConfig
	datasource := config.GlobalConfig.GetDatabaseConfig(databaseConfig)
	if datasource == nil {
		dbError := errors.New(fmt.Sprintf("no %s db config in global config file", databaseConfig))
		dbc.logger.Log([]byte(dbError.Error()), databaseConfig)
		return dbError
	}
	conn, err := sql.Open("postgres", datasource.GetDatasource())
	if err != nil {
		return err
	}
	dbc.conn = conn
	err = dbc.conn.Ping()
	if err != nil {
		dbc.logger.Log([]byte(err.Error()), databaseConfig)
		dbc.connected = false
		return err
	}
	dbc.connected = true
	return nil
}

func (dbc *PgConnection) Query(exp string, params ...interface{}) (models.QueryResult, error) {
	rows, err := dbc.conn.Query(exp, params...)
	if err != nil {
		dbc.logger.Log([]byte(err.Error()), dbc.config)
		return nil, err
	}
	result := models.QueryResult{}
	err = result.ReadRows(rows)
	if err != nil {
		dbc.logger.Log([]byte(err.Error()), dbc.config)
		return nil, err
	}
	return result, nil
}

func (dbc *PgConnection) Ping() (bool, error) {
	err := dbc.conn.Ping()
	if err != nil {
		dbc.connected = false
		return false, err
	}
	dbc.connected = true
	return true, nil
}

func (dbc *PgConnection) IsConnected() bool {
	return dbc.connected
}

func (dbc *PgConnection) Insert(table string, args []map[string]interface{}, returnId bool) (models.QueryResult, error) {
	var attrValues []interface{}
	var valuesArr []string

	result := models.QueryResult{}
	values := ""
	attrs := "("
	counter := 1
	isFirst := true

	for idx, unit := range args {
		isFirst = true
		values = ""
		for key, value := range unit {
			if idx == 0 {
				attrs += key + ", "
			}
			if !isFirst {
				values += fmt.Sprintf(", $%d", counter)
			} else {
				isFirst = false
				values += fmt.Sprintf("ssss,$%d", counter)
			}
			attrValues = append(attrValues, value)
			counter++
		}
		values = "(" + values[5:] + "),"
		valuesArr = append(valuesArr, values)

	}

	attrs = attrs[:len(attrs)-2] + ")"
	valuesArr[len(valuesArr)-1] = valuesArr[len(valuesArr)-1][:len(valuesArr[len(valuesArr)-1])-1]
	exp := fmt.Sprintf("insert into %s %s values %s", table, attrs, strings.Join(valuesArr, ""))
	if returnId {
		exp += " returning id"
	}
	rows, err := dbc.conn.Query(exp, attrValues...)
	if err != nil {
		return nil, err
	}

	err = result.ReadRows(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (dbc *PgConnection) UpdateQuery(table string, id int, args map[string]interface{}) (models.QueryResult, error) {
	result := models.QueryResult{}
	updateParamsStr := ""
	counter := 1
	var values []interface{}
	for key, value := range args {
		updateParamsStr += fmt.Sprintf("%s = $%d,", key, counter)
		values = append(values, value)
		counter++
	}
	updateParamsStr = updateParamsStr[:len(updateParamsStr)-1]
	exp := fmt.Sprintf("update %s set %s where id = %d returning id", table, updateParamsStr, id)
	rows, err := dbc.conn.Query(exp, values...)
	if err != nil {
		return nil, err
	}
	result.ReadRows(rows)
	return result, nil
}

func (dbc *PgConnection) DeleteQuery(table string, id ...int) (models.QueryResult, error) {
	result := models.QueryResult{}
	ids := ""
	for _, val := range id {
		ids += fmt.Sprintf("%d,", val)
	}
	ids = "(" + ids[:len(ids)-1] + ")"
	exp := fmt.Sprintf("delete from %s where id in %s returning id", table, ids)
	rows, err := dbc.conn.Query(exp)
	if err != nil {
		return nil, err
	}
	result.ReadRows(rows)
	return result, nil
}

func (dbc *PgConnection) GetCopy() (models.DbConnection, error) {
	cp := &PgConnection{}
	err := cp.Connect(dbc.config)
	if err != nil {
		return nil, err
	}
	return cp, nil
}
