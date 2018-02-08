package command

import (
	"fmt"
	"strings"

	"crypto/tls"
	"io/ioutil"
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

func (a *APIGateway) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		a.log.Error("Configuration file could not load. Run `ginger init` before.")
		return nil
	}
	defer func() {
		c.API.Sort()
		c.Write()
	}()
	switch ctx.At(1) {
	case API_CREATE:
		return a.createEndpoint(c, ctx)
	case API_DELETE:
		return a.deleteEndpoint(c, ctx)
	case API_INVOKE:
		return a.invokeEndpoint(c, ctx)
	case API_LINK:
		return a.linkFunction(c, ctx)
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
	a.log.Info("Endpoint deleted successfully.")
	return nil
}

func (a *APIGateway) invokeEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		a.log.Error("Endpoint path is required. Run with -p, --path option.")
		return nil
	} else if !c.API.Exists(path) {
		a.log.Errorf("Endpoint %s does not defined.\n", path)
		return nil
	}
	rs := c.API.Find(path)
	if rs.Id == "" {
		a.log.Errorf("Endpoint %s hasn't not deployed yet.\n", path)
		return nil
	}
	method := strings.ToUpper(ctx.Strig("method"))
	data := ctx.String("data")
	stage := ctx.String("stage")
	host := fmt.Sprintf("%s.execute-api.%s.amazonaws.com", c.API.RestId, c.Project.Region)
	callUrl := fmt.Sprintf("https://%s/%s%s", host, c.API.RestId, c.Project.Region, stage, path)

	a.log.Printf("Send HTTP request to %s\n", callUrl)

	req := http.NewRequest(strings.ToUpper(method), callUrl, strings.NewReader(data))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName: host,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		a.log.Errorf("Request Failed: %s\n", err.Error())
		return nil
	}
	defer resp.Body.Close()
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		a.log.Errorf("Failed to dump response: %s\n", err.Error())
		return nil
	}
	a.log.Info("Response received")
	a.log.Print(string(dump))
	return nil
}

func (a *APIGateway) linkFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		a.log.Error("Function name didn't supplied. Run with --name option.")
		return nil
	} else if !c.Functions.Exists(name) {
		a.log.Error("Function not defined.")
		return nil
	}
	fn := c.Functions.Find(name)
	path := ctx.String("path")
	if path == "" {
		a.log.Error("Endpoint path is required. Run with -p, --path option.")
		return nil
	} else if !c.API.Exists(path) {
		a.log.Errorf("Endpoint %s does not defined.\n", path)
		return nil
	}
	rs := c.API.Find(path)
	if rs.Id != "" && rs.Integration != nil {
		if !input.Bool("%s has already have integration to %s. Override it?", rs.Path, rs.Integration.LambdaFunction) {
			a.log.Print("Canceled.")
			return nil
		}
		api := request.NewAPIGateway(c)
		api.DeleteMethod(c.API.RestId, rs.Id)
		api.DeleteIntegration(c.API.RestId, rs.Id)
	}
	rs.Integration.LambdaFunction = name
	a.log.Infof("Linked to function %s.\n", name)
	return nil
}
