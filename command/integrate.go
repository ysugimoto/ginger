package command

import (
	"fmt"
	"strings"

	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
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
	} else if _, err := c.LoadFunction(name); err != nil {
		return exception("Function %s could't find in your project.", name)
	}

	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	}
	rs, err := c.LoadResource(path)
	if err != nil {
		return exception("Endpoint %s couldn't find in your project.\n", path)
	}

	method := "ANY"
	if m := ctx.String("method"); m != "" {
		method = strings.ToUpper(m)
		if method != "ANY" {
			i.log.Warn("Method didn't be supplied, so use \"ANY\" method to handle all methods.")
		}
	}

	ig := rs.GetIntegration(method)
	if ig != nil {
		// If interagraion already exists, need to delete it.
		inquiry := fmt.Sprintf("%s has already have integration for %s %s. Override it?", rs.Path, method, ig.String())
		if !input.Bool(inquiry) {
			return exception("Canceled.")
		}
		if rs.IntegrationId != nil {
			api := request.NewAPIGateway(c)
			api.DeleteMethod(c.RestApiId, *rs.IntegrationId, method)
			api.DeleteIntegration(c.RestApiId, *rs.IntegrationId, method)
		}
		rs.DeleteIntegration(method)
	}

	ig = entity.NewIntegration("lambda", name, rs.Path)
	rs.AddIntegration(method, ig)
	i.log.Infof("Integrated: Function %s to resource %s.\n", name, path)
	return nil
}

// integrateStorage makes integration between API Gateway resource and S3 storage
func (i *Integrate) integrateStorage(c *config.Config, ctx *args.Context) error {
	bucket := ctx.String("bucket")
	if bucket == "" {
		bucket = c.S3BucketName
	}
	if bucket != c.S3BucketName {
		i.log.Warnf("Target bucket %s is external bucket. ginger only manages integration.\n", bucket)
	}
	path := ctx.String("path")
	if path == "" {
		return exception("Resource path is required. Run with -p, --path option.")
	}
	rs, err := c.LoadResource(path)
	if err != nil {
		return exception("Resource %s could not find in your project.\n", path)
	}

	// On storage integration, we only support "GET" method.
	method := "GET"
	ig := rs.GetIntegration(method)
	if ig != nil {
		// If interagraion already exists. need to delete it.
		inquiry := fmt.Sprintf("%s has already have integration for %s %s. Override it?", rs.Path, method, ig.String())
		if !input.Bool(inquiry) {
			return exception("Canceled.")
		}
		if rs.IntegrationId != nil {
			api := request.NewAPIGateway(c)
			api.DeleteMethod(c.RestApiId, *rs.IntegrationId, method)
			api.DeleteIntegration(c.RestApiId, *rs.IntegrationId, method)
		}
		rs.DeleteIntegration(method)
	}

	ig = entity.NewIntegration("s3", bucket, rs.Path)
	rs.AddIntegration(method, ig)
	i.log.Infof("Integrated: Storage %s to resource %s.\n", bucket, path)
	return nil
}
