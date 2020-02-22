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
	Charset      string
	Collation    string
	JsonTag      bool
	TablePrefix  string
	ColumnPrefix string
	NoNullType   bool
	NullStyle    string
	Package      string

	InputFile  string
	OutputFile string
	Sql        string
}

func parseFlag() options {
	args := options{}
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
	return opt
}

func main() {
	args := parseFlag()

	var output io.Writer
	if args.OutputFile != "" {
		f, err := os.OpenFile(args.OutputFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Printf("open %s failed, %s\n", args.OutputFile, err)
			return
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
				fmt.Printf("read %s failed, %s\n", args.InputFile, err)
				return
			}
			sql = string(b)
		} else {
			fmt.Println("no SQL input")
			return
		}
	}

	opt := getOptions(args)
	if opt == nil {
		return
	}

	err := parser.ParseSqlToWrite(sql, output, opt...)
	if err != nil {
		fmt.Println(err)
	}
}
