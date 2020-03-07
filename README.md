# sql2gorm

sql2gorm is a command line tool to generate Go code with [Gorm](https://gorm.io/) Tag and Json Tag from SQL.

Web tool: https://sql2gorm.mccode.info/

## Download

```
go get -u github.com/cascax/sql2gorm/...
```

## Usage (Command Line)

```
sql2gorm -f file.sql -o model.go
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
