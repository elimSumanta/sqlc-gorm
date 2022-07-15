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
	"{{.ProjectPath}}/internal/model/config"
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

	templateApiInit = `
{{define "deliveryUtil"}}
package delivery

import (	
	{{range .StructNameList}}{{.}}API "{{$.ProjectPath}}/internal/delivery/api/{{.}}" 
	{{end}}	

	"{{.ProjectPath}}/internal/delivery/api/helper"
	"{{.ProjectPath}}/internal/model/config"
	"{{.ProjectPath}}/internal/usecase/util"
	"github.com/kataras/iris/v12"
)

func InitDeliveryAPI(cfg *config.Config, app *iris.Application, caseWrapper *util.UseCaseWrapper) {
	middleware := helper.NewMiddleware(cfg, app)
	{{range .StructNameList}}{{.}}API.InitRoute(cfg, middleware, caseWrapper)
	{{end}}
}
{{end}}`

	templateUsecaseInit = `
{{define "usecaseUtil"}}
package util

import (
	"{{.ProjectPath}}/internal/model/config"
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

	templateUsecaseInterface = `
{{define "usecaseInterface"}}
package usecase

import (
	"{{.ProjectPath}}/internal/model/types"
	"context"
)

type {{.Name}}UseCase interface {
	Submit(ctx context.Context, data types.{{.Name}}Entity) error
	SubmitMultiple(ctx context.Context, data []types.{{.Name}}Entity) error
	GetListWithFilter(ctx context.Context, filter types.DataTableCommonFilter) (types.DataTableCommonResponse, error)
{{- if .IDExists}}
	GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, found bool, err error)
	UpdateByID(ctx context.Context, data types.{{.Name}}Entity) (err error)
{{- end}}
}
{{end}}`

	templateUseCaseImpl = `
{{define "usecaseImpl"}}
package {{.Name}}

import (
	"{{.ProjectPath}}/internal/model/config"
	"{{.ProjectPath}}/internal/model/types"
	"{{.ProjectPath}}/internal/repository"
	"{{.ProjectPath}}/internal/repository/util"
	"context"
	"errors"
	"{{.ProjectPath}}/pkg/lib"
	"github.com/rs/zerolog/log"
)

type {{.Name}}Case struct {
	cfg         *config.Config
	repo{{.Name}} repository.{{.Name}}Repository
}

func New(cfg *config.Config, repoWrapper *util.RepositoryWrapper) *{{.Name}}Case {
	return &{{.Name}}Case{
		cfg:         cfg,
		repo{{.Name}}: repoWrapper.Repo{{.Name}},
	}
}

func (c *{{.Name}}Case) Submit(ctx context.Context, data types.{{.Name}}Entity) (err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.Submit")
	defer span.End()
	defer span.RecordError(err)

	err = c.Repo{{.Name}}.Submit(ctx, data)
	if err != nil {
		log.Error().Err(err)
	}
	return
}

func (c *{{.Name}}Case) SubmitMultiple(ctx context.Context, data []types.{{.Name}}Entity) (err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.SubmitMultiple")
	defer span.End()
	defer span.RecordError(err)

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

func (c *{{.Name}}Case) GetListWithFilter(ctx context.Context, filter types.DataTableCommonFilter) (types.DataTableCommonResponse, error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.GetListWithFilter")
	defer span.End()
	defer span.RecordError(err)
	var tableInfo types.DataTableAttribute

	data.Data, tableInfo, err = c.userSecurityRepo.GetListByFilter(ctx, filter)
	lib.GetDataTableVal(filter, &tableInfo)
	data.ItemCount = tableInfo.ItemCount
	data.TotalPages = tableInfo.TotalPages
	data.First = tableInfo.First
	data.Last = tableInfo.Last
	data.TotalCount = tableInfo.TotalCount
	
	return
}

{{- if .IDExists}}

func (c *{{.Name}}Case) GetByID(ctx context.Context, id {{.IDType}}) (data types.{{.Name}}Entity, found bool, err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.GetByID")
	defer span.End()
	defer span.RecordError(err)

	data, found, err = c.Repo{{.Name}}.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err)
	}
	return
}

func (c *{{.Name}}Case) UpdateByID(ctx context.Context, data types.{{.Name}}Entity) (err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.UpdateByID")
	defer span.End()
	defer span.RecordError(err)

	data, found, err = c.Repo{{.Name}}.UpdateByID(ctx, data)
	if err != nil {
		log.Error().Err(err)
	}
	return
}

{{- end}}
{{end}}`

	templateRepoUtil = `
{{define "repoUtil"}}
package util

import (
	"{{.ProjectPath}}/internal/model/config"
	"{{.ProjectPath}}/internal/repository"
	"{{.ProjectPath}}/internal/repository/postgre"
	{{range .StructNameList}}{{.}}Repo "{{$.ProjectPath}}/internal/repository/postgre/{{.}}"
	{{end}}	
)

type RepositoryWrapper struct {
	{{range .StructNameList}}Repo{{.}} repository.{{.}}Repository
	{{end}}	
}

func InitRepo(cfg *config.Config) *RepositoryWrapper {
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
	templatePostgre = `
{{define "repoInterfaceFile"}}
package repository

import (
	"{{.ProjectPath}}/internal/model/types"
	"context"
)

type {{.Name}}Repository interface {
	Submit(ctx context.Context, data types.{{.Name}}Entity) error
	SubmitMultiple(ctx context.Context, data []types.{{.Name}}Entity) error		
	GetListByFilter(ctx context.Context, filter types.DataTableCommonFilter) (data []types.{{.Name}}Entity, info types.DataTableAttribute, err error)
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

func (d *{{.Name}}Data) GetListByFilter(ctx context.Context, filter types.DataTableCommonFilter) (data []types.{{.Name}}Entity, info types.DataTableAttribute, err error) {
	ctx, span := lib.StartRepositorySpan(ctx, "{{.Name}}Data.GetListByFilter")
	defer span.End()
	defer span.RecordError(err)

	var modelData types.{{.Name}}Entity
	rs := d.SlaveDB.WithContext(ctx).Model(&modelData).Count(&info.TotalCount)
	err = rs.Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}

	rs = d.SlaveDB.WithContext(ctx).Offset(filter.Offset).Limit(filter.Limit).Find(&data)
	info.ItemCount = rs.RowsAffected
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
