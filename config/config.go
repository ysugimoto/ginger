package config

import (
	"fmt"
	"os"
	"sync"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ysugimoto/ginger/entity"
)

var mu sync.Mutex

func Load() *Config {
	cwd, _ := os.Getwd()
	c := &Config{
		Root:         cwd,
		Path:         filepath.Join(cwd, "Ginger.toml"),
		FunctionPath: filepath.Join(cwd, "functions"),
		APIPath:      filepath.Join(cwd, "apis"),
		VendorPath:   filepath.Join(cwd, "vendor"),
	}

	if _, err := os.Stat(c.Path); err == nil {
		c.exists = true
		if _, err = toml.DecodeFile(c.Path, c); err != nil {
			fmt.Println("Syntax error found on configuration file!")
			os.Exit(1)
		}
	}
	return c
}

type Config struct {
	Root         string `toml:"-"`
	Path         string `toml:"-"`
	FunctionPath string `toml:"-"`
	APIPath      string `toml:"-"`
	VendorPath   string `toml:"-"`

	exists bool `toml:"-"`

	Project   entity.Project   `toml:"project"`
	Functions entity.Functions `toml:"function"`
	API       entity.API       `toml:"api"`
}

func (c *Config) Exists() bool {
	return c.exists
}

func (c *Config) Write() {
	mu.Lock()
	defer mu.Unlock()
	fp, _ := os.OpenFile(c.Path, os.O_WRONLY|os.O_CREATE, 0644)
	defer fp.Close()
	enc := toml.NewEncoder(fp)
	enc.Encode(c)
}
