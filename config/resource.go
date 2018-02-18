package config

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/internal/util"
)

var ResourceNotFound = errors.New("Resource not found")

func (c *Config) LoadResource(path string) (*entity.Resource, error) {
	path = util.FormatPath(path)
	for _, r := range c.Resources {
		if r.Path == path {
			return r, nil
		}
	}
	return nil, ResourceNotFound
}

func (c *Config) FindSubResources(prefix string) []*entity.Resource {
	rs := []*entity.Resource{}
	prefix = util.FormatPath(prefix)
	for _, r := range c.Resources {
		if strings.HasPrefix(r.Path, prefix) {
			rs = append(rs, r)
		}
	}
	return rs
}

func (c *Config) DeleteResource(path string) error {
	path = util.FormatPath(path)
	for i, r := range c.Resources {
		if r.Path == path {
			c.Resources = append(c.Resources[0:i], c.Resources[i+1:]...)
			return nil
		}
	}
	return ResourceNotFound
}
