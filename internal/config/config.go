package config

import (
	"fmt"
	"os"
	"sync"

	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/ysugimoto/ginger/internal/entity"
)

var mu sync.Mutex

// Load loads configuration and map to Config struct.
// this function always returns although the config file didn't exist.
// Then you can confirm as Exists() on config file exists or not.
func Load() *Config {
	cwd, _ := os.Getwd()
	c := &Config{
		Root:         cwd,
		Path:         filepath.Join(cwd, "Ginger.toml"),
		FunctionPath: filepath.Join(cwd, "functions"),
		VendorPath:   filepath.Join(cwd, "vendor"),
		Project:      entity.Project{},
		Functions:    entity.Functions{},
		API: entity.API{
			Resources: make([]*entity.Resource, 0),
		},
	}

	if _, err := os.Stat(c.Path); err == nil {
		c.exists = true
		if _, err = toml.DecodeFile(c.Path, c); err != nil {
			fmt.Println("Syntax error found on configuration file!", err)
			os.Exit(1)
		}
	}
	// Resources need to sort by short paths.
	c.API.Sort()
	return c
}

// Config is the struct which maps configuration file into this.
// Ensure call Write() to update configuration.
type Config struct {
	Root         string `toml:"-"`
	Path         string `toml:"-"`
	FunctionPath string `toml:"-"`
	VendorPath   string `toml:"-"`

	exists bool `toml:"-"`

	Project   entity.Project   `toml:"project"`
	Functions entity.Functions `toml:"function"`
	API       entity.API       `toml:"api"`
}

// Exists() returns bool which config file exists or not.
func (c *Config) Exists() bool {
	return c.exists
}

// Write() writes configuration to file.
func (c *Config) Write() {
	mu.Lock()
	defer mu.Unlock()
	fp, _ := os.OpenFile(c.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer fp.Close()
	enc := toml.NewEncoder(fp)
	enc.Encode(c)
}
