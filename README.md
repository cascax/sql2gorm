# sql2gorm

sql2gorm is a command line tool to generate Go code with [Gorm](https://gorm.io/) Tag and Json Tag from SQL.

Web tool: https://sql2gorm.mccode.info/

## Download

```
go get -u github.com/cascax/sql2gorm/...
```

## Usage (Command Line)

get struct from a sql file and write struct to the file

```
sql2gorm -f file.sql -o model.go
```

get struct from mysql

```
sql2gorm -db-dsn=root:123456@/msir -db-table=fund_info
```

get struct from arguments

```
sql2gorm -sql="CREATE TABLE person_info (
  age INT(11) unsigned NULL,
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT NOT NULL COMMENT '这是id',
  name VARCHAR(30) NOT NULL DEFAULT 'default_name' COMMENT '这是名字',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  sex VARCHAR(2) NULL,
  num INT(11) DEFAULT 3 NULL,
  comment TEXT
  ) COMMENT='person info';"
```

## Library usage

```go
import "github.com/cascax/sql2gorm/parser"

sql := `CREATE TABLE t_person_info (
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT NOT NULL COMMENT 'primary id',
  name VARCHAR(30) NOT NULL DEFAULT 'default_name',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  age INT(11) unsigned NULL,
  sex VARCHAR(2) NULL
  );`
data, err := parser.ParseSql(sql, WithTablePrefix("t_"), WithJsonTag())
```
