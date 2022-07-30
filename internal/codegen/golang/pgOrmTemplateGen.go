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
func (r *apiRouter) getList(c *gin.Context) (data interface{}, httpCode int, err error){
	var filter types.DataTableCommonFilter
	if err := c.ShouldBind(&filter); err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}.getList - ShouldBind")
		return nil, http.StatusBadRequest, err
	}
	
	data, err = r.{{.Name}}Case.GetListWithFilter(c, filter)
	if err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}.getList - GetListWithFilter")
		return data, http.StatusInternalServerError, err
	}
	return
}

{{- if .IDExists}}
// get {{.Name}} by ID
func (r *apiRouter) getByID(c *gin.Context) (data interface{}, httpCode int, err error){
	var id types.ParamID{{.IDType}}
	if err = c.ShouldBindUri(&id); err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}.getByID - ShouldBindUri")
		return nil, http.StatusBadRequest, err
	}

	data, _, err = r.{{.Name}}Case.GetByID(c, id.ID)
	if err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}.getByID - GetByID")
		return data, http.StatusInternalServerError, err
	}
	return
}
{{- end}}

// submit new {{.Name}}
func (r *apiRouter) submit(c *gin.Context) (data interface{}, httpCode int, err error){
	var payload types.{{.Name}}Entity
	if err = c.BindJSON(&payload); err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}.submit - BindJSON")
		return nil, http.StatusBadRequest, err
	}

	err = r.{{.Name}}Case.Submit(c, payload)
	if err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}.submit - Submit")
		return nil, http.StatusInternalServerError, err
	}
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
	"{{.ProjectPath}}/internal/usecase"
	"{{.ProjectPath}}/internal/delivery/helper"
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
	r.Register(helper.NewMiddleware(cfg, group, uc.Auth))
}

func (r *apiRouter) Register(m helper.Middleware) {
	m.POST("/submit", r.submit)
	m.GET("/list", r.getList)	
	{{- if .IDExists}}
	m.GET("/get/:id", r.getByID)
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
	{{.}}API.InitRoute(d.cfg, group{{.}}, d.caseWrapper)
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
		log.Error().Stack().Err(err).Msg("{{.Name}}Case.Submit - Submit")
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
			log.Error().Stack().Err(err).Msg("{{.Name}}Case.SubmitMultiple - SubmitMultiple")
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
	if filter.Limit == 0 {
		filter.Limit = 10
	}
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

	data, found, err = c.repo{{.Name}}.GetByID(ctx, id)
	if err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}Case.GetByID - repo{{.Name}}.GetByID")
	}
	return
}

func (c *{{.Name}}Case) UpdateByID(ctx context.Context, data types.{{.Name}}Entity) (err error){
	ctx, span := lib.StartUseCaseSpan(ctx, "{{.Name}}Case.UpdateByID")
	defer span.End()
	defer span.RecordError(err)

	err = c.repo{{.Name}}.UpdateByID(ctx, data)
	if err != nil {
		log.Error().Stack().Err(err).Msg("{{.Name}}Case.UpdateByID - repo{{.Name}}.UpdateByID")
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
		log.Error().Stack().Err(err).Msg("util.InitRepo - InitDB")
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
	{{- if .NoDeletedAt}}
	DeletedAt time.Time {{$.Q}}form:"-" gorm:"autoCreateTime:false" json:"-"{{$.Q}}
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
			log.Error().Stack().Err(err).Msg("{{.Name}}Data.Submit")
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
			log.Error().Stack().Err(err).Msg("{{.Name}}Data.SubmitMultiple")
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

	rs := d.SlaveDB.WithContext(ctx).Take(&data, "id = ?", id)
	if rs.Error == gorm.ErrRecordNotFound {
		err = nil
	}else{
		err = rs.Error
	}
	found = rs.RowsAffected > 0
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
			log.Error().Stack().Err(err).Msg("{{.Name}}Data.UpdateByID")
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
