package parser

import (
	"github.com/iancoleman/strcase"
	"github.com/knocknote/vitess-sqlparser/tidbparser/ast"
	"github.com/knocknote/vitess-sqlparser/tidbparser/dependency/mysql"
	"github.com/knocknote/vitess-sqlparser/tidbparser/parser"
	"go/format"
	"io"
	"strings"
	"sync"
	"text/template"
)

var (
	structTmplRaw string
	fileTmplRaw   string
	structTmpl    *template.Template
	fileTmpl      *template.Template
	tmplParseOnce sync.Once
)

type ModelCodes struct {
	Package    string
	ImportPath []string
	StructCode []string
}

func ParseSql(sql string, options ...Option) (ModelCodes, error) {
	initTemplate()
	opt := parseOption(options)

	stmts, err := parser.New().Parse(sql, opt.Charset, opt.Collation)
	if err != nil {
		return ModelCodes{}, err
	}
	tableStr := make([]string, 0, len(stmts))
	importPath := make(map[string]struct{})
	for _, stmt := range stmts {
		if ct, ok := stmt.(*ast.CreateTableStmt); ok {
			s, ipt, err := makeCode(ct, opt)
			if err != nil {
				return ModelCodes{}, err
			}
			tableStr = append(tableStr, s)
			for _, s := range ipt {
				importPath[s] = struct{}{}
			}
		}
	}
	importPathArr := make([]string, 0, len(importPath))
	for s := range importPath {
		importPathArr = append(importPathArr, s)
	}
	return ModelCodes{
		Package:    opt.Package,
		ImportPath: importPathArr,
		StructCode: tableStr,
	}, nil
}

func ParseSqlToWrite(sql string, writer io.Writer, options ...Option) error {
	data, err := ParseSql(sql, options...)
	if err != nil {
		return err
	}
	err = fileTmpl.Execute(writer, data)
	if err != nil {
		return err
	}
	return nil
}

type tmplData struct {
	TableName    string
	NameFunc     bool
	RawTableName string
	Fields       []tmplField
}

type tmplField struct {
	Name    string
	GoType  string
	Tag     string
	Comment string
}

func makeCode(stmt *ast.CreateTableStmt, opt options) (string, []string, error) {
	importPath := make([]string, 0, 1)
	data := tmplData{
		TableName:    stmt.Table.Name.String(),
		RawTableName: stmt.Table.Name.String(),
		Fields:       make([]tmplField, 0, 1),
	}
	tablePrefix := opt.TablePrefix
	if tablePrefix != "" && strings.HasPrefix(data.TableName, tablePrefix) {
		data.NameFunc = true
		data.TableName = data.TableName[len(tablePrefix):]
	}
	data.TableName = strcase.ToCamel(data.TableName)

	columnPrefix := opt.ColumnPrefix
	for _, col := range stmt.Cols {
		colName := col.Name.Name.String()
		goFieldName := colName
		if columnPrefix != "" && strings.HasPrefix(goFieldName, columnPrefix) {
			goFieldName = goFieldName[len(columnPrefix):]
		}

		goType, pkg := mysqlToGoType(col.Tp.Tp)
		field := tmplField{
			Name:   strcase.ToCamel(goFieldName),
			GoType: goType,
		}
		if pkg != "" {
			importPath = append(importPath, pkg)
		}

		tags := make([]string, 0, 4)
		// 生成GORM tag和修正类型
		gormTag := strings.Builder{}
		gormTag.WriteString("column:")
		gormTag.WriteString(colName)
		isPrimaryKey := false
		isNotNull := false
		for _, o := range col.Options {
			switch o.Tp {
			case ast.ColumnOptionPrimaryKey:
				gormTag.WriteString(";primary_key")
				isPrimaryKey = true
			case ast.ColumnOptionNotNull:
				isNotNull = true
			case ast.ColumnOptionAutoIncrement:
				gormTag.WriteString(";AUTO_INCREMENT")
			case ast.ColumnOptionDefaultValue:
				gormTag.WriteString(";default:")
				gormTag.WriteString(o.Expr.GetDatum().GetString())
			case ast.ColumnOptionUniqKey:
				gormTag.WriteString(";unique")
			case ast.ColumnOptionNull:
				//gormTag.WriteString(";NULL")
				if !opt.NoNullType {
					if opt.NullStyle == NullInPointer {
						field.GoType = "*" + field.GoType
					} else {
						importPath = append(importPath, "database/sql")
						if strings.Index(field.GoType, ".") < 0 {
							if strings.Index(field.GoType, "int") >= 0 {
								field.GoType = "sql.NullInt64"
							} else {
								field.GoType = "sql.Null" + strings.ToUpper(field.GoType[:1]) + field.GoType[1:]
							}
						}
					}
				}
			case ast.ColumnOptionOnUpdate: // For Timestamp and Datetime only.
			case ast.ColumnOptionFulltext:
			case ast.ColumnOptionComment:
				field.Comment = o.Expr.GetDatum().GetString()
			default:
				//return "", nil, errors.Errorf(" unsupport option %d\n", o.Tp)
			}
		}
		if !isPrimaryKey && isNotNull {
			gormTag.WriteString(";NOT NULL")
		}
		tags = append(tags, "gorm", gormTag.String())

		if opt.JsonTag {
			tags = append(tags, "json", colName)
		}

		field.Tag = makeTagStr(tags)
		data.Fields = append(data.Fields, field)
	}

	builder := strings.Builder{}
	err := structTmpl.Execute(&builder, data)
	if err != nil {
		return "", nil, err
	}
	code, err := format.Source([]byte(builder.String()))
	return string(code), importPath, err
}

