package config

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
)

func (c *Config) LoadScheduler(name string) (*entity.Scheduler, error) {
	path := filepath.Join(c.SchedulerPath, fmt.Sprintf("%s.toml", name))
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrap(err, "Configuration file does not exist")
	}

	sc := &entity.Scheduler{}
	if _, err := toml.DecodeFile(path, sc); err != nil {
		return nil, errors.Wrap(err, "Failed to decode configuration file")
	}
	return sc, nil
}

func (c *Config) DeleteScheduler(name string) error {
	path := filepath.Join(c.SchedulerPath, fmt.Sprintf("%s.toml", name))
	if _, err := os.Stat(path); err != nil {
		return errors.Wrap(err, "Function directory does not exist")
	}

	if err := os.Remove(path); err != nil {
		return errors.Wrapf(err, "Failed to delete file: %s.toml", name)
	}
	return nil
}

func (c *Config) LoadAllSchedulers() ([]*entity.Scheduler, error) {
	scs := []*entity.Scheduler{}
	err := filepath.Walk(c.SchedulerPath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		} else if info.IsDir() {
			return nil
		} else if path == c.SchedulerPath {
			return nil
		} else if filepath.Ext(path) != ".toml" {
			return nil
		}
		name := info.Name()
		sc, err := c.LoadScheduler(name[0 : len(name)-5])
		if err != nil {
			c.log.Warnf("Skip: couldn't list scheduler \"%s\": %s\n", info.Name(), err.Error())
			return nil
		}
		scs = append(scs, sc)
		return nil
	})
	return scs, err
}

func (c *Config) ChooseScheduler() string {
	choose := []string{}
	filepath.Walk(c.SchedulerPath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		} else if info.IsDir() {
			return nil
		} else if path == c.SchedulerPath {
			return nil
		} else if filepath.Ext(path) != ".toml" {
			return nil
		}
		name := info.Name()
		choose = append(choose, name[0:len(name)-5])
		return nil
	})
	if len(choose) == 0 {
		return ""
	}
	return input.Choice("Select target scheduler", choose)
}
