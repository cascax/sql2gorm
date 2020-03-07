package parser

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func GetCreateTableFromDB(dsn, tableName string) (string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", errors.WithMessage(err, "open db error")
	}
	defer db.Close()
	rows, err := db.Query("SHOW CREATE TABLE " + tableName)
	if err != nil {
		return "", errors.WithMessage(err, "query show create table error")
	}
	defer rows.Close()
	if !rows.Next() {
		return "", errors.Errorf("table(%s) not found", tableName)
	}
	var table string
	var createSql string
	err = rows.Scan(&table, &createSql)
	if err != nil {
		return "", err
	}
	return createSql, nil
}

func ParseSqlFromDB(dsn, tableName string, options ...Option) (ModelCodes, error) {
	createSql, err := GetCreateTableFromDB(dsn, tableName)
	if err != nil {
		return ModelCodes{}, err
	}
	return ParseSql(createSql, options...)
}