func mysqlToGoType(colTp byte) (string, string) {
	switch colTp {
	case mysql.TypeTiny, mysql.TypeShort, mysql.TypeInt24, mysql.TypeLong:
		return "int", ""
	case mysql.TypeLonglong:
		return "int64", ""
	case mysql.TypeFloat, mysql.TypeDouble:
		return "float64", ""
	case mysql.TypeString, mysql.TypeVarchar, mysql.TypeVarString,
		mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob:
		return "string", ""
	case mysql.TypeTimestamp, mysql.TypeDatetime, mysql.TypeDate:
		return "time.Date", "time"
	case mysql.TypeDecimal, mysql.TypeNewDecimal:
		return "string", ""
	case mysql.TypeJSON:
		return "string", ""
	default:
		return "UnSupport", ""
	}
}

func makeTagStr(tags []string) string {
	builder := strings.Builder{}
	for i := 0; i < len(tags)/2; i++ {
		builder.WriteString(tags[i*2])
		builder.WriteString(`:"`)
		builder.WriteString(tags[i*2+1])
		builder.WriteString(`" `)
	}
	if builder.Len() > 0 {
		return builder.String()[:builder.Len()-1]
	}
	return builder.String()
}

func initTemplate() {
	tmplParseOnce.Do(func() {
		var err error
		structTmpl, err = template.New("goStruct").Parse(structTmplRaw)
		if err != nil {
			panic(err)
		}
		fileTmpl, err = template.New("goFile").Parse(fileTmplRaw)
		if err != nil {
			panic(err)
		}
	})
}

func init() {
	structTmplRaw = `
type {{.TableName}} struct {
{{- range .Fields}}
	{{.Name}} {{.GoType}} {{if .Tag}}` + "`{{.Tag}}`" + `{{end}}{{if .Comment}} // {{.Comment}}{{end}}
{{- end}}
}
{{if .NameFunc}}
func (m *{{.TableName}}) TableName() string {
	return "{{.RawTableName}}"
}
{{end}}`
	fileTmplRaw = `// Code generated by sql2gorm. DO NOT EDIT.
package {{.Package}}
{{if .ImportPath}}
import (
	{{- range .ImportPath}}
	"{{.}}"
	{{- end}}
)
{{- end}}
{{range .StructCode}}
{{.}}
{{end}}
`
}
