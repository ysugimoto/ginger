package config

import (
	"os"

	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
)

// Load loads configuration and map to Config struct.
// this function always returns although the config file didn't exist.
// Then you can confirm as Exists() on config file exists or not.
func Load() *Config {
	root := findUp()

	c := &Config{
		Root:         root,
		Path:         filepath.Join(root, "Ginger.toml"),
		FunctionPath: filepath.Join(root, "functions"),
		StoragePath:  filepath.Join(root, "storage"),
		LibPath:      filepath.Join(root, ".ginger"),
		StagePath:    filepath.Join(root, "stages"),
		Resources:    make([]*entity.Resource, 0),
		Queue:        make(map[string]*entity.Function, 0),
		log:          logger.WithNamespace("ginger.config"),
	}

	if _, err := os.Stat(c.Path); err == nil {
		c.exists = true
		if _, err = toml.DecodeFile(c.Path, c); err != nil {
			c.log.Errorf("Syntax error found on configuration file!\n", err)
			os.Exit(1)
		}
	}

	c.SortResources()
	return c
}

// findUp finds ginger project root from current working directory.
func findUp() string {
	path, _ := os.Getwd()

	for {
		if path == "/" {
			break
		}
		if _, err := os.Stat(filepath.Join(path, "Ginger.toml")); err == nil {
			return path
		}
		path = filepath.Dir(path)
	}
	path, _ = os.Getwd()
	return path
}
