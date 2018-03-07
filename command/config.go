package command

import (
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/logger"
)

// Config is a struct for update project configuration.
type Config struct {
	Command
	log *logger.Logger
}

func NewConfig() *Config {
	return &Config{
		log: logger.WithNamespace("ginger.config"),
	}
}

// Show this command help.
func (c *Config) Help() string {
	return COMMAND_HEADER + `
config - Set or update project confiugration.

Usage:
  $ ginger config [options]

Options:
  --profile : Using profile name
  --region  : Set project region
  --role    : Set default lambda execution role
  --bucket  : Set S3 bucket name which you want to use
  --hook    : Register deploy hook command
`
}

// Run the config command.
//
// >>> doc
//
// ## Update configuration
//
// Update configurations by supplied command options.
//
// ```
// $ ginger config [options]
// ```
//
// | option    | description                                                                  |
// |:---------:|:----------------------------------------------------------------------------:|
// | --profile | Accout profile name. If empty, ginger uses `default` or environment variable |
// | --region  | Region name to deploy                                                        |
// | --bucket  | S3 bucket name                                                               |
// | --hook    | Deploy hook command                                                          |
//
// <<< doc
func (c *Config) Run(ctx *args.Context) {
	conf := config.Load()
	if !conf.Exists() {
		c.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	defer conf.Write()

	var v string
	if v = ctx.String("profile"); v != "" {
		conf.Profile = v
		c.log.Printf("Set use profile as \"%s\"\n", v)
	}
	if v = ctx.String("region"); v != "" {
		conf.Region = v
		c.log.Printf("Set AWS region as \"%s\"\n", v)
	}
	if v = ctx.String("role"); v != "" {
		conf.DefaultLambdaRole = v
		c.log.Printf("Set Lambda execution role as \"%s\"\n", v)
	}
	if v = ctx.String("bucket"); v != "" {
		conf.S3BucketName = v
		c.log.Printf("Set S3 bucket name as \"%s\"\n", v)
	}
	if v = ctx.String("hook"); v != "" {
		conf.DeployHookCommand = v
		c.log.Print("Set deploy hook command.")
	}
	c.log.Info("Configuration updated!")
}
