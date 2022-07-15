package golang

import "fmt"

var (
	templateApiHandlerGorm = `
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
		log.Error().Err(err)
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
		log.Error().Err(err)
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
		log.Error().Err(err)
		return
	}
	err = h.CaseWrapper.{{.Name}}UseCase.Submit(cx, formPayload)
	if err != nil {
		log.Error().Err(err)
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
		log.Error().Err(err)
		return
	}
	err = h.CaseWrapper.{{.Name}}UseCase.UpdateByPK(cx, formPayload)
	if err != nil {
		log.Error().Err(err)
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
		log.Error().Err(err)
		return
	}
	err = h.CaseWrapper.{{.Name}}UseCase.DeleteByPK(cx, formPayload)
	if err != nil {
		log.Error().Err(err)
		return
	}
	data.ResponseBody = "{{.Name}} data has been deleted"
	return
}
{{end}}`

	templateApiRouterGorm = `
{{define "deliveryRouter"}}
package {{.Name}}

import (
	"{{.ProjectPath}}/internal/delivery/api/helper"
	"{{.ProjectPath}}/internal/model/core"
	"{{.ProjectPath}}/internal/usecase/util"
)

type ApiWrapper struct {
	CaseWrapper *util.UseCaseWrapper
	Config      *config.Config
}

func InitRoute(cfg *config.Config, middleware *helper.Middleware, caseWrapper *util.UseCaseWrapper) {
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

	templateApiInitGorm = `
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

func InitDeliveryAPI(cfg *config.Config, app *iris.Application, caseWrapper *util.UseCaseWrapper) {
	middleware := helper.NewMiddleware(cfg, app)
	{{range .StructNameList}}{{.}}API.InitRoute(cfg, middleware, caseWrapper)
	{{end}}
}
{{end}}`

	templateUsecaseInitGorm = `
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

func InitUsecase(cfg *config.Config, repoWrapper *repoUtil.RepositoryWrapper) (wrapper *UseCaseWrapper, err error) {
	return &UseCaseWrapper{
		{{range .StructNameList}}{{.}}UseCase: {{.}}UCase.New(cfg, repoWrapper),
		{{end}}	
	}, nil
}
{{end}}`

	templateUsecaseInterfaceGorm = `
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

	templateUseCaseImplGorm = `
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
	Cfg         *config.Config
	Repo{{.Name}} repository.{{.Name}}Repository
}

func New(cfg *config.Config, repoWrapper *util.RepositoryWrapper) *{{.Name}}Case {
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
		log.Error().Err(err)
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
			log.Error().Err(err)
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
		log.Error().Err(err)
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
		log.Error().Err(err)
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
			log.Error().Err(err)
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
		log.Error().Err(err)
	}
	return
}
{{- end}}
{{end}}`

	templateRepoUtilGorm = `
{{define "repoUtil"}}
package util

import (
	"github.com/rs/zerolog/log"
	"{{.ProjectPath}}/internal/model/config"
	"{{.ProjectPath}}/internal/repository/postgre"
	"{{.ProjectPath}}/internal/repository/redis"
	"{{.ProjectPath}}/internal/repository"
	"{{.ProjectPath}}/pkg/lib"
	{{range .StructNameList}}{{.}}Repo "{{$.ProjectPath}}/internal/repository/postgre/{{.}}"
	{{end}}		
)

type RepositoryWrapper struct {
	{{range .StructNameList}}Repo{{.}} repository.{{.}}Repository
	{{end}}	
}

func NewRepository(ctx context.Context, cfg *config.Config) (*RepositoryWrapper, error) {
	ctx, span := lib.StartInitSpan(ctx, "NewRepository")
	defer span.End()

	// init redis repo
	redisRepo := redis.NewRedisRepo(ctx, cfg)

	// init postgre-sql
	DBWrapper, err := postgre.InitDB(ctx, cfg)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	//add init master & slave db here based on config file
	
	//eof master & slave db init
	return &RepositoryWrapper{
		{{range .StructNameList}}Repo{{.}}: {{.}}Repo.New{{.}}Wrapper(DBWrapper.MasterDB, DBWrapper.SlaveDB),
		{{end}}	
	}
}
{{end}}
`

	templateModelGorm = `
{{define "entityFile"}}
package types

import (
	{{range $i, $impList :=  .ImportList}}"{{$impList}}"
	{{end}}
)

type {{.Name}}Entity struct {	
	{{- range .Fields}}
	{{- if .Comment}}
	{{comment .Comment}}{{else}}
	{{- end}}
	{{.Name}} {{.Type}} {{$.Q}}{{.Tag}}{{$.Q}}
	{{- end}}
}

type Tabler interface {
  TableName() string
}

func ({{.Name}}Entity) TableName() string {
  return "{{.TableName}}"
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
	templatePostgreGorm = `
{{define "repoInterfaceFile"}}
package repository

import (
	"{{.ProjectPath}}/internal/model/types"
	"context"
)

type {{.Name}}Repository interface {
	Submit(ctx context.Context, data types.{{.Name}}Entity) error
	SubmitMultiple(ctx context.Context, data []types.{{.Name}}Entity) error	
	GetList(ctx context.Context, offset, limit int) (data []types.{{.Name}}Entity, count int64, err error)
	GetAll(ctx context.Context) (data []types.{{.Name}}Entity, count int64, err error)
{{- if .IDExists}}
	GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, found bool, err error)
	UpdateByID(ctx context.Context, data types.{{.Name}}Entity) error
{{- end}}
}
{{end}}

