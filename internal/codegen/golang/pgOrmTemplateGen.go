package golang

import "fmt"

var (
	templateApiHandler = `
{{define "deliveryHandler"}}
package {{.Name}}

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"{{.ProjectPath}}/internal/model/types"
	"net/http"
)

// get {{.Name}} by filter
func (r *apiRouter) getList(c *gin.Context) {
	var (
		filter types.DataTableCommonFilter
		err    error
		resp   = types.JSONResponse{
			ResponseHeader: types.JSONHeader{
				Code: http.StatusOK,
			},
		}
	)

	if err := c.ShouldBind(&filter); err != nil {
		log.Error().Stack().Err(err)
		resp.ResponseHeader = types.JSONHeader{
			ErrMessage: err,
			Code:       http.StatusBadRequest,
		}
		c.JSON(resp.ResponseHeader.Code, resp)
		return
	}

	resp.ResponseBody, err = r.{{.Name}}Case.GetListWithFilter(c, filter)
	if err != nil {
		resp.ResponseHeader = types.JSONHeader{
			ErrMessage: err,
			Code:       http.StatusInternalServerError,
		}
		c.JSON(resp.ResponseHeader.Code, resp)
		return
	}
	c.JSON(resp.ResponseHeader.Code, resp)
	return
}

{{- if .IDExists}}
// get {{.Name}} by ID
func (r *apiRouter) GetByID(c *gin.Context) { 
	var (
		id types.ParamID{{.IDType}}
		err    error
		resp   = types.JSONResponse{
			ResponseHeader: types.JSONHeader{
				Code: http.StatusOK,
			},
		}
	)

	if err = c.ShouldBindUri(&id); err != nil {
		log.Error().Stack().Err(err)
		resp.ResponseHeader = types.JSONHeader{
			ErrMessage: err,
			Code:       http.StatusBadRequest,
		}
		c.JSON(resp.ResponseHeader.Code, resp)
		return
	}

	resp.ResponseBody, _, err = r.DtsUserListCase.GetByID(c, id.ID)
	if err != nil {
		resp.ResponseHeader = types.JSONHeader{
			ErrMessage: err,
			Code:       http.StatusInternalServerError,
		}
		c.JSON(resp.ResponseHeader.Code, resp)
		return
	}
	c.JSON(resp.ResponseHeader.Code, resp)
	return
}
{{- end}}

// submit new {{.Name}}
func (r *apiRouter) Submit(ctx iris.Context) (data *types.JSONResponse, err error) {
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

func (r *apiRouter) Update(ctx iris.Context) (data *types.JSONResponse, err error) {
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
func (r *apiRouter) Delete(ctx iris.Context) (data *types.JSONResponse, err error) {
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
	"github.com/gin-gonic/gin"
	"{{.ProjectPath}}/internal/model/config"
	"{{.ProjectPath}}/internal/usecase"
	"{{.ProjectPath}}/internal/usecase/util"
)

type apiRouter struct {
	cfg           *config.Config
	{{.Name}}Case usecase.{{.Name}}UseCase
}

func InitRoute(cfg *config.Config, group *gin.RouterGroup, uc *util.UseCaseWrapper) {
	r := apiRouter{
		cfg:           cfg,
		{{.Name}}Case: uc.{{.Name}}UseCase,
	}
	r.Register(group)
}

func (r *apiRouter) Register(rg *gin.RouterGroup) {
	rg.POST("/submit", r.submit)
	rg.POST("/submit_multiple", r.submitMultiple)
	rg.GET("/list", r.getList)
	{{- if .IDExists}}
	rg.GET("/get/:id", r.getByID)
	rg.POST("/update", r.update)
	{{- end}}
}
{{end}}`

	templateApiInit = `
{{define "deliveryUtil"}}
package delivery

import (	
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	{{range .StructNameList}}{{.}}API "{{$.ProjectPath}}/internal/delivery/api/{{.}}" 
	{{end}}
	"{{.ProjectPath}}/internal/model/config"
	"{{.ProjectPath}}/internal/usecase/util"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)


func (d *APIDelivery) initRoutingGroup() {
	{{range .StructNameList}}	
	group{{.}} := d.Router.Group("/{{.}}")
	{{.}}.InitRoute(d.cfg, group{{.}}, d.caseWrapper)
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

	err = c.repo{{.Name}}.Submit(ctx, data)
	if err != nil {
		log.Error().Stack().Err(err)
	}
	return
}

func (c *{{.Name}}Case) SubmitMultiple(ctx context.Context, data []types.{{.Name}}Entity) (err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.SubmitMultiple")
	defer span.End()
	defer span.RecordError(err)

	if len(data) > 0 {
		err = c.repo{{.Name}}.SubmitMultiple(ctx, data)
		if err != nil {
			log.Error().Stack().Err(err)
		}
	}else{
		err = errors.New("data input is empty")
	}
	return
}

func (c *{{.Name}}Case) GetListWithFilter(ctx context.Context, filter types.DataTableCommonFilter) (data types.DataTableCommonResponse, err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.GetListWithFilter")
	defer span.End()
	defer span.RecordError(err)
	var tableInfo types.DataTableAttribute

	data.Data, tableInfo, err = c.repo{{.Name}}.GetListByFilter(ctx, filter)
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

	err = c.repo{{.Name}}.GetByID(ctx, id)
	if err != nil {
		log.Error().Stack().Err(err)
	}
	return
}

func (c *{{.Name}}Case) UpdateByID(ctx context.Context, data types.{{.Name}}Entity) (err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.UpdateByID")
	defer span.End()
	defer span.RecordError(err)

	data, found, err = c.repo{{.Name}}.UpdateByID(ctx, data)
	if err != nil {
		log.Error().Stack().Err(err)
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
	ctx, span := lib.StartInitSpan(ctx, "NewRepository")
	defer span.End()

	// init postgre-sql
	DBWrapper, err := postgre.InitDB(ctx, cfg)
	if err != nil {
		log.Error().Stack().Err(err)
		return nil, err
	}
	
	//eof master & slave db init
	return &RepositoryWrapper{
		{{range .StructNameList}}Repo{{.}}: {{.}}Repo.New{{.}}Wrapper(DBWrapper.MasterDB, DBWrapper.SlaveDB),
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
	{{- if .NoCreatedAt}}
	CreatedAt time.Time {{$.Q}}form:"-" gorm:"autoCreateTime:false" json:"-"{{$.Q}}
	{{- end}}
	{{- if .NoUpdatedAt}}
	UpdatedAt time.Time {{$.Q}}form:"-" gorm:"autoCreateTime:false" json:"-"{{$.Q}}
	{{- end}}
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
			log.Error().Stack().Err(err)
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
			log.Error().Stack().Err(err)
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

	err = d.SlaveDB.WithContext(ctx).Take(&data, "id = ?", id).Error
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
			log.Error().Stack().Err(err)
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
