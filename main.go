package main

import (
	"flag"
	"fmt"
	"github.com/cascax/sql2gorm/parser"
	"io"
	"io/ioutil"
	"os"
)

type options struct {
	Charset        string
	Collation      string
	JsonTag        bool
	TablePrefix    string
	ColumnPrefix   string
	NoNullType     bool
	NullStyle      string
	Package        string
	GormType       bool
	ForceTableName bool

	InputFile  string
	OutputFile string
	Sql        string

	MysqlDsn   string
	MysqlTable string
}

func exitWithInfo(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func parseFlag() options {
	args := options{}
	//flagSet := flag.NewFlagSet("optional", flag.ExitOnError)

	flag.StringVar(&args.InputFile, "f", "", "input file")
	flag.StringVar(&args.OutputFile, "o", "", "output file")
	flag.StringVar(&args.Sql, "sql", "", "input SQL")

	flag.BoolVar(&args.JsonTag, "json", false, "generate json tag")
	flag.StringVar(&args.TablePrefix, "table-prefix", "", "table name prefix")
	flag.StringVar(&args.ColumnPrefix, "col-prefix", "", "column name prefix")
	flag.BoolVar(&args.NoNullType, "no-null", false, "do not use Null type")
	flag.StringVar(&args.NullStyle, "null-style", "",
		"null type: sql.NullXXX(use 'sql') or *xxx(use 'ptr')")
	flag.StringVar(&args.Package, "pkg", "", "package name, default: model")
	flag.BoolVar(&args.GormType, "with-type", false, "write type in gorm tag")
	flag.BoolVar(&args.ForceTableName, "with-tablename", false, "write TableName func force")

	flag.StringVar(&args.MysqlDsn, "db-dsn", "", "mysql dsn([user]:[pass]@/[database][?charset=xxx&...])")
	flag.StringVar(&args.MysqlTable, "db-table", "", "mysql table name")

	flag.Parse()
	return args
}

func getOptions(args options) []parser.Option {
	opt := make([]parser.Option, 0, 1)
	if args.Charset != "" {
		opt = append(opt, parser.WithCharset(args.Charset))
	}
	if args.Collation != "" {
		opt = append(opt, parser.WithCollation(args.Collation))
	}
	if args.JsonTag {
		opt = append(opt, parser.WithJsonTag())
	}
	if args.TablePrefix != "" {
		opt = append(opt, parser.WithTablePrefix(args.TablePrefix))
	}
	if args.ColumnPrefix != "" {
		opt = append(opt, parser.WithColumnPrefix(args.ColumnPrefix))
	}
	if args.NoNullType {
		opt = append(opt, parser.WithNoNullType())
	}
	if args.NullStyle != "" {
		switch args.NullStyle {
		case "sql":
			opt = append(opt, parser.WithNullStyle(parser.NullInSql))
		case "ptr":
			opt = append(opt, parser.WithNullStyle(parser.NullInPointer))
		default:
			fmt.Printf("invalid null style: %s\n", args.NullStyle)
			return nil
		}
	}
	if args.Package != "" {
		opt = append(opt, parser.WithPackage(args.Package))
	}
	if args.GormType {
		opt = append(opt, parser.WithGormType())
	}
	if args.ForceTableName {
		opt = append(opt, parser.WithForceTableName())
	}
	return opt
}

func main() {
	args := parseFlag()

	var output io.Writer
	if args.OutputFile != "" {
		f, err := os.OpenFile(args.OutputFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			exitWithInfo("open %s failed, %s\n", args.OutputFile, err)
		}
		defer f.Close()
		output = f
	} else {
		output = os.Stdout
	}
	sql := args.Sql
	if sql == "" {
		if args.InputFile != "" {
			b, err := ioutil.ReadFile(args.InputFile)
			if err != nil {
				exitWithInfo("read %s failed, %s\n", args.InputFile, err)
			}
			sql = string(b)
		} else if args.MysqlDsn != "" {
			if args.MysqlTable == "" {
				exitWithInfo("miss mysql table")
			}
			var err error
			sql, err = parser.GetCreateTableFromDB(args.MysqlDsn, args.MysqlTable)
			if err != nil {
				exitWithInfo("get create table error: %s", err)
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "no SQL input(-sql|-f|-db-dsn)\n\n")
			flag.Usage()
			os.Exit(2)
		}
	}

	opt := getOptions(args)
	if opt == nil {
		return
	}

	err := parser.ParseSqlToWrite(sql, output, opt...)
	if err != nil {
		exitWithInfo(err.Error())
	}
}
