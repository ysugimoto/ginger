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

func Load() (c *Config, err error) {
	c.Root, _ = os.Getwd()
	c.Path = filepath.Join(c.Root, "Ginger.toml")
	c.FunctionPath = filepath.Join(c.Root, "functions")
	c.APIPath = filepath.Join(c.Root, "apis")
	c.VendorPath = filepath.Join(c.Root, "vendor")

	if _, err = os.Stat(c.Path); err == nil {
		c.exists = true
		if _, err = toml.DecodeFile(c.Path, c); err != nil {
			fmt.Println("Syntax error found on configuration file!")
			os.Exit(1)
		}
	}
	return c, err
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
	Apis      entity.APIs      `toml:"api"`
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
