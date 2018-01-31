package command

import (
	"fmt"
	"os"

	"github.com/ysugimoto/go-args"
)

type Build struct {
	Command
	conf *Config
}

func NewBuild() *Build {
	return &Build{
		conf: NewConfig(),
	}
}

func (b *Build) Run(ctx *args.Context) (err error) {
	if !b.conf.Exists {
		fmt.Println("[Create] Configuration file isn't exist. Run `ginger init` befgore.")
		os.Exit(1)
	}
	envs := commandEnvironment()
	for _, f := range b.conf.Project.Functions {
		if err := f.Build(b.conf.Root, envs); err != nil {
			fmt.Println("[Build] Failed: ", err)
		}
	}
	return nil
}