{{define "repoImplFile"}}
package postgre

import (
	"context"
	"github.com/rs/zerolog/log"
	types "{{.ProjectPath}}/internal/model/types"
	"{{.ProjectPath}}/pkg/lib"
	"gorm.io/gorm"
)

type {{.Name}}Data struct {
	MasterDB *gorm.DB
	SlaveDB  *gorm.DB
}

func New{{.Name}}Wrapper(master, slave *gorm.DB) *{{.Name}}Data {
	return &{{.Name}}Data{
		MasterDB: master,
		SlaveDB:  slave,
	}
}

func (d *{{.Name}}Data) Submit(ctx context.Context, data types.{{.Name}}Entity) (err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.Submit")
	defer span.End()
	defer span.RecordError(err)
	
	tx := d.MasterDB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
		if err != nil {
			log.Error().Err(err)
		}
	}()

	if err = tx.Error; err != nil {
		return 
	}

	if err = tx.Create(&data).Error; err != nil {
		tx.Rollback()
		return 
	}

	err = tx.Commit().Error
	return
}

func (d *{{.Name}}Data) SubmitMultiple(ctx context.Context, data []types.{{.Name}}Entity) (err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.SubmitMultiple")
	defer span.End()
	defer span.RecordError(err)
	
	tx := d.MasterDB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
		if err != nil {
			log.Error().Err(err)
		}
	}()

	if err = tx.Error; err != nil {
		return 
	}

	if err = tx.Create(&data).Error; err != nil {
		tx.Rollback()
		return 
	}

	err = tx.Commit().Error
	return
}

{{- if .IDExists}}
func (d *{{.Name}}Data) GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, found bool, err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.GetByID")
	defer span.End()
	defer span.RecordError(err)

	err = d.SlaveDB.WithContext(ctx).Take(&data, "id = ?", ID).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	found = err == nil
	return
}

func (d *{{.Name}}Data) UpdateByID(ctx context.Context, data types.{{.Name}}Entity) (err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.UpdateByID")
	defer span.End()
	defer span.RecordError(err)

	tx := d.MasterDB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
		if err != nil {
			log.Error().Err(err)
		}
	}()

	if err = tx.Error; err != nil {
		return
	}

	if err = tx.Updates(&data).Where("id = ?", data.ID).Error; err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit().Error

	return
}
{{- end}}

func (d *{{.Name}}Data) GetList(ctx context.Context, offset, limit int) (data []types.{{.Name}}Entity, count int64, err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.GetList")
	defer span.End()
	defer span.RecordError(err)

	rs := d.SlaveDB.WithContext(ctx).Offset(offset).Limit(limit).Find(&data)
	count = rs.RowsAffected
	err = rs.Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	
	return
}

func (d *{{.Name}}Data) GetAll(ctx context.Context) (data []types.{{.Name}}Entity, count int64, err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.GetAll")
	defer span.End()
	defer span.RecordError(err)

	rs := d.SlaveDB.WithContext(ctx).Find(&data)
	count = rs.RowsAffected
	err = rs.Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	
	return
}

{{end}}`

	templateOut = fmt.Sprintf(
		"%s %s %s %s %s %s %s %s %s",
		templateRepoUtilGorm,
		templateModelGorm,
		templatePostgreGorm,
		templateUsecaseInterfaceGorm,
		templateUseCaseImplGorm,
		templateUsecaseInitGorm,
		templateApiHandlerGorm,
		templateApiRouterGorm,
		templateApiInitGorm,
	)
)
