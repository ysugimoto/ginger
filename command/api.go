package command

import (
	"fmt"
	"strings"

	"crypto/tls"
	"net/http"
	"net/http/httputil"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
	"github.com/ysugimoto/go-args"
)

const (
	API_CREATE = "create"
	API_DELETE = "delete"
	API_INVOKE = "invoke"
	API_LINK   = "link"
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
Usage:
  $ ginger api [operation] [options]

Operation:
  create : Create new endpoint
  delete : Delete endpoint
  invoke : invoke endpoint

Options:
  -p, --path   : [all] Path name
  -n, --name   : [link] Function name
  -s, --stage  : [invoke] Target stage
  -m, --method : [invoke] Method name (default=GET)
  -d, --data   : [invoke] Request payload (POST/PUT method only)
`
}

func (a *APIGateway) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		a.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	var err error
	defer func() {
		if err != nil {
			a.log.Error(err.Error())
			debugTrace(err)
		}
		c.API.Sort()
		c.Write()
	}()
	switch ctx.At(1) {
	case API_CREATE:
		err = a.createEndpoint(c, ctx)
	case API_DELETE:
		err = a.deleteEndpoint(c, ctx)
	case API_INVOKE:
		err = a.invokeEndpoint(c, ctx)
	case API_LINK:
		err = a.linkFunction(c, ctx)
	default:
		fmt.Println(a.Help())
	}
}

func (a *APIGateway) createEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if c.API.Exists(path) {
		return exception("Endpoint %s is already exists.\n", path)
	}

	c.API.Resources = append(c.API.Resources, entity.NewResource("", path))
	a.log.Infof("API for path %s created successfully!\n", path)
	return nil
}

func (a *APIGateway) deleteEndpoint(c *config.Config, ctx *args.Context) error {
	if c.API.RestId == "" {
		return exception("Any REST API isn't created yet.")
	}

	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if !c.API.Exists(path) {
		return exception("Endpoint %s does not defined.\n", path)
	}

	rs := c.API.Find(path)
	api := request.NewAPIGateway(c)
	if !api.ResourceExists(c.API.RestId, rs.Id) {
		return exception("Rsource for %s not found on AWS.\n", path)
	} else if err := api.DeleteResource(c.API.RestId, rs.Id); err != nil {
		a.log.Error("Failed to delete from AWS. Please delete manually.")
	}
	c.API.Remove(path)
	a.log.Info("Endpoint deleted successfully.")
	return nil
}

func (a *APIGateway) invokeEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if !c.API.Exists(path) {
		return exception("Endpoint %s does not defined.\n", path)
	}

	rs := c.API.Find(path)
	if rs.Id == "" {
		return exception("Endpoint %s hasn't not deployed yet.\n", path)
	}
	method := strings.ToUpper(ctx.String("method"))
	data := ctx.String("data")
	stage := ctx.String("stage")
	host := fmt.Sprintf("%s.execute-api.%s.amazonaws.com", c.API.RestId, c.Project.Region)
	callUrl := fmt.Sprintf("https://%s/%s%s", host, c.API.RestId, c.Project.Region, stage, path)

	a.log.Printf("Send HTTP request to %s\n", callUrl)

	req, err := http.NewRequest(strings.ToUpper(method), callUrl, strings.NewReader(data))
	if err != nil {
		return exception("Failed to create HTTP request: %s\n", err.Error())
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName: host,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return exception("Request Failed: %s\n", err.Error())
	}
	defer resp.Body.Close()
	if dump, err := httputil.DumpResponse(resp, true); err != nil {
		return exception("Failed to dump response: %s\n", err.Error())
	} else {
		a.log.Info("========== Response received =========")
		fmt.Println(string(dump))
	}
	return nil
}

func (a *APIGateway) linkFunction(c *config.Config, ctx *args.Context) error {
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

	rs := c.API.Find(path)
	if rs.Id != "" && rs.Integration != nil {
		inquiry := fmt.Sprintf("%s has already have integration to %s. Override it?", rs.Path, rs.Integration.LambdaFunction)
		if !input.Bool(inquiry) {
			return exception("Canceled.")
		}
		api := request.NewAPIGateway(c)
		api.DeleteMethod(c.API.RestId, rs.Id)
		api.DeleteIntegration(c.API.RestId, rs.Id)
	}
	rs.Integration.LambdaFunction = name
	a.log.Infof("Linked to function %s.\n", name)
	return nil
}
