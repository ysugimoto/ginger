package command

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/k0kubun/pp"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/go-args"
)

type Project struct {
	Functions []entity.Function `toml:"function"`
}
type Config struct {
	Command
	Root    string  `toml:"-"`
	Exists  bool    `toml:"-"`
	Project Project `toml:"ginger"`
}

func NewConfig() *Config {
	c := &Config{}
	cwd, _ := os.Getwd()
	confFile := filepath.Join(cwd, ".ginger")
	if _, err := os.Stat(confFile); err == nil {
		if _, err := toml.DecodeFile(filepath.Join(cwd, ".ginger"), c); err != nil {
			fmt.Println("Syntax error found on configuration file!")
			os.Exit(1)
		}
		c.Exists = true
	}
	c.Root = cwd
	return c
}

func (c *Config) Path() string {
	return filepath.Join(c.Root, ".ginger")
}

func (c *Config) addFunc(fn entity.Function) {
	c.Project.Functions = append(c.Project.Functions, fn)
}

func (c *Config) write() {
	fp, _ := os.OpenFile(filepath.Join(c.Root, ".ginger"), os.O_WRONLY|os.O_CREATE, 0644)
	enc := toml.NewEncoder(fp)
	enc.Encode(c)
}

func (c *Config) Run(ctx *args.Context) (err error) {
	pp.Print(c)
	return nil
}
