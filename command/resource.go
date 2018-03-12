package command

import (
	"fmt"
	"strings"

	"crypto/tls"
	"net/http"
	"net/http/httputil"

	"github.com/mattn/go-tty"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
)

const (
	RESOURCECREATE = "create"
	RESOURCEDELETE = "delete"
	RESOURCEINVOKE = "invoke"
	RESOURCEDEPLOY = "deploy"
	RESOURCELIST   = "list"
	RESOURCEHELP   = "help"
)

// Resource is the struct of AWS API Gateway resource management command.
// This struct will be dispatched on "ginger resource" subcommand.
// This command operates with above constant string.
type Resource struct {
	Command
	log *logger.Logger
}

func NewResource() *Resource {
	return &Resource{
		log: logger.WithNamespace("ginger.api"),
	}
}

// Display help string
func (r *Resource) Help() string {
	return commandHeader() + `
resource - AWS Resource resource management command.

Usage:
  $ ginger resource [operation] [options]

Operation:
  create  : Create new resource
  delete  : Delete resource
  invoke  : Invoke resources
  deploy  : Deploy resource
  list    : List resources
  help    : Show this help

Options:
  -p, --path   : [all] Path name
  -s, --stage  : [invoke] Target stage
  -m, --method : [invoke] Method name (default=GET)
  -b, --body   : [invoke] Request payload (POST/PUT method only)
`
}

// Run runs command with some options
func (r *Resource) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		r.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}

	var err error
	defer func() {
		if err != nil {
			r.log.Error(err.Error())
			debugTrace(err)
		}
		c.SortResources()
		c.Write()
	}()

	switch ctx.At(1) {
	case RESOURCECREATE:
		err = r.createEndpoint(c, ctx)
	case RESOURCEDELETE:
		err = r.deleteEndpoint(c, ctx)
	case RESOURCEINVOKE:
		err = r.invokeEndpoint(c, ctx)
	case RESOURCEDEPLOY:
		err = NewDeploy().deployResource(c, ctx)
	case RESOURCELIST:
		err = r.listEndpoint(c, ctx)
	default:
		fmt.Println(r.Help())
	}
}

// createEndpoint creates new resource by supplied path
//
// >>> doc
//
// ## Create new resource
//
// Create new API Gateway resource.
//
// ```
// $ ginger resource create [options]
// ```
//
// | option  | description                                                                                              |
// |:-------:|:---------------------------------------------------------------------------------------------------------|
// | --path  | Resource path. If this option isn't supplied, ginger will ask it                                         |
//
// API Gateway manages resources as 'path part' by each segments,
// But ginger can manage resource as path, so you don't need to care about it.
//
// <<< doc
func (r *Resource) createEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		path = input.String("Type resource path which you want to create")
	}
	rs, err := c.LoadResource(path)
	if err == nil {
		if rs.UserDefined {
			return exception("Resource path \"%s\" is already exists.\n", path)
		}
		rs.UserDefined = true
		r.log.Infof("Resource \"%s\" created successfully.\n", path)
		return nil
	}

	nr := entity.NewResource("", path)
	nr.UserDefined = true
	c.Resources = append(c.Resources, nr)
	r.log.Infof("Resource \"%s\" created successfully.\n", path)
	return nil
}

