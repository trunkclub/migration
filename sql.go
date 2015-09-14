package migration

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

type InsertStatment struct {
	columns []string
	queryString string
	db *sql.DB
}

func orderRecordValues(columns []string, values map[string]interface{}) []interface{} {
	vals := make([]interface{}, len(values))
	for i, column := range columns {
		vals[i] = values[column]
	}
	return vals
}

type Id int64
func (is InsertStatment) Execute(values map[string]interface{}) (Id, error) {
	vals := orderRecordValues(is.columns, values)
	var id int;
	err:= is.db.QueryRow(is.queryString, vals...).Scan(&id)

	if err != nil {
		return 0, err
	}

	return Id(id), nil
}

func generateParamString(paramCount int) string {
	params := make([]string, paramCount)
	for i := 0; i < paramCount; i++ {
		params[i] = fmt.Sprintf("$%v", i+1)
	}
	return strings.Join(params, ",")
}

func NewInsertStatement(db *sql.DB, table string, columns []string) InsertStatment {
	sqlString := fmt.Sprintf("INSERT INTO %v(%v) VALUES (%v) RETURNING id", table, strings.Join(columns, ","), generateParamString(len(columns)))
	
	return InsertStatment{columns: columns, queryString: sqlString, db: db}
}
