package golang

import "fmt"

var (
	templateUsecaseInterface = `
{{define "usecaseInterface"}}
package usecase

import (
	types "{{.ProjectPath}}/internal/model/types/{{.Name}}"
	"context"
)

type {{.Name}}UseCase interface {
	Submit(ctx context.Context, data types.{{.Name}}Entity) error
	SubmitMultiple(ctx context.Context, data []*types.{{.Name}}Entity) error
	UpdateByPK(ctx context.Context, data types.{{.Name}}Entity) error
	DeleteByPK(ctx context.Context, data types.{{.Name}}Entity) error
	GetList(ctx context.Context, start, limit int) (data []types.{{.Name}}Entity, count int, err error)
{{- if .IDExists}}
	GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, err error)
{{- end}}
}
{{end}}`

	templateUseCaseImpl = `
{{define "usecaseImpl"}}
package {{.Name}}

import (
	"{{.ProjectPath}}/internal/model/core"
	types "{{.ProjectPath}}/internal/model/types/{{.Name}}"
	"{{.ProjectPath}}/internal/repository"
	"{{.ProjectPath}}/internal/repository/util"
	"context"
	"errors"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

type {{.Name}}Case struct {
	Cfg         *core.Config
	Repo{{.Name}} repository.{{.Name}}Repository
}

func New(cfg *core.Config, repoWrapper *util.RepositoryWrapper) *{{.Name}}Case {
	return &{{.Name}}Case{
		Cfg:         cfg,
		Repo{{.Name}}: repoWrapper.Repo{{.Name}},
	}
}

func (c *{{.Name}}Case) Submit(ctx context.Context, data types.{{.Name}}sEntity) (err error){
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "usecase.Submit")
		defer span.Finish()
	}
	err = c.Repo{{.Name}}.Submit(ctx, data)
	if err != nil {
		log.Errorf("[{{.Name}}CaseImpl][Submit] error Submit : %+v", err)
	}
	return
}
func (c *{{.Name}}Case) SubmitMultiple(ctx context.Context, data []*types.{{.Name}}sEntity) (err error){
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "usecase.SubmitMultiple")
		defer span.Finish()
	}
	if len(data) > 0 {
		err = c.Repo{{.Name}}.SubmitMultiple(ctx, data)
		if err != nil {
			log.Errorf("[{{.Name}}CaseImpl][SubmitMultiple] error SubmitMultiple : %+v", err)
		}
	}else{
		err = errors.New("data input is empty")
	}
	return
}
func (c *{{.Name}}Case) UpdateByPK(ctx context.Context, data types.{{.Name}}sEntity) (err error){
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "usecase.UpdateByPK")
		defer span.Finish()
	}
	err = c.Repo{{.Name}}.UpdateByPK(ctx, data)
	if err != nil {
		log.Errorf("[{{.Name}}CaseImpl][UpdateByPK] error UpdateByPK : %+v", err)
	}
	return
}
func (c *{{.Name}}Case) DeleteByPK(ctx context.Context, data types.{{.Name}}sEntity) (err error){
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "usecase.DeleteByPK")
		defer span.Finish()
	}
	err = c.Repo{{.Name}}.DeleteByPK(ctx, data)
	if err != nil {
		log.Errorf("[{{.Name}}CaseImpl][DeleteByPK] error DeleteByPK : %+v", err)
	}
	return
}
func (c *{{.Name}}Case) GetList(ctx context.Context, start, limit int) (data []types.{{.Name}}sEntity, count int, err error){
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "usecase.GetList")
		defer span.Finish()
	}
	if start >= 0 && limit >=1 {
		data, count, err = c.Repo{{.Name}}.GetList(ctx, start, limit)
		if err != nil {
			log.Errorf("[{{.Name}}CaseImpl][GetList] error GetList : %+v", err)
		}
	}else{
		err = errors.New("invalid offset and page limit")
	}
	return
}
{{- if .IDExists}}
func (c *{{.Name}}Case) GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}sEntity, err error){
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "usecase.GetByID")
		defer span.Finish()
	}
	data, err = c.Repo{{.Name}}.GetByID(ctx, id)
	if err != nil {
		log.Errorf("[{{.Name}}CaseImpl][GetByID] error GetByID : %+v", err)
	}
	return
}
{{- end}}
{{end}}`

	templateRepoUtil = `
{{define "repoUtil"}}
package util

import (
	"{{.ProjectPath}}/internal/model/core"
	"{{.ProjectPath}}/internal/repository"
	"{{.ProjectPath}}/internal/repository/postgre"
	{{range .StructNameList}}
	{{.}}Repo "/internal/repository/postgre/{{.}}"
	{{end}}	
)

type RepositoryWrapper struct {
	{{range .StructNameList}}
	Repo{{.}} repository.{{.}}Repository
	{{end}}	
}

func InitRepo(cfg *core.Config) *RepositoryWrapper {
	var (
		MasterDB, SlaveDB *pg.DB
	)
	dbWrapper := postgre.InitPostgre(cfg)
	//add init master & slave db here based on config file
	
	//eof master & slave db init
	return &RepositoryWrapper{
		{{range .StructNameList}}
		Repo{{.}}: {{.}}Repo.New{{.}}Wrapper(MasterDB, SlaveDB),
		{{end}}	
	}
}
{{end}}
`

	templateModel = `
{{define "entityFile"}}
package types

import (
	{{range $i, $impList :=  .ImportList}}
	"{{$impList}}"
	{{end}}
)

type {{.Name}} struct {
	tableName struct{} {{$.Q}}pg:{{.Name}},alias:_{{.Name}}{{$.Q}}
	{{- range .Fields}}
	{{- if .Comment}}
	{{comment .Comment}}{{else}}
	{{- end}}
	{{.Name}} {{.Type}} {{$.Q}}{{.Tag}}{{$.Q}}
	{{- end}}
}
{{end}}

{{define "payloadFile"}}
package types

import (
	{{range imports .SourceName}}
	{{range .}}{{.}}
	{{end}}
	{{end}}
)

type {{.Name}}Payload struct {
	{{- range .Fields}}
	{{- if .Comment}}
	{{comment .Comment}}{{else}}
	{{- end}}
	{{.Name}} {{.Type}} {{$.Q}}{{.Tag}}{{$.Q}} 
	{{- end}}
}
{{end}}`
	templatePostgre = `
{{define "repoInterfaceFile"}}
package repository

import (
	types "{{.ProjectPath}}/internal/model/types/{{.Name}}"
	"context"
)

type {{.Name}}Repository interface {
	Submit(ctx context.Context, data types.{{.Name}}Entity) error
	SubmitMultiple(ctx context.Context, data []*types.{{.Name}}Entity) error
	UpdateByPK(ctx context.Context, data types.{{.Name}}Entity) error
	DeleteByPK(ctx context.Context, data types.{{.Name}}Entity) error
	GetList(ctx context.Context, start, limit int) (data []types.{{.Name}}Entity, count int, err error)
{{- if .IDExists}}
	GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, err error)
{{end}}
}
{{end}}

{{define "repoImplFile"}}
package postgre

import (
	types "{{.ProjectPath}}/internal/model/types/{{.Name}}"
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

type {{.Name}}Data struct {
	MasterDB *pg.DB
	SlaveDB  *pg.DB
}

func New{{.Name}}Wrapper(master, slave *pg.DB) *{{.Name}}Data {
	return &{{.Name}}Data{
		MasterDB: master,
		SlaveDB:  slave,
	}
}

func (d *{{.Name}}Data) Submit(ctx context.Context, data types.{{.Name}}Entity) (err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "repo.Submit")
		defer span.Finish()
	}
	_, err = d.MasterDB.ModelContext(ctx, &data).Insert()
	if err != nil {
		log.Errorf("[{{.Name}}RepoImpl][Submit] error Submit : %+v", err)
	}
	return
}

func (d *{{.Name}}Data) SubmitMultiple(ctx context.Context, data []*types.{{.Name}}Entity) (err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "repo.SubmitMultiple")
		defer span.Finish()
	}
	_, err = d.MasterDB.ModelContext(ctx, &data).Insert()
	if err != nil {
		log.Errorf("[{{.Name}}RepoImpl][SubmitMultiple] error SubmitMultiple : %+v", err)
	}
	return
}

{{- if .IDExists}}
func (d *{{.Name}}Data) GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "repo.GetByID")
		defer span.Finish()
	}
	err = d.SlaveDB.ModelContext(ctx, &data).Where("id = ?", id).Select()
	if err != nil {
		log.Errorf("[{{.Name}}RepoImpl][GetByID] error GetByID : %+v", err)
	}
	return
}
{{- end}}

func (d *{{.Name}}Data) GetList(ctx context.Context, offset, limit int) (data []types.{{.Name}}Entity, count int, err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "repo.GetList")
		defer span.Finish()
	}
	count, err = d.SlaveDB.ModelContext(ctx, &data).Offset(offset).Limit(limit).SelectAndCount()
	if err != nil {
		log.Errorf("[{{.Name}}RepoImpl][GetList] error GetList : %+v", err)
	}
	return
}

func (d *{{.Name}}Data) UpdateByPK(ctx context.Context, data types.{{.Name}}Entity) (err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "repo.UpdateByPK")
		defer span.Finish()
	}
	_, err = d.MasterDB.ModelContext(ctx, &data).WherePK().Update()
	if err != nil {
		log.Errorf("[{{.Name}}RepoImpl][UpdateByPK] error UpdateByPK : %+v", err)
	}
	return
}

func (d *{{.Name}}Data) DeleteByPK(ctx context.Context, data types.{{.Name}}Entity) (err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "repo.DeleteByPK")
		defer span.Finish()
	}
	_, err = d.MasterDB.ModelContext(ctx, &data).WherePK().Delete()
	if err != nil {
		log.Errorf("[{{.Name}}RepoImpl][DeleteByPK] error DeleteByPK : %+v", err)
	}
	return
}

{{end}}`

	templateEsmart = fmt.Sprintf("%s %s %s %s %s", templateRepoUtil, templateModel, templatePostgre, templateUsecaseInterface, templateUseCaseImpl)
)
