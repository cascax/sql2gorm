package parser

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var testData = [][]string{
	{
		"CREATE TABLE info (age INT(11) NULL);",
		"Age int `gorm:\"column:age\"`", "",
	},
	{
		"CREATE TABLE info (age BIGINT(11) NULL COMMENT 'is age');",
		"Age int64 `gorm:\"column:age\"` // is age", "",
	},
	{
		"CREATE TABLE info (id BIGINT(11) PRIMARY KEY AUTO_INCREMENT);",
		"Id int64 `gorm:\"column:id;primary_key;AUTO_INCREMENT\"`", "",
	},
	{
		"CREATE TABLE info (created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP);",
		"CreatedAt time.Time `gorm:\"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL\"`", "time",
	},
	{
		"CREATE TABLE info (num INT(11) DEFAULT 3 NULL);",
		"Num int `gorm:\"column:num;default:3\"`", "",
	},
	{
		"CREATE TABLE info (num double(5,6) DEFAULT 31.50 NULL);",
		"Num float64 `gorm:\"column:num;default:31.50\"`", "",
	},
	{
		"CREATE TABLE info (comment TEXT);",
		"Comment string `gorm:\"column:comment\"`", "",
	},
	{
		"CREATE TABLE info (comment TINYTEXT);",
		"Comment string `gorm:\"column:comment\"`", "",
	},
	{
		"CREATE TABLE info (comment LONGTEXT);",
		"Comment string `gorm:\"column:comment\"`", "",
	},
}

func TestParseSql(t *testing.T) {
	sql := `CREATE TABLE t_person_info (
  age INT(11) unsigned NULL,
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT NOT NULL COMMENT '这是id',
  name VARCHAR(30) NOT NULL DEFAULT 'default_name' COMMENT '这是名字',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  sex VARCHAR(2) NULL,
  num INT(11) DEFAULT 3 NULL,
  comment TEXT
  ) COMMENT="person info";`
	data, err := ParseSql(sql, WithTablePrefix("t_"), WithJsonTag())
	assert.Nil(t, err)
	for _, s := range data.StructCode {
		t.Log(s)
	}
	t.Log(data.ImportPath)
}

func TestParseSqlToWrite(t *testing.T) {
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

func TestParseSql1(t *testing.T) {
	for _, test := range testData {
		data, err := ParseSql(test[0], WithNoNullType())
		if !assert.Nil(t, err) {
			continue
		}
		//t.Log(data.StructCode)
		//t.Log(data.ImportPath)
		if assert.Equal(t, 1, len(data.StructCode)) {
			lines := strings.Split(strings.TrimSpace(data.StructCode[0]), "\n")
			if assert.Equal(t, 3, len(lines)) {
				assert.Equal(t, test[1], strings.TrimSpace(lines[1]))
			}
		}
		if test[2] == "" {
			assert.Equal(t, 0, len(data.ImportPath))
		} else {
			if assert.Equal(t, 1, len(data.ImportPath)) {
				assert.Equal(t, test[2], data.ImportPath[0])
			}
		}
	}
}
