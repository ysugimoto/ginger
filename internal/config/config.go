package config

import (
	"fmt"
	"os"
	"sync"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/ysugimoto/ginger/internal/entity"
)

// Load loads configuration and map to Config struct.
// this function always returns although the config file didn't exist.
// Then you can confirm as Exists() on config file exists or not.
func Load() *Config {
	root, err := findUp()
	if err != nil {
		fmt.Println("Unexpected Error", err)
		os.Exit(1)
	}

	c := &Config{
		Root:         root,
		Path:         filepath.Join(root, "Ginger.toml"),
		FunctionPath: filepath.Join(root, "functions"),
		VendorPath:   filepath.Join(root, "vendor"),
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

// findUp finds ginger project root from current working directory.
func findUp() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if path == "/" {
			break
		}
		if _, err := os.Stat(filepath.Join(path, "Ginger.toml")); err == nil {
			return path, nil
		}
		path = filepath.Dir(path)
	}
	return "", errors.New("Failed to find up configuration file")
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

// Mutex for file I/O
var mu sync.Mutex

// Write() writes configuration to file.
func (c *Config) Write() {
	mu.Lock()
	defer mu.Unlock()
	fp, _ := os.OpenFile(c.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer fp.Close()
	enc := toml.NewEncoder(fp)
	enc.Encode(c)
}