// deleteEndpoint deletes resource by supplied path.
//
// >>> doc
//
// ## Delete resource
//
// Delete resource endpoint.
//
// ```
// $ ginger resource delete [options]
// ```
//
// | option  | description                                       |
// |:-------:|:--------------------------------------------------|
// | --path  | [Required] Resource path which you want to delete |
//
// Note that if sub-path exists on target path, those paths also will be deleted.
//
// <<< doc
func (r *Resource) deleteEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	}
	rs, err := c.LoadResource(path)
	if err != nil || !rs.UserDefined {
		return exception("Endpoint %s could not find.\n", path)
	}

	if !ctx.Has("force") && !input.Bool("Subpath also removed. Are you sure?") {
		r.log.Warn("Aborted.")
		return nil
	}

	if c.RestApiId != "" {
		api := request.NewAPIGateway(c)
		if !api.ResourceExists(c.RestApiId, rs.Id) {
			return exception("Resource not found for %s on AWS.\n", path)
		} else if err := api.DeleteResource(c.RestApiId, rs.Id); err != nil {
			r.log.Error("Failed to delete from AWS. Please delete manually.")
		}
	}
	c.DeleteResource(path)
	// Remove recursive
	for _, sr := range c.FindSubResources(path) {
		c.DeleteResource(sr.Path)
	}
	r.log.Info("Resource deleted successfully.")
	return nil
}

// invokeEndpoint invokes resource with HTTP request.
//
// >>> doc
//
// ## Invoke resource
//
// Send HTTP request to destination endpoint.
//
// ```
// $ ginger resource invoke [options]
// ```
//
// | option   | description                                                       |
// |:--------:|:------------------------------------------------------------------|
// | --stage  | [Required] Target stage which you deployed to                     |
// | --path   | Resource path. If this option isn't supplied, ginger will ask it  |
// | --method | HTTP request method. Default is GET                               |
// | --body   | Request body string. This option enables only POST or PUT request |
//
// <<< doc
// Ensure some lambda fucntion integartion has set to handle request. Otherwise request will be failed.
func (r *Resource) invokeEndpoint(c *config.Config, ctx *args.Context) error {
	stage := ctx.String("stage")
	if stage == "" {
		return exception("Stage couldn't find. please run with --stage [stage name] option.")
	}
	path := ctx.String("path")
	if path == "" {
		path = input.String("Type invoke path")
	}
	if path == "" {
		path = "/"
	}

	// Build request
	host := fmt.Sprintf("%s.execute-api.%s.amazonaws.com", c.RestApiId, c.Region)
	callUrl := fmt.Sprintf("https://%s/%s%s", host, stage, path)

	method := "GET"
	if m := ctx.String("method"); m != "" {
		method = strings.ToUpper(m)
	}

	r.log.Printf("Send HTTP request to %s %s\n", method, callUrl)
	req, err := http.NewRequest(method, callUrl, strings.NewReader(ctx.String("body")))
	if err != nil {
		return exception("Failed to create HTTP request: %s\n", err.Error())
	}
	// Send request with TLS transport because API Gateway always supports TLS connection.
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
		r.log.Info("========== Response received =========")
		fmt.Println(string(dump))
	}
	return nil
}

// listEndpoint shows endpoint list
//
// >>> doc
//
// ## List resources
//
// Display list of registered API Gateway resources.
//
// ```
// $ ginger resource list
// ```
//
// <<< doc
func (r *Resource) listEndpoint(c *config.Config, ctx *args.Context) error {
	t, err := tty.Open()
	if err != nil {
		return exception("Couldn't open tty")
	}
	defer t.Close()
	w, _, err := t.Size()
	if err != nil {
		return exception("Couldn't get tty size")
	}
	line := strings.Repeat("=", w)
	fmt.Println(line)
	fmt.Printf("%-36s %-16s %-36s %4s\n", "Path", "ResourceId", "Integrations", "Deployed")
	fmt.Println(line)
	for i, r := range c.Resources {
		if !r.UserDefined {
			continue
		}
		d := "no"
		if r.Id != "" {
			d = "yes"
		}
		f := "-"
		if igs := r.GetIntegrations(); igs != nil {
			for m, i := range igs {
				f += fmt.Sprintf("%s:%s", m, i.String())
			}
		}
		fmt.Printf("%-36s %-16s %-36s %-4s\n", r.Path, r.Id, f, d)
		if i != len(c.Resources)-1 {
			fmt.Println(strings.Repeat("-", w))
		}
	}
	return nil
}
