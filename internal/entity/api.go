package entity

import (
	"sort"
	"strings"
)

// formatPath formats strict path.
func formatPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

// API is struct that manages REST API id and registered resource endpoints.
type API struct {
	RestId    string      `toml:"rest_id"`
	Resources []*Resource `toml:"resource"`
}

// Exists() returns bool which resource is registered or not.
func (a API) Exists(path string) bool {
	path = formatPath(path)
	for _, r := range a.Resources {
		if r.Path == path {
			return true
		}
	}
	return false
}

// Find() returns resource pointer.
func (a API) Find(path string) *Resource {
	path = formatPath(path)
	for _, r := range a.Resources {
		if r.Path == path {
			return r
		}
	}
	return nil
}

// Remove() remvoes resource from slice.
func (a API) Remove(path string) {
	path = formatPath(path)
	for i, r := range a.Resources {
		if r.Path == path {
			a.Resources = append(a.Resources[0:i], a.Resources[i+1:]...)
		}
	}
}

// Sort() sorts resources order by shorter path.
// This function is implementation for sort.Sort() interface.
func (a API) Sort() {
	sort.Slice(a.Resources, func(i, j int) bool {
		return len(strings.Split(a.Resources[i].Path, "/")) < len(strings.Split(a.Resources[j].Path, "/"))
	})
}
