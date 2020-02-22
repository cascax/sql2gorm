package parser

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseSql(t *testing.T) {
	sql := `CREATE TABLE IF NOT EXISTS t_person_info (
  age INT(11) unsigned NULL,
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT NOT NULL COMMENT '这是id',
  name VARCHAR(30) NOT NULL DEFAULT 'default_name' COMMENT '这是名字',
  sex VARCHAR(2) NULL,
  comment TEXT
  ) ENGINE=InnoDB;`
	data, err := ParseSql(sql, WithTablePrefix("t_"))
	assert.Nil(t, err)
	for _, s := range data.StructCode {
		t.Log(s)
	}
	t.Log(data.ImportPath)
}

func TestParseSqlToWritee(t *testing.T) {
	sql := `CREATE TABLE IF NOT EXISTS t_person_info (
  age INT(11) unsigned NULL,
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT NOT NULL COMMENT '这是id',
  name VARCHAR(30) NOT NULL DEFAULT 'default_name' COMMENT '这是名字',
  sex VARCHAR(2) NULL,
  comment TEXT
  ) ENGINE=InnoDB;`
	w := strings.Builder{}
	err := ParseSqlToWrite(sql, &w, WithTablePrefix("t_"))
	if !assert.Nil(t, err) {
		t.Log(err)
	}
}
