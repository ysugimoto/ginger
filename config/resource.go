package config

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/internal/util"
)

var ResourceNotFound = errors.New("Resource not found")

func (c *Config) LoadResource(path string) (*entity.Resource, error) {
	if path == "" {
		return nil, errors.New("Coulnd't function for empty path.")
	}
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

func (c *Config) ChooseResource() string {
	choose := []string{}
	for _, r := range c.Resources {
		if !r.UserDefined {
			continue
		}
		choose = append(choose, r.Path)
	}
	if len(choose) == 0 {
		return ""
	}
	return input.Choice("Select target path", choose)
}
