package entity

import (
	"github.com/ysugimoto/ginger/internal/util"
)

// Resource is the entity struct which maps 'api.resources' slice in configuration.
type Resource struct {
	Id           string                  `toml:"id"`
	Path         string                  `toml:"path"`
	Integrations map[string]*Integration `toml:"integrations"`
	UserDefined  bool                    `toml:"user_defined"`
}

func NewResource(id, path string) *Resource {
	return &Resource{
		Id:           id,
		Path:         util.FormatPath(path),
		Integrations: nil,
		UserDefined:  false,
	}
}

func (r *Resource) GetIntegrations() map[string]*Integration {
	return r.Integrations
}

func (r *Resource) AddIntegration(method string, ig *Integration) {
	if r.Integrations == nil {
		r.Integrations = make(map[string]*Integration)
	}
	r.Integrations[method] = ig
}

func (r *Resource) GetIntegration(method string) *Integration {
	if r.Integrations == nil {
		return nil
	}
	if i, ok := r.Integrations[method]; !ok {
		return nil
	} else {
		return i
	}
}

func (r *Resource) DeleteIntegration(method string) {
	if r.Integrations == nil {
		return
	}
	if _, ok := r.Integrations[method]; ok {
		delete(r.Integrations, method)
	}
}
