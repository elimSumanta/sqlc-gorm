package golang

import "fmt"

var (
	templateApiHandler = `
{{define "deliveryHandler"}}
package {{.Name}}

import (
	"{{.ProjectPath}}/internal/model/types"
	"context"
	"github.com/kataras/iris/v12"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

//GET
func (h *ApiWrapper) GetList(ctx iris.Context) (data *types.JSONResponse, err error) {
	span := opentracing.GlobalTracer().StartSpan("api.GetList")
	defer span.Finish()
	cx := opentracing.ContextWithSpan(context.Background(), span)
	data = &types.JSONResponse{
		ResponseHeader: types.JSONHeader{
			Code:    200,
		},
	}
	var (
		pageListParam = types.PaginationParam{}
		respData = types.ListResponseData{}
	)
	err = ctx.ReadQuery(&pageListParam)
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	respData.Data, respData.Count, err = h.CaseWrapper.{{.Name}}UseCase.GetList(cx, pageListParam.Page, pageListParam.Limit)
	if err != nil {
		log.Errorf("[handler][GetList] error GetList : %+v", err)
	}
	data.ResponseBody = respData
	return
}
{{- if .IDExists}}
//GET
func (h *ApiWrapper) GetByID(ctx iris.Context) (data *types.JSONResponse, err error) {
	span := opentracing.GlobalTracer().StartSpan("api.GetByID")
	defer span.Finish()
	cx := opentracing.ContextWithSpan(context.Background(), span)
	data = &types.JSONResponse{
		ResponseHeader: types.JSONHeader{
			Code:    200,
		},
	}
	var itemID int
	itemID, err = ctx.URLParamInt("id")

	if err != nil {
		return
	}
	data.ResponseBody, err = h.CaseWrapper.{{.Name}}UseCase.GetByID(cx, {{.IDType}}(itemID))
	if err != nil {
		log.Errorf("[handler][GetByID] error GetByID : %+v", err)
	}
	return
}
{{- end}}
//POST
func (h *ApiWrapper) Submit(ctx iris.Context) (data *types.JSONResponse, err error) {
	span := opentracing.GlobalTracer().StartSpan("api.Submit")
	defer span.Finish()
	cx := opentracing.ContextWithSpan(context.Background(), span)
	data = &types.JSONResponse{
		ResponseHeader: types.JSONHeader{
			Code:    200,
		},
	}
	formPayload := types.{{.Name}}Entity{}
	err = ctx.ReadForm(&formPayload)
	if err != nil {
		log.Errorf("[handler][Submit] error parse formPayload : %+v", err)
		return
	}
	err = h.CaseWrapper.{{.Name}}UseCase.Submit(cx, formPayload)
	if err != nil {
		log.Errorf("[handler][Submit] error submit {{.Name}} : %+v", err)
		return
	}
	data.ResponseBody = "{{.Name}} data has been submitted"
	return
}
//PUT
func (h *ApiWrapper) Update(ctx iris.Context) (data *types.JSONResponse, err error) {
	span := opentracing.GlobalTracer().StartSpan("api.Update")
	defer span.Finish()
	cx := opentracing.ContextWithSpan(context.Background(), span)
	data = &types.JSONResponse{
		ResponseHeader: types.JSONHeader{
			Code:    200,
		},
	}
	formPayload := types.{{.Name}}Entity{}
	err = ctx.ReadForm(&formPayload)
	if err != nil {
		log.Errorf("[handler][Update] error parse formPayload : %+v", err)
		return
	}
	err = h.CaseWrapper.{{.Name}}UseCase.UpdateByPK(cx, formPayload)
	if err != nil {
		log.Errorf("[handler][Update] error update {{.Name}} : %+v", err)
		return
	}
	data.ResponseBody = "{{.Name}} data has been updated"
	return
}
//DELETE
func (h *ApiWrapper) Delete(ctx iris.Context) (data *types.JSONResponse, err error) {
	span := opentracing.GlobalTracer().StartSpan("api.Delete")
	defer span.Finish()
	cx := opentracing.ContextWithSpan(context.Background(), span)
	data = &types.JSONResponse{
		ResponseHeader: types.JSONHeader{
			Code:    200,
		},
	}
	formPayload := types.{{.Name}}Entity{}
	err = ctx.ReadForm(&formPayload)
	if err != nil {
		log.Errorf("[handler][Delete] error parse formPayload : %+v", err)
		return
	}
	err = h.CaseWrapper.{{.Name}}UseCase.DeleteByPK(cx, formPayload)
	if err != nil {
		log.Errorf("[handler][Delete] error delete {{.Name}} : %+v", err)
		return
	}
	data.ResponseBody = "{{.Name}} data has been deleted"
	return
}
{{end}}`

	templateApiRouter = `
{{define "deliveryRouter"}}
package {{.Name}}

import (
	"{{.ProjectPath}}/internal/delivery/api/helper"
	"{{.ProjectPath}}/internal/model/core"
	"{{.ProjectPath}}/internal/usecase/util"
)

type ApiWrapper struct {
	CaseWrapper *util.UseCaseWrapper
	Config      *core.Config
}

func InitRoute(cfg *core.Config, middleware *helper.Middleware, caseWrapper *util.UseCaseWrapper) {
	api := ApiWrapper{
		CaseWrapper: caseWrapper,
		Config:      cfg,
	}
	api.registerRouter(middleware)
}

func (h *ApiWrapper) registerRouter(m *helper.Middleware) {
	m.POST("/{{.TableName}}/submit", h.Submit)
	m.PUT("/{{.TableName}}/update", h.Update)
	m.DELETE("/{{.TableName}}/delete", h.Delete)
	//base_url/{{.TableName}}/list?page=0&limit=10
	m.GET("/{{.TableName}}/list", h.GetList)
{{- if .IDExists}}
	//base_url/{{.TableName}}/{item_ID}
	m.GET("/{{.TableName}}/{id}", h.GetByID)
{{- end}}
}
{{end}}`

	templateApiInit = `
{{define "deliveryUtil"}}
package delivery

import (	
	{{range .StructNameList}}{{.}}API "{{$.ProjectPath}}/internal/delivery/api/{{.}}" 
	{{end}}	

	"{{.ProjectPath}}/internal/delivery/api/helper"
	"{{.ProjectPath}}/internal/model/core"
	"{{.ProjectPath}}/internal/usecase/util"
	"github.com/kataras/iris/v12"
)

func InitDeliveryAPI(cfg *core.Config, app *iris.Application, caseWrapper *util.UseCaseWrapper) {
	middleware := helper.NewMiddleware(cfg, app)
	{{range .StructNameList}}{{.}}API.InitRoute(cfg, middleware, caseWrapper)
	{{end}}
}
{{end}}`

	templateUsecaseInit = `
{{define "usecaseUtil"}}
package util

import (
	"{{.ProjectPath}}/internal/model/core"
	repoUtil "{{.ProjectPath}}/internal/repository/util"
	"{{.ProjectPath}}/internal/usecase"
	{{range .StructNameList}}{{.}}UCase "{{$.ProjectPath}}/internal/usecase/{{.}}"
	{{end}}	
)

type UseCaseWrapper struct {
	{{range .StructNameList}}{{.}}UseCase usecase.{{.}}UseCase
	{{end}}	
}

func InitUsecase(cfg *core.Config, repoWrapper *repoUtil.RepositoryWrapper) (wrapper *UseCaseWrapper, err error) {
	return &UseCaseWrapper{
		{{range .StructNameList}}{{.}}UseCase: {{.}}UCase.New(cfg, repoWrapper),
		{{end}}	
	}, nil
}
{{end}}`

	templateUsecaseInterface = `
{{define "usecaseInterface"}}
package usecase

import (
	"{{.ProjectPath}}/internal/model/types"
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
	"{{.ProjectPath}}/internal/model/types"
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

func (c *{{.Name}}Case) Submit(ctx context.Context, data types.{{.Name}}Entity) (err error){
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
func (c *{{.Name}}Case) SubmitMultiple(ctx context.Context, data []*types.{{.Name}}Entity) (err error){
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
func (c *{{.Name}}Case) UpdateByPK(ctx context.Context, data types.{{.Name}}Entity) (err error){
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
func (c *{{.Name}}Case) DeleteByPK(ctx context.Context, data types.{{.Name}}Entity) (err error){
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
func (c *{{.Name}}Case) GetList(ctx context.Context, start, limit int) (data []types.{{.Name}}Entity, count int, err error){
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
func (c *{{.Name}}Case) GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, err error){
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
	{{range .StructNameList}}{{.}}Repo "{{$.ProjectPath}}/internal/repository/postgre/{{.}}"
	{{end}}	
)

type RepositoryWrapper struct {
	{{range .StructNameList}}Repo{{.}} repository.{{.}}Repository
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
		{{range .StructNameList}}Repo{{.}}: {{.}}Repo.New{{.}}Wrapper(MasterDB, SlaveDB),
		{{end}}	
	}
}
{{end}}
`

	templateModel = `
{{define "entityFile"}}
package types

import (
	{{range $i, $impList :=  .ImportList}}"{{$impList}}"
	{{end}}
)

type {{.Name}}Entity struct {
	tableName struct{} {{$.Q}}pg:"{{.TableName}}" alias:"_{{.Name}}"{{$.Q}}
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
	"{{.ProjectPath}}/internal/model/types"
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
{{- end}}
}
{{end}}

{{define "repoImplFile"}}
package postgre

import (
	"{{.ProjectPath}}/internal/model/types"
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

	templateEsmart = fmt.Sprintf(
		"%s %s %s %s %s %s %s %s %s",
		templateRepoUtil,
		templateModel,
		templatePostgre,
		templateUsecaseInterface,
		templateUseCaseImpl,
		templateUsecaseInit,
		templateApiHandler,
		templateApiRouter,
		templateApiInit,
	)
)
