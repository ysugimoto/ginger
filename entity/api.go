package entity

import (
	"sort"
	"strings"
)

func formatPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

type API struct {
	RestId    string      `toml:"rest_id"`
	Resources []*Resource `toml:"resource"`
}

func (a API) Exists(path string) bool {
	path = formatPath(path)
	for _, r := range a.Resources {
		if r.Path == path {
			return true
		}
	}
	return false
}

func (a API) Find(path string) *Resource {
	path = formatPath(path)
	for _, r := range a.Resources {
		if r.Path == path {
			return r
		}
	}
	return nil
}

func (a API) Remove(path string) {
	path = formatPath(path)
	for i, r := range a.Resources {
		if r.Path == path {
			a.Resources = append(a.Resources[0:i], a.Resources[i+1:]...)
		}
	}
}

func (a API) Sort() {
	sort.Slice(a.Resources, func(i, j int) bool {
		return len(strings.Split(a.Resources[i].Path, "/")) < len(strings.Split(a.Resources[j].Path, "/"))
	})
}
