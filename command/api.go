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
	API_CREATE = "create"
	API_DELETE = "delete"
	API_INVOKE = "invoke"
	API_DEPLOY = "deploy"
	API_LIST   = "list"
	API_HELP   = "help"
)

// AWS APIGateway REST API operation command struct
// This struct will be dispatched on "ginger api" subcommand.
// This command operates with above constant string.
type APIGateway struct {
	Command
	log *logger.Logger
}

func NewAPI() *APIGateway {
	return &APIGateway{
		log: logger.WithNamespace("ginger.api"),
	}
}

// Display help string
func (a *APIGateway) Help() string {
	return COMMAND_HEADER + `
api - AWS APIGateway management command.

Usage:
  $ ginger api [operation] [options]

Operation:
  create  : Create new endpoint
  delete  : Delete endpoint
  invoke  : Invoke endpoint
  deploy  : Deploy endpoint
  list    : List endpoint
  help    : Show this help

Options:
  -p, --path   : [all] Path name
  -s, --stage  : [invoke] Target stage
  -m, --method : [invoke] Method name (default=GET)
  -b, --body   : [invoke] Request payload (POST/PUT method only)
`
}

// Run runs command with some options
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
	case API_DEPLOY:
		err = NewDeploy().deployAPI(c, ctx)
	case API_LIST:
		err = a.listEndpoint(c, ctx)
	default:
		fmt.Println(a.Help())
	}
}

// createEndpoint creates new resource by supplied path
// -p, --path options is required.
func (a *APIGateway) createEndpoint(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if c.API.Exists(path) {
		r := c.API.Find(path)
		if r.UserDefined {
			return exception("Endpoint %s is already exists.\n", path)
		}
		r.UserDefined = true
		a.log.Infof("API for path %s created successfully!\n", path)
		return nil
	}

	r := entity.NewResource("", path)
	r.UserDefined = true
	c.API.Resources = append(c.API.Resources, r)
	a.log.Infof("API for path %s created successfully!\n", path)
	return nil
}

// deleteEndpoint deletes resource by supplied path.
// Note that if sub-path exists on target path, those paths also will be deleted.
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
	if !rs.UserDefined {
		return exception("Endpoint %s does not defined.\n", path)
	}

	if !ctx.Has("force") && !input.Bool("Subpath also removed. Are you sure?") {
		a.log.Warn("Aborted.")
		return nil
	}

	api := request.NewAPIGateway(c)
	if !api.ResourceExists(c.API.RestId, rs.Id) {
		return exception("Rsource for %s not found on AWS.\n", path)
	} else if err := api.DeleteResource(c.API.RestId, rs.Id); err != nil {
		a.log.Error("Failed to delete from AWS. Please delete manually.")
	}
	c.API.Remove(path)
	// Remove recursive
	for _, r := range c.API.Resources {
		if !strings.HasPrefix(r.Path, rs.Path) {
			continue
		}
		api.DeleteResource(c.API.RestId, r.Id)
		c.API.Remove(r.Path)
	}
	a.log.Info("Endpoint deleted successfully.")
	return nil
}

// invokeEndpoint invokes resource with HTTP request.
// Ensure some lambda fucntion integartion has set to handle request. Otherwise request will be failed.
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

	// Build request
	host := fmt.Sprintf("%s.execute-api.%s.amazonaws.com", c.API.RestId, c.Project.Region)
	callUrl := fmt.Sprintf("https://%s/%s%s/_invoke", host, ctx.String("stage"), path)

	method := "GET"
	if m := ctx.String("method"); m != "" {
		method = strings.ToUpper(m)
	}

	a.log.Printf("Send HTTP request to %s\n", callUrl)
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
		a.log.Info("========== Response received =========")
		fmt.Println(string(dump))
	}
	return nil
}

// listEndpoint shows endpoint list
func (a *APIGateway) listEndpoint(c *config.Config, ctx *args.Context) error {
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
	for i, r := range c.API.Resources {
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
		if i != len(c.Functions)-1 {
			fmt.Println(strings.Repeat("-", w))
		}
	}
	return nil
}
