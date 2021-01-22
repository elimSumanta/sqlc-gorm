package main

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	pgx "github.com/jackc/pgx/v4"

	"github.com/ujunglangit-id/sqlc/internal/sql/ast"
	"github.com/ujunglangit-id/sqlc/internal/sql/catalog"
)

// https://stackoverflow.com/questions/25308765/postgresql-how-can-i-inspect-which-arguments-to-a-procedure-have-a-default-valu
const catalogFuncs = `
SELECT p.proname as name,
  format_type(p.prorettype, NULL),
  array(select format_type(unnest(p.proargtypes), NULL)),
  p.proargnames,
  p.proargnames[p.pronargs-p.pronargdefaults+1:p.pronargs]
FROM pg_catalog.pg_proc p
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname OPERATOR(pg_catalog.~) '^(pg_catalog)$'
  AND p.proargmodes IS NULL
  AND pg_function_is_visible(p.oid)
ORDER BY 1;
`

// https://dba.stackexchange.com/questions/255412/how-to-select-functions-that-belong-in-a-given-extension-in-postgresql
//
// Extension functions are added to the public schema
const extensionFuncs = `
WITH extension_funcs AS (
  SELECT p.oid
  FROM pg_catalog.pg_extension AS e
      INNER JOIN pg_catalog.pg_depend AS d ON (d.refobjid = e.oid)
      INNER JOIN pg_catalog.pg_proc AS p ON (p.oid = d.objid)
      INNER JOIN pg_catalog.pg_namespace AS ne ON (ne.oid = e.extnamespace)
      INNER JOIN pg_catalog.pg_namespace AS np ON (np.oid = p.pronamespace)
  WHERE d.deptype = 'e' AND e.extname = $1
)
SELECT p.proname as name,
  format_type(p.prorettype, NULL),
  array(select format_type(unnest(p.proargtypes), NULL)),
  p.proargnames,
  p.proargnames[p.pronargs-p.pronargdefaults+1:p.pronargs]
FROM pg_catalog.pg_proc p
JOIN extension_funcs ef ON ef.oid = p.oid
WHERE p.proargmodes IS NULL
  AND pg_function_is_visible(p.oid)
ORDER BY 1;
`

const catalogTmpl = `
// Code generated by sqlc-pg-gen. DO NOT EDIT.

package {{.Pkg}}

import (
	"github.com/ujunglangit-id/sqlc/internal/sql/ast"
	"github.com/ujunglangit-id/sqlc/internal/sql/catalog"
)

func {{.Name}}() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = []*catalog.Function{
	    {{- range .Funcs}}
		{
			Name: "{{.Name}}",
			Args: []*catalog.Argument{
				{{range .Args}}{
				{{- if .Name}}
				Name: "{{.Name}}",
				{{- end}}
				{{- if .HasDefault}}
				HasDefault: true,
				{{- end}}
				Type: &ast.TypeName{Name: "{{.Type.Name}}"},
				},
				{{end}}
			},
			ReturnType: &ast.TypeName{Name: "{{.ReturnType.Name}}"},
		},
		{{- end}}
	}
	return s
}
`

const loaderFuncTmpl = `
// Code generated by sqlc-pg-gen. DO NOT EDIT.

package postgresql

import (
	"github.com/ujunglangit-id/sqlc/internal/engine/postgresql/contrib"
	"github.com/ujunglangit-id/sqlc/internal/sql/catalog"
)

func loadExtension(name string) *catalog.Schema {
	switch name {
	{{- range .}}
	case "{{.Name}}":
		return contrib.{{.Func}}()
	{{- end}}
	}
	return nil
}
`

