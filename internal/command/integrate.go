package command

import (
	"fmt"
	"strings"

	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/entity"
	"github.com/ysugimoto/ginger/internal/input"
	"github.com/ysugimoto/ginger/internal/logger"
	"github.com/ysugimoto/ginger/internal/request"
)

const (
	INTEGRATE_FUNCTION = "function"
	INTEGRATE_STORAGE  = "storage"
)

// Integrate is the struct for making integration on AWS API Gateway.
// This command operates with above constant subcommand string.
type Integrate struct {
	Command
	log *logger.Logger
}

func NewIntegrate() *Integrate {
	return &Integrate{
		log: logger.WithNamespace("ginger.integrate"),
	}
}

// Show function command help.
func (i *Integrate) Help() string {
	return COMMAND_HEADER + `
integrate - API Gateway integrate management command.

Usage:
  $ ginger integrate [operation] [options]

Operation:
  function : Make integration with function (AWS Lambda)
  storage  : Make integration with storage (AWS S3)
  help     : Show this help

Options:
  -p, --path    : [all] integration resource path (required)
  -n, --name    : [function] function name
  -m, --method  : [all] integration HTTP method
`
}

// Run the command.
func (i *Integrate) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		i.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	var err error
	defer func() {
		if err != nil {
			i.log.Error(err.Error())
			debugTrace(err)
		}
		c.Write()
	}()

	switch ctx.At(1) {
	case INTEGRATE_FUNCTION:
		err = i.integrateFunction(c, ctx)
	case INTEGRATE_STORAGE:
		err = i.integrateStorage(c, ctx)
	default:
		fmt.Println(i.Help())
	}
}

// integrateFunction makes integration between API Gateway resource and Lambda function
func (i *Integrate) integrateFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if !c.Functions.Exists(name) {
		return exception("Function %s doesn't defined in your project.")
	}

	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if !c.API.Exists(path) {
		return exception("Endpoint %s does not defined in your project.\n", path)
	}

	method := "ANY"
	if m := ctx.String("method"); m != "" {
		method = strings.ToUpper(m)
		if method != "ANY" {
			i.log.Warn("method didn't be defined. Use \"ANY\" method to handle all methods.")
		}
	}

	rs := c.API.Find(path)
	ig := rs.GetIntegration(method)
	if ig != nil {
		// If interagraion already exists. need to delete it.
		inquiry := fmt.Sprintf("%s has already have integration for %s %s. Override it?", rs.Path, method, ig.String())
		if !input.Bool(inquiry) {
			return exception("Canceled.")
		}
		if rs.IntegrationId != "" {
			api := request.NewAPIGateway(c)
			api.DeleteMethod(c.API.RestId, rs.IntegrationId, method)
			api.DeleteIntegration(c.API.RestId, rs.IntegrationId, method)
		}
		rs.DeleteIntegration(method)
	}

	ig = entity.NewIntegration("lambda", name, rs.Path)
	rs.AddIntegration(method, ig)
	i.log.Infof("Linked function %s to resource %s.\n", name, path)
	return nil
}

// integrateStorage makes integration between API Gateway resource and S3 storage
func (i *Integrate) integrateStorage(c *config.Config, ctx *args.Context) error {
	bucket := ctx.String("bucket")
	if bucket == "" {
		bucket = c.Project.S3BucketName
	}
	if bucket != c.Project.S3BucketName {
		i.log.Warnf("Target bucket %s is external bucket. ginger only manages integration.\n", bucket)
	}
	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if !c.API.Exists(path) {
		return exception("Endpoint %s does not defined in your project.\n", path)
	}

	// On storage integration, we only support "GET" method.
	method := "GET"
	rs := c.API.Find(path)
	ig := rs.GetIntegration(method)
	if ig != nil {
		// If interagraion already exists. need to delete it.
		inquiry := fmt.Sprintf("%s has already have integration for %s %s. Override it?", rs.Path, method, ig.String())
		if !input.Bool(inquiry) {
			return exception("Canceled.")
		}
		if rs.IntegrationId != "" {
			api := request.NewAPIGateway(c)
			api.DeleteMethod(c.API.RestId, rs.IntegrationId, method)
			api.DeleteIntegration(c.API.RestId, rs.IntegrationId, method)
		}
		rs.DeleteIntegration(method)
	}

	ig = entity.NewIntegration("s3", bucket, rs.Path)
	rs.AddIntegration(method, ig)
	i.log.Infof("Linked storage %s to resource %s.\n", bucket, path)
	return nil
}
