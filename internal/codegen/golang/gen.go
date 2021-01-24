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
	Table          core.FQN
	ProjectPath    string
	Name           string
	IDTypeUpper    string
	TableName      string
	Fields         []Field
	Comment        string
	IDExists       bool
	IDType         string
	ImportList     map[string]string
	StructNameList []string
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
		StructNameList                                                                []string
		targetApiPath, targetRootApi, targetRepoPath, targetRootPath, targetModelpath string
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

	//---------------------- init delivery folder -----------------//
	//prepare repository folder
	targetApiPath, targetRootApi, err = makeMultiDirectoryIfNotExists(golang.Out, map[int]string{
		0: "delivery",
		1: "api",
	})
	if err != nil {
		fmt.Printf("failed create delivery folder, err : %+v \n", err)
	}
	//prepare repo util folder
	err = makeDirIfNotExists(targetRootApi + "/util")
	if err != nil {
		fmt.Printf("failed create target delivery util folder, err : %+v \n", err)
	}
	fmt.Println("done create delivery folder")
	//---------------------- eof init delivery folder -----------------//

	//---------------------- init repo folder -----------------//
	//prepare repository folder
	targetRepoPath, targetRootPath, err = makeMultiDirectoryIfNotExists(golang.Out, map[int]string{
		0: "repository",
		1: "postgre",
	})
	if err != nil {
		fmt.Printf("failed create repo folder, err : %+v \n", err)
	}
	//prepare repo util folder
	err = makeDirIfNotExists(targetRootPath + "/util")
	if err != nil {
		fmt.Printf("failed create target repo util folder, err : %+v \n", err)
	}
	fmt.Println("done create repo folder")
	//---------------------- eof init repo folder -----------------//

	//---------------------- init usecase folder -----------------//
	//prepare usecase folder
	err = makeDirIfNotExists(golang.Out + "/usecase")
	if err != nil {
		fmt.Printf("failed create target usecase folder, err : %+v \n", err)
	}
	//prepare usecase util folder
	err = makeDirIfNotExists(golang.Out + "/usecase/util")
	if err != nil {
		fmt.Printf("failed create target usecase util folder, err : %+v \n", err)
	}
	fmt.Println("done create usecase folder")
	//---------------------- eof init usecase folder -----------------//

	//---------------------- init model folder -----------------//
	targetModelpath, _, err = makeMultiDirectoryIfNotExists(golang.Out, map[int]string{
		0: "model",
		1: "types",
	})

	if err != nil {
		fmt.Printf("failed create model folder, err : %+v \n\n", err)
	}
	fmt.Println("done create model folder")
	//---------------------- eof init model folder -----------------//

	tmpl := template.Must(template.New("table").Funcs(funcMap).Parse(templateEsmart))
	//---------------------- init file generator -----------------//
	for _, v := range structs {
		StructNameList = append(StructNameList, v.Name)
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
			IDExists:            v.IDExists,
			IDType:              v.IDType,
			IDTypeUpper:         strings.Title(v.IDType),
			ImportList:          v.ImportList,
			TableName:           v.TableName,
		}

		//delivery API file generator
		executeAPI := func(name, templateName string) error {
			targetPath := targetApiPath + "/" + v.Name
			//prepare api sub folder
			err = makeDirIfNotExists(targetPath)
			if err != nil {
				fmt.Printf("failed create target sub api folder, err : %+v \n", err)
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
				FilePath: targetPath,
				FileBody: string(code),
				FileName: name,
			})
			return nil
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
			} else {
				targetSubpath = targetRootPath
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
		executeModel := func(name, templateName string) error {
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
				FilePath: targetModelpath,
				FileBody: string(code),
				FileName: name,
			})
			return nil
		}
		//usecase file generator
		executeUsecase := func(subPath bool, name, templateName string) error {
			targetPath := golang.Out + "/usecase"
			if subPath {
				//prepare sub repo folder if not exist
				targetPath = targetPath + "/" + v.Name
				err = makeDirIfNotExists(targetPath)
				if err != nil {
					fmt.Printf("failed create target sub usecase folder, err : %+v \n\n", err)
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
				FilePath: targetPath,
				FileBody: string(code),
				FileName: name,
			})
			return nil
		}

		//generate delivery API file
		if err := executeAPI("router.go", "deliveryRouter"); err != nil {
			return nil, err
		}
		if err := executeAPI("handler.go", "deliveryHandler"); err != nil {
			return nil, err
		}

		//generate repository file
		if err := executeRepo(false, v.Name+"Repository.go", "repoInterfaceFile"); err != nil {
			return nil, err
		}
		if err := executeRepo(true, v.Name+"RepoImpl.go", "repoImplFile"); err != nil {
			return nil, err
		}

		//generate usecase file
		if err := executeUsecase(false, v.Name+"UseCase.go", "usecaseInterface"); err != nil {
			return nil, err
		}
		if err := executeUsecase(true, v.Name+"CaseImpl.go", "usecaseImpl"); err != nil {
			return nil, err
		}

		//generate model file
		if err := executeModel(v.Name+"Entity.go", "entityFile"); err != nil {
			return nil, err
		}
	}
	//---------------------- eof init file generator -----------------//

	//generate util file
	utilCtx := tmplCtx{
		Settings:            settings.Global,
		EmitInterface:       golang.EmitInterface,
		EmitJSONTags:        golang.EmitJSONTags,
		EmitDBTags:          golang.EmitDBTags,
		EmitFormTags:        golang.EmitFormTags,
		EmitPreparedQueries: golang.EmitPreparedQueries,
		EmitEmptySlices:     golang.EmitEmptySlices,
		Q:                   "`",
		Package:             golang.Package,
		ProjectPath:         golang.ProjectPath,
		GoQueries:           queries,
		Enums:               enums,
		StructNameList:      StructNameList,
	}
	//util file generator
	executeUtilFile := func(utilPath, name, templateName string) error {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		utilCtx.SourceName = name
		err := tmpl.ExecuteTemplate(w, templateName, &utilCtx)
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
			FilePath: utilPath,
			FileBody: string(code),
			FileName: name,
		})
		return nil
	}

	//generate repo util file
	if err := executeUtilFile(golang.Out+"/repository/util", "init.go", "repoUtil"); err != nil {
		return nil, err
	}
	//generate usecase util file
	if err := executeUtilFile(golang.Out+"/usecase/util", "init.go", "usecaseUtil"); err != nil {
		return nil, err
	}

	return
}

func makeMultiDirectoryIfNotExists(rootPath string, pathList map[int]string) (targetPath, targetRootPath string, err error) {
	//var targetpath string
	targetPath, err = os.Getwd()
	if err != nil {
		return
	}
	targetPath = targetPath + "/" + rootPath
	for k, v := range pathList {
		targetPath = targetPath + "/" + v
		if k == 0 {
			targetRootPath = targetPath
		}
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
