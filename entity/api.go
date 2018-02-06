package entity

import (
	"strings"
)

type Resource struct {
	Id                 string `toml:"id"`
	Path               string `toml:"path"`
	IntegratedFunction string `toml:"function"`
}

type API struct {
	RestId    string      `toml:"rest_id"`
	Resources []*Resource `toml:"resource"`
}

func (a API) FormatPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

func (a API) Exists(path string) bool {
	for _, r := range a.Resources {
		if rs.Path == a.FormatPath(path) {
			return true
		}
	}
}
