package command

import (
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/logger"
)

type Config struct {
	Command
	log *logger.Logger
}

func NewConfig() *Config {
	return &Config{
		log: logger.WithNamespace("ginger.config"),
	}
}

func (c *Config) Help() string {
	return `
config - Set or update project confiugration.

Usage:
  $ ginger config [options]

Options:
  --profile : Using profile name
  --region  : Set project region
  --role    : Set lambda execution role
`
}

func (c *Config) Run(ctx *args.Context) {
	conf := config.Load()
	if !conf.Exists() {
		c.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	defer conf.Write()

	var v string
	if v = ctx.String("profile"); v != "" {
		conf.Project.Profile = v
		c.log.Printf("Set use profile as \"%s\"\n", v)
	}
	if v = ctx.String("region"); v != "" {
		conf.Project.Region = v
		c.log.Printf("Set AWS region as \"%s\"\n", v)
	}
	if v = ctx.String("role"); v != "" {
		conf.Project.LambdaExecutionRole = v
		c.log.Printf("Set Lambda execution role as \"%s\"\n", v)
	}
	c.log.Info("Configuration updated!")
}