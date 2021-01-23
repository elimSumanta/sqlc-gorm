package golang

import (
	"strings"

	"github.com/ujunglangit-id/sqlc/internal/config"
	"github.com/ujunglangit-id/sqlc/internal/core"
)

type Struct struct {
	Table       core.FQN
	ProjectPath string
	Name        string
	Fields      []Field
	Comment     string
}

func StructName(name string, settings config.CombinedSettings) string {
	if rename := settings.Rename[name]; rename != "" {
		return rename
	}
	out := ""
	for _, p := range strings.Split(name, "_") {
		if p == "id" {
			out += "ID"
		} else {
			out += strings.Title(p)
		}
	}
	return out
}
