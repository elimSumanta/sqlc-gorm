package golang

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ujunglangit-id/sqlc/internal/core"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/ujunglangit-id/sqlc/internal/codegen"
	"github.com/ujunglangit-id/sqlc/internal/compiler"
	"github.com/ujunglangit-id/sqlc/internal/config"
)

type Generateable interface {
	Structs(settings config.CombinedSettings) []Struct
	GoQueries(settings config.CombinedSettings) []Query
	Enums(settings config.CombinedSettings) []Enum
}

type TargetFiles struct {
	FilePath string
	FileBody string
	FileName string
}

var xtemplateSet = `
{{define "dbFile"}}// Code generated by sqlc. DO NOT EDIT.

package {{.Package}}

import (
	{{range imports .SourceName}}
	{{range .}}{{.}}
	{{end}}
	{{end}}
)

{{template "dbCode" . }}
{{end}}

{{define "dbCode"}}
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

{{if .EmitPreparedQueries}}
func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	{{- if eq (len .GoQueries) 0 }}
	_ = err
	{{- end }}
	{{- range .GoQueries }}
	if q.{{.FieldName}}, err = db.PrepareContext(ctx, {{.ConstantName}}); err != nil {
		return nil, fmt.Errorf("error preparing query {{.MethodName}}: %w", err)
	}
	{{- end}}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	{{- range .GoQueries }}
	if q.{{.FieldName}} != nil {
		if cerr := q.{{.FieldName}}.Close(); cerr != nil {
			err = fmt.Errorf("error closing {{.FieldName}}: %w", cerr)
		}
	}
	{{- end}}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Row) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}
{{end}}

type Queries struct {
	db DBTX

    {{- if .EmitPreparedQueries}}
	tx         *sql.Tx
	{{- range .GoQueries}}
	{{.FieldName}}  *sql.Stmt
	{{- end}}
	{{- end}}
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
     	{{- if .EmitPreparedQueries}}
		tx: tx,
		{{- range .GoQueries}}
		{{.FieldName}}: q.{{.FieldName}},
		{{- end}}
		{{- end}}
	}
}
{{end}}

{{define "interfaceFile"}}// Code generated by sqlc. DO NOT EDIT.

package {{.Package}}

import (
	{{range imports .SourceName}}
	{{range .}}{{.}}
	{{end}}
	{{end}}
)

{{template "interfaceCode" . }}
{{end}}

{{define "interfaceCode"}}
type Querier interface {
	{{- range .GoQueries}}
	{{- if eq .Cmd ":one"}}
	{{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) ({{.Ret.Type}}, error)
	{{- end}}
	{{- if eq .Cmd ":many"}}
	{{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) ([]{{.Ret.Type}}, error)
	{{- end}}
	{{- if eq .Cmd ":exec"}}
	{{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) error
	{{- end}}
	{{- if eq .Cmd ":execrows"}}
	{{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) (int64, error)
	{{- end}}
	{{- if eq .Cmd ":execresult"}}
	{{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) (sql.Result, error)
	{{- end}}
	{{- end}}
}

var _ Querier = (*Queries)(nil)
{{end}}

{{define "modelsFile"}}// Code generated by sqlc. DO NOT EDIT.

package {{.Package}}

import (
	{{range imports .SourceName}}
	{{range .}}{{.}}
	{{end}}
	{{end}}
)

{{template "modelsCode" . }}
{{end}}

{{define "modelsCode"}}
{{range .Enums}}
{{if .Comment}}{{comment .Comment}}{{end}}
type {{.Name}} string

const (
	{{- range .Constants}}
	{{.Name}} {{.Type}} = "{{.Value}}"
	{{- end}}
)

func (e *{{.Name}}) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = {{.Name}}(s)
	case string:
		*e = {{.Name}}(s)
	default:
		return fmt.Errorf("unsupported scan type for {{.Name}}: %T", src)
	}
	return nil
}
{{end}}

{{range .Structs}}
{{if .Comment}}{{comment .Comment}}{{end}}
type {{.Name}} struct { 
  tableName      struct{} {{$.Q}}pg:{{.Name}},alias:n{{$.Q}}
  {{- range .Fields}}
  {{- if .Comment}}
  {{comment .Comment}}{{else}}
  {{- end}}
  {{.Name}} {{.Type}} {{if or ($.EmitJSONTags) ($.EmitDBTags)}}{{$.Q}}{{.Tag}}{{$.Q}}{{end}}
  {{- end}}
}
{{end}}
{{end}}

{{define "queryFile"}}// Code generated by sqlc. DO NOT EDIT.
// source: {{.SourceName}}

package {{.Package}}

import (
	{{range imports .SourceName}}
	{{range .}}{{.}}
	{{end}}
	{{end}}
)

{{template "queryCode" . }}
{{end}}

{{define "queryCode"}}
{{range .GoQueries}}
{{if $.OutputQuery .SourceName}}
const {{.ConstantName}} = {{$.Q}}-- name: {{.MethodName}} {{.Cmd}}
{{escape .SQL}}
{{$.Q}}

{{if .Arg.EmitStruct}}
type {{.Arg.Type}} struct { {{- range .Arg.Struct.Fields}}
  {{.Name}} {{.Type}} {{if or ($.EmitJSONTags) ($.EmitDBTags)}}{{$.Q}}{{.Tag}}{{$.Q}}{{end}}
  {{- end}}
}
{{end}}

{{if .Ret.EmitStruct}}
type {{.Ret.Type}} struct { {{- range .Ret.Struct.Fields}}
  {{.Name}} {{.Type}} {{if or ($.EmitJSONTags) ($.EmitDBTags)}}{{$.Q}}{{.Tag}}{{$.Q}}{{end}}
  {{- end}}
}
{{end}}

{{if eq .Cmd ":one"}}
{{range .Comments}}//{{.}}
{{end -}}
func (q *Queries) {{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) ({{.Ret.Type}}, error) {
  	{{- if $.EmitPreparedQueries}}
	row := q.queryRow(ctx, q.{{.FieldName}}, {{.ConstantName}}, {{.Arg.Params}})
	{{- else}}
	row := q.db.QueryRowContext(ctx, {{.ConstantName}}, {{.Arg.Params}})
	{{- end}}
	var {{.Ret.Name}} {{.Ret.Type}}
	err := row.Scan({{.Ret.Scan}})
	return {{.Ret.Name}}, err
}
{{end}}

{{if eq .Cmd ":many"}}
{{range .Comments}}//{{.}}
{{end -}}
func (q *Queries) {{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) ([]{{.Ret.Type}}, error) {
  	{{- if $.EmitPreparedQueries}}
	rows, err := q.query(ctx, q.{{.FieldName}}, {{.ConstantName}}, {{.Arg.Params}})
  	{{- else}}
	rows, err := q.db.QueryContext(ctx, {{.ConstantName}}, {{.Arg.Params}})
  	{{- end}}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	{{- if $.EmitEmptySlices}}
	items := []{{.Ret.Type}}{}
	{{else}}
	var items []{{.Ret.Type}}
	{{end -}}
	for rows.Next() {
		var {{.Ret.Name}} {{.Ret.Type}}
		if err := rows.Scan({{.Ret.Scan}}); err != nil {
			return nil, err
		}
		items = append(items, {{.Ret.Name}})
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
{{end}}

{{if eq .Cmd ":exec"}}
{{range .Comments}}//{{.}}
{{end -}}
func (q *Queries) {{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) error {
  	{{- if $.EmitPreparedQueries}}
	_, err := q.exec(ctx, q.{{.FieldName}}, {{.ConstantName}}, {{.Arg.Params}})
  	{{- else}}
	_, err := q.db.ExecContext(ctx, {{.ConstantName}}, {{.Arg.Params}})
  	{{- end}}
	return err
}
{{end}}

{{if eq .Cmd ":execrows"}}
{{range .Comments}}//{{.}}
{{end -}}
func (q *Queries) {{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) (int64, error) {
  	{{- if $.EmitPreparedQueries}}
	result, err := q.exec(ctx, q.{{.FieldName}}, {{.ConstantName}}, {{.Arg.Params}})
  	{{- else}}
	result, err := q.db.ExecContext(ctx, {{.ConstantName}}, {{.Arg.Params}})
  	{{- end}}
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
{{end}}

{{if eq .Cmd ":execresult"}}
{{range .Comments}}//{{.}}
{{end -}}
func (q *Queries) {{.MethodName}}(ctx context.Context, {{.Arg.Pair}}) (sql.Result, error) {
  	{{- if $.EmitPreparedQueries}}
	return q.exec(ctx, q.{{.FieldName}}, {{.ConstantName}}, {{.Arg.Params}})
  	{{- else}}
	return q.db.ExecContext(ctx, {{.ConstantName}}, {{.Arg.Params}})
  	{{- end}}
}
{{end}}
{{end}}
{{end}}
{{end}}
`

