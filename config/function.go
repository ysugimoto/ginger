package config

import (
	"os"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
)

func (c *Config) LoadFunction(name string) (*entity.Function, error) {
	path := filepath.Join(c.FunctionPath, name, "Function.toml")
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrap(err, "Configuration file does not exist")
	}

	fn := &entity.Function{}
	if _, err := toml.DecodeFile(path, fn); err != nil {
		return nil, errors.Wrap(err, "Failed to decode configuration file")
	}
	c.Queue[name] = fn
	return fn, nil
}

func (c *Config) DeleteFunction(name string) error {
	path := filepath.Join(c.FunctionPath, name)
	if _, err := os.Stat(path); err != nil {
		return errors.Wrap(err, "Function directory does not exist")
	}

	if err := os.RemoveAll(path); err != nil {
		return errors.Wrap(err, "Failed to delete directory")
	}
	if _, ok := c.Queue[name]; ok {
		delete(c.Queue, name)
	}
	return nil
}

func (c *Config) LoadAllFunctions() ([]*entity.Function, error) {
	functions := []*entity.Function{}
	err := filepath.Walk(c.FunctionPath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		} else if !info.IsDir() {
			return nil
		} else if path == c.FunctionPath {
			return nil
		}
		fn, err := c.LoadFunction(info.Name())
		if err != nil {
			c.log.Warnf("Skip: couldn't list function \"%s\": %s\n", info.Name(), err.Error())
			return nil
		}
		functions = append(functions, fn)
		return nil
	})
	return functions, err
}

func (c *Config) ChooseFunction() string {
	choose := []string{}
	filepath.Walk(c.FunctionPath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		} else if !info.IsDir() {
			return nil
		} else if path == c.FunctionPath {
			return nil
		}
		fn, err := c.LoadFunction(info.Name())
		if err != nil {
			c.log.Warnf("Skip: couldn't list function \"%s\": %s\n", info.Name(), err.Error())
			return nil
		}
		choose = append(choose, fn.Name)
		return nil
	})
	if len(choose) == 0 {
		return ""
	}
	return input.Choice("Select target fucntion", choose)
}
