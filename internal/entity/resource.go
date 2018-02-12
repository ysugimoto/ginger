package entity

// Resource is the entity struct which maps 'api.resources' slice in configuration.
type Resource struct {
	Id           string                  `toml:"id"`
	Path         string                  `toml:"path"`
	Integrations map[string]*Integration `toml:"integrations"`
	UserDefined  bool                    `toml:"user_defined"`
}

func NewResource(id, path string) *Resource {
	return &Resource{
		Id:          id,
		Path:        formatPath(path),
		Integration: nil,
		UserDefined: false,
	}
}

func (r *Resource) GetIntegration(method string) *Integration {
	if r.Intergrations == nil {
		return nil
	}
	if i, ok := r.Integrations[method]; !ok {
		return nil
	} else {
		return i
	}
}
