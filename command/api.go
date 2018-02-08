package command

import (
	"fmt"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
	"github.com/ysugimoto/go-args"
)

const (
	API_CREATE = "create"
	API_DELETE = "delete"
	API_INVOKE = "invoke"
)

type APIGateway struct {
	Command
	log *logger.Logger
}

func NewAPI() *APIGateway {
	return &APIGateway{
		log: logger.WithNamespace("ginger.api"),
	}
}

func (a *APIGateway) Help() string {
	return `
ginger api [subcommand] [options]

Subcommand:
  create: Create new endpoint
  delete: Delete endpoint
  invoke: invoke endpoint

Options:
  -p, path:   Path name
  -m, method: Method name (default=GET)
  -d, data:   Request payload (POST/PUT method only)
`
}

func (a *APIGateway) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		a.log.Error("Configuration file could not load. Run `ginger init` before.")
		return nil
	}
	switch ctx.At(1) {
	case API_CREATE:
		return a.createEndpoint(c, ctx)
	case API_DELETE:
		return a.deleteEndpoint(c, ctx)
	case API_INVOKE:
		return a.invokeEndpoint(c, ctx)
	default:
		fmt.Println(a.Help())
		return nil
	}
}

func (a *APIGateway) createEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		a.log.Error("Endpoint path is required. Run with -p, --path option.")
		return nil
	} else if c.API.Exists(path) {
		a.log.Errorf("Endpoint %s is already exists.\n", path)
	}
	api := entity.NewResource("", path)
	c.API.Resources = append(c.API.Resources, api)
	c.API.Sort()
	c.Write()
	a.log.Infof("API for path %s created successfully!\n", path)
	return nil
}

func (a *APIGateway) deleteEndpoint(c *config.Config, ctx *args.Context) error {
	if c.API.RestId == "" {
		a.log.Error("Any REST API isn't created yet.")
		return nil
	}
	path := ctx.String("path")
	if path == "" {
		a.log.Error("Endpoint path is required. Run with -p, --path option.")
		return nil
	} else if !c.API.Exists(path) {
		a.log.Errorf("Endpoint %s does not defined.\n", path)
		return nil
	}
	rs := c.API.Find(path)
	api := request.NewAPIGateway(c)
	if !api.ResourceExists(c.API.RestId, rs.Id) {
		a.log.Errorf("Rsource for %s not found on AWS.\n", path)
	} else if err := api.DeleteResource(c.API.RestId, rs.Id); err != nil {
		a.log.Error("Failed to delete from AWS. Please delete manually.")
	}
	c.API.Remove(path)
	c.API.Sort()
	c.Write()
	a.log.Info("Endpoint deleted successfully.")
	return nil
}

func (a *APIGateway) invokeEndpoint(c *config.Config, ctx *args.Context) error {
	return nil
}
