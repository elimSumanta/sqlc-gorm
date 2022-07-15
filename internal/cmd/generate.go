package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ujunglangit-id/sqlc/internal/codegen/golang"
	"github.com/ujunglangit-id/sqlc/internal/compiler"
	"github.com/ujunglangit-id/sqlc/internal/config"
	"github.com/ujunglangit-id/sqlc/internal/debug"
	"github.com/ujunglangit-id/sqlc/internal/multierr"
	"github.com/ujunglangit-id/sqlc/internal/opts"
)

const errMessageNoVersion = `The configuration file must have a version number.
Set the version to 1 at the top of sqlc.json:

{
  "version": "1"
  ...
}
`

const errMessageUnknownVersion = `The configuration file has an invalid version number.
The only supported version is "1".
`

const errMessageNoPackages = `No packages are configured`

func printFileErr(stderr io.Writer, dir string, fileErr *multierr.FileError) {
	filename := strings.TrimPrefix(fileErr.Filename, dir+"/")
	fmt.Fprintf(stderr, "%s:%d:%d: %s\n", filename, fileErr.Line, fileErr.Column, fileErr.Err)
}

type outPair struct {
	Gen config.SQLGen
	config.SQL
}

func Generate(e Env, dir string, stderr io.Writer) (map[string]string, error) {
	var yamlMissing, jsonMissing bool
	yamlPath := filepath.Join(dir, "sqlc.yaml")
	jsonPath := filepath.Join(dir, "sqlc.json")

	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		yamlMissing = true
	}
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		jsonMissing = true
	}

	if yamlMissing && jsonMissing {
		fmt.Fprintln(stderr, "error parsing sqlc.json: file does not exist")
		return nil, errors.New("config file missing")
	}

	if !yamlMissing && !jsonMissing {
		fmt.Fprintln(stderr, "error: both sqlc.json and sqlc.yaml files present")
		return nil, errors.New("sqlc.json and sqlc.yaml present")
	}

	configPath := yamlPath
	if yamlMissing {
		configPath = jsonPath
	}
	base := filepath.Base(configPath)

	blob, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(stderr, "error parsing %s: file does not exist\n", base)
		return nil, err
	}

	conf, err := config.ParseConfig(bytes.NewReader(blob))
	if err != nil {
		switch err {
		case config.ErrMissingVersion:
			fmt.Fprintf(stderr, errMessageNoVersion)
		case config.ErrUnknownVersion:
			fmt.Fprintf(stderr, errMessageUnknownVersion)
		case config.ErrNoPackages:
			fmt.Fprintf(stderr, errMessageNoPackages)
		}
		fmt.Fprintf(stderr, "error parsing %s: %s\n", base, err)
		return nil, err
	}

	debug, err := opts.DebugFromEnv()
	if err != nil {
		fmt.Fprintf(stderr, "error parsing SQLCDEBUG: %s\n", err)
		return nil, err
	}

	output := map[string]string{}
	errored := false

	var pairs []outPair
	for _, sql := range conf.SQL {
		if sql.Gen.Go != nil {
			pairs = append(pairs, outPair{
				SQL: sql,
				Gen: config.SQLGen{Go: sql.Gen.Go},
			})
		}
		if sql.Gen.Kotlin != nil {
			pairs = append(pairs, outPair{
				SQL: sql,
				Gen: config.SQLGen{Kotlin: sql.Gen.Kotlin},
			})
		}
	}

	for _, sql := range pairs {
		combo := config.Combine(conf, sql.SQL)

		// TODO: This feels like a hack that will bite us later
		joined := make([]string, 0, len(sql.Schema))
		for _, s := range sql.Schema {
			joined = append(joined, filepath.Join(dir, s))
		}
		sql.Schema = joined

		var name string
		parseOpts := opts.Parser{
			Debug: debug,
		}
		name = combo.Go.Package

		result, errored := parse(e, name, dir, sql.SQL, combo, parseOpts, stderr)
		if errored {
			break
		}

		var targetFiles = []*golang.TargetFiles{}
		targetFiles, err = golang.Generate(result, combo)

		if err != nil {
			fmt.Fprintf(stderr, "# package %s\n", name)
			fmt.Fprintf(stderr, "error generating code: %#v\n", err)
			errored = true
			continue
		}
		total := len(targetFiles)
		executed := 0
		for _, fileData := range targetFiles {
			executed++
			filename := filepath.Join(fileData.FilePath, fileData.FileName)
			output[filename] = fileData.FileBody
			fmt.Printf("done generate %d of %d files, : %s\n", executed, total, fileData.FileName)
		}
	}

	if errored {
		return nil, fmt.Errorf("errored")
	}
	return output, nil
}

func parse(e Env, name, dir string, sql config.SQL, combo config.CombinedSettings, parserOpts opts.Parser, stderr io.Writer) (*compiler.Result, bool) {
	c := compiler.NewCompiler(sql, combo)
	if err := c.ParseCatalog(sql.Schema); err != nil {
		fmt.Fprintf(stderr, "# package %s\n", name)
		if parserErr, ok := err.(*multierr.Error); ok {
			for _, fileErr := range parserErr.Errs() {
				printFileErr(stderr, dir, fileErr)
			}
		} else {
			fmt.Fprintf(stderr, "error parsing schema: %s\n", err)
		}
		return nil, true
	}
	if parserOpts.Debug.DumpCatalog {
		debug.Dump(c.Catalog())
	}
	if err := c.ParseQueries(sql.Queries, parserOpts); err != nil {
		fmt.Fprintf(stderr, "# package %s\n", name)
		if parserErr, ok := err.(*multierr.Error); ok {
			for _, fileErr := range parserErr.Errs() {
				printFileErr(stderr, dir, fileErr)
			}
		} else {
			fmt.Fprintf(stderr, "error parsing queries: %s\n", err)
		}
		return nil, true
	}
	return c.Result(), false
}