type tmplCtx struct {
	Pkg   string
	Name  string
	Funcs []catalog.Function
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

type Proc struct {
	Name       string
	ReturnType string
	ArgTypes   []string
	ArgNames   []string
	HasDefault []string
}

func clean(arg string) string {
	arg = strings.TrimSpace(arg)
	arg = strings.Replace(arg, "\"any\"", "any", -1)
	arg = strings.Replace(arg, "\"char\"", "char", -1)
	arg = strings.Replace(arg, "\"timestamp\"", "char", -1)
	return arg
}

func (p Proc) Func() catalog.Function {
	return catalog.Function{
		Name:       p.Name,
		Args:       p.Args(),
		ReturnType: &ast.TypeName{Name: clean(p.ReturnType)},
	}
}

func (p Proc) Args() []*catalog.Argument {
	defaults := map[string]bool{}
	var args []*catalog.Argument
	if len(p.ArgTypes) == 0 {
		return args
	}
	for _, name := range p.HasDefault {
		defaults[name] = true
	}
	for i, arg := range p.ArgTypes {
		var name string
		if i < len(p.ArgNames) {
			name = p.ArgNames[i]
		}
		args = append(args, &catalog.Argument{
			Name:       name,
			HasDefault: defaults[name],
			Type:       &ast.TypeName{Name: clean(arg)},
		})
	}
	return args
}

func scanFuncs(rows pgx.Rows) ([]catalog.Function, error) {
	defer rows.Close()
	// Iterate through the result set
	var funcs []catalog.Function
	for rows.Next() {
		var p Proc
		err := rows.Scan(
			&p.Name,
			&p.ReturnType,
			&p.ArgTypes,
			&p.ArgNames,
			&p.HasDefault,
		)
		if err != nil {
			return nil, err
		}

		// TODO: Filter these out in SQL
		if strings.HasPrefix(p.ReturnType, "SETOF") {
			continue
		}

		// The internal pseudo-type is used to declare functions that are meant
		// only to be called internally by the database system, and not by
		// direct invocation in an SQL query. If a function has at least one
		// internal-type argument then it cannot be called from SQL. To
		// preserve the type safety of this restriction it is important to
		// follow this coding rule: do not create any function that is declared
		// to return internal unless it has at least one internal argument
		//
		// https://www.postgresql.org/docs/current/datatype-pseudo.html
		var skip bool
		for i := range p.ArgTypes {
			if p.ArgTypes[i] == "internal" {
				skip = true
			}
		}
		if skip {
			continue
		}
		if p.ReturnType == "internal" {
			continue
		}

		funcs = append(funcs, p.Func())
	}
	return funcs, rows.Err()
}

func run(ctx context.Context) error {
	tmpl, err := template.New("").Parse(catalogTmpl)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Generate internal/engine/postgresql/pg_catalog.gen.go
	rows, err := conn.Query(ctx, catalogFuncs)
	if err != nil {
		return err
	}
	funcs, err := scanFuncs(rows)
	if err != nil {
		return err
	}
	out := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(out, tmplCtx{Pkg: "postgresql", Name: "genPGCatalog", Funcs: funcs}); err != nil {
		return err
	}
	code, err := format.Source(out.Bytes())
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join("internal", "engine", "postgresql", "pg_catalog.go"), code, 0644)
	if err != nil {
		return err
	}

	loaded := []extensionPair{}

	for _, extension := range extensions {
		name := strings.Replace(extension, "-", "_", -1)

		var funcName string
		for _, part := range strings.Split(name, "_") {
			funcName += strings.Title(part)
		}

		_, err := conn.Exec(ctx, fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", extension))
		if err != nil {
			log.Printf("error creating %s: %s", extension, err)
			continue
		}

		rows, err := conn.Query(ctx, extensionFuncs, extension)
		if err != nil {
			return err
		}
		funcs, err := scanFuncs(rows)
		if err != nil {
			return err
		}
		if len(funcs) == 0 {
			log.Printf("no functions in %s, skipping", extension)
			continue
		}
		out := bytes.NewBuffer([]byte{})
		if err := tmpl.Execute(out, tmplCtx{Pkg: "contrib", Name: funcName, Funcs: funcs}); err != nil {
			return err
		}
		code, err := format.Source(out.Bytes())
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join("internal", "engine", "postgresql", "contrib", name+".go"), code, 0644)
		if err != nil {
			return err
		}

		loaded = append(loaded, extensionPair{Name: extension, Func: funcName})
	}

	{
		tmpl, err := template.New("").Parse(loaderFuncTmpl)
		if err != nil {
			return err
		}
		out := bytes.NewBuffer([]byte{})
		if err := tmpl.Execute(out, loaded); err != nil {
			return err
		}
		code, err := format.Source(out.Bytes())
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join("internal", "engine", "postgresql", "extension.go"), code, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

type extensionPair struct {
	Name string
	Func string
}

// https://www.postgresql.org/docs/current/contrib.html
var extensions = []string{
	"adminpack",
	"amcheck",
	"auth_delay",
	"auto_explain",
	"bloom",
	"btree_gin",
	"btree_gist",
	"citext",
	"cube",
	"dblink",
	"dict_int",
	"dict_xsyn",
	"earthdistance",
	"file_fdw",
	"fuzzystrmatch",
	"hstore",
	"intagg",
	"intarray",
	"isn",
	"lo",
	"ltree",
	"pageinspect",
	"passwordcheck",
	"pg_buffercache",
	"pgcrypto",
	"pg_freespacemap",
	"pg_prewarm",
	"pgrowlocks",
	"pg_stat_statements",
	"pgstattuple",
	"pg_trgm",
	"pg_visibility",
	"postgres_fdw",
	"seg",
	"sepgsql",
	"spi",
	"sslinfo",
	"tablefunc",
	"tcn",
	"test_decoding",
	"tsm_system_rows",
	"tsm_system_time",
	"unaccent",
	"uuid-ossp",
	"xml2",
}
