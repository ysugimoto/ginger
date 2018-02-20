package config

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/ysugimoto/ginger/entity"
)

func (c *Config) LoadStage(name string) (*entity.Stage, error) {
	path := filepath.Join(c.StagePath, fmt.Sprintf("%s.toml", name))
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrap(err, "Stage configuration file does not exist")
	}

	stg := &entity.Stage{}
	if _, err := toml.DecodeFile(path, stg); err != nil {
		return nil, errors.Wrap(err, "Failed to decode stage file")
	}
	return stg, nil
}

func (c *Config) DeleteStage(name string) error {
	path := filepath.Join(c.StagePath, name, fmt.Sprintf("%s.toml"))
	if _, err := os.Stat(path); err != nil {
		return errors.Wrap(err, "Stage configuration file does not exist")
	}

	if err := os.Remove(path); err != nil {
		return errors.Wrap(err, "Failed to delete stage file")
	}
	return nil
}

func (c *Config) LoadAllStages() ([]*entity.Stage, error) {
	stages := []*entity.Stage{}
	err := filepath.Walk(c.StagePath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		} else if info.IsDir() {
			return nil
		} else if filepath.Ext(info.Name()) != ".toml" {
			return nil
		}
		stg, err := c.LoadStage(filepath.Base(info.Name()))
		if err != nil {
			fmt.Printf("Skip: couldn't list stage \"%s\": %s\n", info.Name(), err.Error())
			return nil
		}
		stages = append(stages, stg)
		return nil
	})
	return stages, err
}