type tmplCtx struct {
	Q          string
	Package    string
	Enums      []Enum
	Structs    []Struct
	StructData Struct
	GoQueries  []Query
	Settings   config.Config

	// TODO: Race conditions
	SourceName string

	EmitJSONTags        bool
	EmitDBTags          bool
	EmitFormTags        bool
	EmitPreparedQueries bool
	EmitInterface       bool
	EmitEmptySlices     bool
	//struct field
	Table       core.FQN
	ProjectPath string
	Name        string
	Fields      []Field
	Comment     string
}

func (t *tmplCtx) OutputQuery(sourceName string) bool {
	return t.SourceName == sourceName
}

//func Generate(r *compiler.Result, settings config.CombinedSettings) (map[string]string, error) {
func Generate(r *compiler.Result, settings config.CombinedSettings) ([]*TargetFiles, error) {
	enums := buildEnums(r, settings)
	structs := buildStructs(r, settings)
	queries := buildQueries(r, settings, structs)
	return generate(settings, enums, structs, queries)
}

//func generate(settings config.CombinedSettings, enums []Enum, structs []Struct, queries []Query) (map[string]string, error) {
func generate(settings config.CombinedSettings, enums []Enum, structs []Struct, queries []Query) (targetFile []*TargetFiles, err error) {
	var (
		targetRepoPath, targetModelpath string
	)
	targetFile = []*TargetFiles{}
	i := &importer{
		Settings: settings,
		Queries:  queries,
		Enums:    enums,
		Structs:  structs,
	}

	funcMap := template.FuncMap{
		"lowerTitle": codegen.LowerTitle,
		"comment":    codegen.DoubleSlashComment,
		"escape":     codegen.EscapeBacktick,
		"imports":    i.Imports,
	}
	golang := settings.Go
	output := map[string]string{}

	//prepare repository folder
	targetRepoPath, err = makeMultiDirectoryIfNotExists(golang.Out, map[int]string{
		0: "repository",
		1: "postgre",
	})
	if err != nil {
		fmt.Printf("failed create repo folder, err : %+v \n\n", err)
	}
	fmt.Println("done create repository folder")

	//prepare model folder
	targetModelpath, err = makeMultiDirectoryIfNotExists(golang.Out, map[int]string{
		0: "model",
		1: "types",
	})

	if err != nil {
		fmt.Printf("failed create model folder, err : %+v \n\n", err)
	}
	fmt.Println("done create model folder")
	//eof root folder

	for _, v := range structs {
		tmpl := template.Must(template.New("table").Funcs(funcMap).Parse(templateModelRepo))

		tctx := tmplCtx{
			Settings:            settings.Global,
			EmitInterface:       golang.EmitInterface,
			EmitJSONTags:        golang.EmitJSONTags,
			EmitDBTags:          golang.EmitDBTags,
			EmitFormTags:        golang.EmitFormTags,
			EmitPreparedQueries: golang.EmitPreparedQueries,
			EmitEmptySlices:     golang.EmitEmptySlices,
			Q:                   "`",
			Package:             golang.Package,
			GoQueries:           queries,
			Enums:               enums,
			Table:               v.Table,
			ProjectPath:         v.ProjectPath,
			Name:                v.Name,
			Fields:              v.Fields,
			Comment:             v.Comment,
		}

		//repo file generator
		executeRepo := func(subPath bool, name, templateName string) error {
			targetSubpath := targetRepoPath
			if subPath {
				//prepare sub repo folder if not exist
				targetSubpath = targetSubpath + "/" + v.Name
				err = makeDirIfNotExists(targetSubpath)
				if err != nil {
					fmt.Printf("failed create target sub repo folder, err : %+v \n\n", err)
				}
			}

			var b bytes.Buffer
			w := bufio.NewWriter(&b)
			tctx.SourceName = name
			err := tmpl.ExecuteTemplate(w, templateName, &tctx)
			w.Flush()
			if err != nil {
				return err
			}
			code, err := format.Source(b.Bytes())
			if err != nil {
				fmt.Println(b.String())
				return fmt.Errorf("source error: %w", err)
			}
			if !strings.HasSuffix(name, ".go") {
				name += ".go"
			}

			output[name] = string(code)
			targetFile = append(targetFile, &TargetFiles{
				FilePath: targetSubpath,
				FileBody: string(code),
				FileName: name,
			})
			return nil
		}
		//model file generator
		executeModel := func(subPath bool, name, templateName string) error {
			targetSubpath := targetModelpath
			if subPath {
				//prepare sub model folder if not exist
				targetSubpath = targetSubpath + "/" + v.Name
				err = makeDirIfNotExists(targetSubpath)
				if err != nil {
					fmt.Printf("failed create target sub model folder, err : %+v \n\n", err)
				}
			}

			var b bytes.Buffer
			w := bufio.NewWriter(&b)
			tctx.SourceName = name
			err := tmpl.ExecuteTemplate(w, templateName, &tctx)
			w.Flush()
			if err != nil {
				return err
			}
			code, err := format.Source(b.Bytes())
			if err != nil {
				fmt.Println(b.String())
				return fmt.Errorf("source error: %w", err)
			}
			if !strings.HasSuffix(name, ".go") {
				name += ".go"
			}

			output[name] = string(code)
			targetFile = append(targetFile, &TargetFiles{
				FilePath: targetSubpath,
				FileBody: string(code),
				FileName: name,
			})
			return nil
		}

		//generate repository file
		if err := executeRepo(false, v.Name+"Repository.go", "repoInterfaceFile"); err != nil {
			return nil, err
		}
		if err := executeRepo(true, v.Name+"RepoImpl.go", "repoImplFile"); err != nil {
			return nil, err
		}

		//generate model file
		if err := executeModel(true, "entity.go", "entityFile"); err != nil {
			return nil, err
		}
		if err := executeModel(true, "payload.go", "payloadFile"); err != nil {
			return nil, err
		}

	}
	//if golang.EmitInterface {
	//	if err := execute("querier.go", "interfaceFile"); err != nil {
	//		return nil, err
	//	}
	//}

	//files := map[string]struct{}{}
	//for _, gq := range queries {
	//	files[gq.SourceName] = struct{}{}
	//}

	//for source := range files {
	//	if err := execute(source, "queryFile"); err != nil {
	//		return nil, err
	//	}
	//}
	return
}

func makeMultiDirectoryIfNotExists(rootPath string, pathList map[int]string) (targetPath string, err error) {
	//var targetpath string
	targetPath, err = os.Getwd()
	if err != nil {
		return
	}
	targetPath = targetPath + "/" + rootPath
	for _, v := range pathList {
		targetPath = targetPath + "/" + v
		err = makeDirIfNotExists(targetPath)
	}
	return
}

func makeDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}
