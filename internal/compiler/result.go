package compiler

import (
	"github.com/ujunglangit-id/sqlc/internal/sql/catalog"
)

type Result struct {
	Catalog *catalog.Catalog
	Queries []*Query
}
