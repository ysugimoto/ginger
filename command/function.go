package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"io/ioutil"
	"os/signal"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/mattn/go-tty"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/assets"
	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/internal/util"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
)

const (
	FUNCTION_CREATE  = "create"
	FUNCTION_DELETE  = "delete"
	FUNCTION_INVOKE  = "invoke"
	FUNCTION_DEPLOY  = "deploy"
	FUNCTION_MOUNT   = "mount"
	FUNCTION_UNMOUNT = "unmount"
	FUNCTION_LIST    = "list"
	FUNCTION_HELP    = "help"
	FUNCTION_LOG     = "log"
)

// Function is the struct of AWS Lambda function operation command.
// This struct will be dispatched on "ginger fn/funtion" subcommand.
// This command operates with above constant string.
type Function struct {
	Command
	log *logger.Logger
}

func NewFunction() *Function {
	return &Function{
		log: logger.WithNamespace("ginger.function"),
	}
}

// Show function command help.
func (f *Function) Help() string {
	return COMMAND_HEADER + `
funtion - (AWS Lambda) management command.

Usage:
  $ ginger fn [operation] [options]

Operation:
  create  : Create new function
  delete  : Delete function
  invoke  : Invoke function
  mount   : Mount function to destination path
  unmount : Unmount function from destination path
  deploy  : Deploy functions
  list    : List functions
  log     : Tail function log
  help    : Show this help

Options:
  -n, --name    : [all] Function name
  -e, --event   : [create] Purpose of function event [s3|apigateway]
  -e, --event   : [invoke] Event source (JSON string) or "@file" for filename
  -p, --path    : [mount] Path name
      --method  : [mount] Method name to integration
`
}

// Run the command.
func (f *Function) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		f.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	var err error
	defer func() {
		if err != nil {
			f.log.Error(err.Error())
			debugTrace(err)
		}
		c.Write()
	}()

	switch ctx.At(1) {
	case FUNCTION_CREATE:
		err = f.createFunction(c, ctx)
	case FUNCTION_DELETE:
		err = f.deleteFunction(c, ctx)
	case FUNCTION_INVOKE:
		err = f.invokeFunction(c, ctx)
	case FUNCTION_DEPLOY:
		err = NewDeploy().deployFunction(c, ctx)
	case FUNCTION_MOUNT:
		err = f.mountFunction(c, ctx)
	case FUNCTION_LIST:
		err = f.listFunction(c, ctx)
	case FUNCTION_LOG:
		err = f.logFunction(c, ctx)
	default:
		fmt.Println(f.Help())
	}
}

// createFunction creates new function in local.
//
// >>> doc
//
// ## Create new function
//
// Create new lambda function.
//
// ```
// $ ginger function create [options]
// ```
//
// | option  | description                                                                                              |
// |:-------:|:--------------------------------------------------------------------------------------------------------:|
// | --name  | Function name. If this option isn't supplied, ginger will ask it                                         |
// | --event | Function event source. function template switches by this option. enable values are `s3` or `apigateway` |
//
// <<< doc
func (f *Function) createFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = input.String("Type function name")
	}

	if _, err := c.LoadFunction(name); err == nil {
		return exception("Function \"%s\" already defined.", name)
	}

	m := ctx.Int("memory")
	if m < 128 {
		return exception("Memory size must be greater than 128.")
	} else if m%64 > 0 {
		return exception("Memory size must be multiple of 64.")
	}

	event := ctx.String("event")
	if event == "" {
		event = input.Choice("What service will you handle?", []string{
			"(None)",
			"API Gateway",
			"S3",
			"CloudWatch Event",
		})
	}

	fn := &entity.Function{
		Name:       name,
		MemorySize: int64(m),
		Timeout:    int64(ctx.Int("timeout")),
		Role:       c.DefaultLambdaRole,
	}
	fnPath := filepath.Join(c.FunctionPath, name)
	if err := os.Mkdir(fnPath, 0755); err != nil {
		return exception("Couldn't create directory: %s", fnPath)
	}
	if err := ioutil.WriteFile(
		filepath.Join(c.FunctionPath, name, "main.go"),
		f.buildTemplate(name, event),
		0644,
	); err != nil {
		return exception("Create function error: %s", err.Error())
	}

	c.Queue[name] = fn
	f.log.Infof("Function \"%s\" created successfully.\n", name)
	return nil
}

// buildTemplate makes lambda function boilterplace from supplied arguments.
func (f *Function) buildTemplate(name, eventSource string) []byte {
	tmpl, _ := assets.Asset("main.go.template")
	binds := []interface{}{}

	switch eventSource {
	case "API Gateway":
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			strcase.ToCamel(name),
			"request events.APIGatewayProxyRequest",
			"(events.APIGatewayProxyResponse, error)",
			`events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{"X-Ginger-Response": "succeed"},
		Body: "Hello, ginger lambda!",
	}, nil`,
			strcase.ToCamel(name),
		)
	case "S3":
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			strcase.ToCamel(name),
			"s3Event events.S3Event",
			"error",
			"nil",
			strcase.ToCamel(name),
		)
	case "CloudWatch Event":
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			strcase.ToCamel(name),
			"cloudWatchEvent events.CloudWatchEvent",
			"error",
			"nil",
			strcase.ToCamel(name),
		)
	default:
		binds = append(binds,
			"",
			strcase.ToCamel(name),
			"input map[string]string",
			"(map[string]string, error)",
			"input, nil",
			strcase.ToCamel(name),
		)
	}
	return []byte(fmt.Sprintf(string(tmpl), binds...))
}

// deleteFunction deletes function.
// If function has been deployed on AWS Lambda, also delete it.
//
// >>> doc
//
// ## Delete function
//
// Delete lambda function.
//
// ```
// $ ginger function delete [options]
// ```
//
// | option  | description              |
// |:-------:|:------------------------:|
// | --name  | [Required] function name |
//
// <<< doc
func (f *Function) deleteFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	}
	_, err := c.LoadFunction(name)
	if err != nil {
		return exception(err.Error())
	}

	f.log.Printf("Deleting function: %s\n", name)
	lambda := request.NewLambda(c)
	f.log.Print("Checking lambda function existence...")
	if lambda.FunctionExists(name) {
		if ctx.Has("force") || input.Bool("Function exists on AWS. Also delete from there?") {
			if err := lambda.DeleteFunction(name); err != nil {
				f.log.Error("Failed to delete from AWS. Please delete manually.")
			}
		}
	} else {
		f.log.Print("Not found in AWS. Skip it.")
	}

	f.log.Print("Deleting files...")
	if err := os.RemoveAll(filepath.Join(c.FunctionPath, name)); err != nil {
		return exception("Delete dierectory error: %s", err.Error())
	}
	c.DeleteFunction(name)
	f.log.Infof("Function \"%s\" deleted successfully.\n", name)
	return nil
}

// invokeFunction invokes lambda function which deployed in AWS.
// Make sure function is deployed to AWS before call it.
//
// >>> doc
//
// ## Invoke function
//
// Invoke lambda function.
//
// ```
// $ ginger function invoke [options]
// ```
//
// | option  | description                                                                                           |
// |:-------:|:-----------------------------------------------------------------------------------------------------:|
// | --name  | [Required] function name                                                                              |
// | --event | Passing event source data. data must be JSON format string, or can specify file name like `@filename` |
//
// <<< doc
func (f *Function) invokeFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseFunction()
	}
	if _, err := c.LoadFunction(name); err != nil {
		return exception("Function \"%s\" could not find.", name)
	}

	lambda := request.NewLambda(c)
	if !lambda.FunctionExists(name) {
		return exception("Function %s couldn't find in AWS. Please deploy it before invoke.", name)
	}

	var payload []byte
	src := ctx.String("event")
	if src != "" {
		if strings.HasPrefix(src, "@") {
			srcFile := src[1:]
			if _, err := os.Stat(srcFile); err != nil {
				return exception("Event source file %s doesn't exist.")
			}
			payload, _ = ioutil.ReadFile(srcFile)
		} else {
			payload = []byte(src)
		}
	}

	lambda.InvokeFunction(name, payload)
	return nil
}

// listFunction shows registered functions.
//
// >>> doc
//
// ## List function
//
// List registered lambda functions.
//
// ```
// $ ginger function list
// ```
//
// <<< doc
func (f *Function) listFunction(c *config.Config, ctx *args.Context) error {
	functions, _ := c.LoadAllFunctions()
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
	fmt.Printf("%-36s %-16s %-16s %-4s\n", "name", "memory", "timeout", "deployed")
	fmt.Println(line)
	for i, fn := range functions {
		d := "no"
		if fn.Arn != "" {
			d = "yes"
		}
		fmt.Printf("%-36s %-16s %-16s %-4s\n", fn.Name, fmt.Sprintf("%d MB", fn.MemorySize), fmt.Sprintf("%d sec", fn.Timeout), d)
		if i != len(functions)-1 {
			fmt.Println(strings.Repeat("-", w))
		}
	}
	return nil
}

// logFunction displays tailing logs via CloudWatch.
//
// >>> doc
//
// ## Log function
//
// Tailing log function output via CloudWatch Log.
//
// ```
// $ ginger function log [options]
// ```
//
// | option  | description                                                                                              |
// |:-------:|:--------------------------------------------------------------------------------------------------------:|
// | --name  | Function name. If this option isn't supplied, ginger will ask it                                         |
//
// <<< doc
func (f *Function) logFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseFunction()
	}
	if _, err := c.LoadFunction(name); err != nil {
		return exception("Function \"%s\" could not find.", name)
	}

	f.log.Warnf("Tailing cloudwatch logs for function \"%s\"...\n", name)
	ctc, cancel := context.WithCancel(context.Background())
	cwl := request.NewCloudWatch(c)
	go cwl.TailLogs(
		ctc,
		fmt.Sprintf("/aws/lambda/%s", name),
		ctx.String("filter"),
	)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-ch:
		cancel()
	}
	return nil
}

// mountFunction makes integration between API Gateway resource and Lambda function
//
// >>> doc
//
// ## Mount function
//
// Create function integration to destination resource.
//
// ```
// $ ginger function mount [options]
// ```
//
// | option   | description                                                      |
// |:--------:|:----------------------------------------------------------------:|
// | --name   | Function name. If this option isn't supplied, ginger will ask it |
// | --path   | Resource path. If this option isn't supplied, ginger will ask it |
// | --method | Integration method                                               |
//
// <<< doc
func (f *Function) mountFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseFunction()
	}
	if _, err := c.LoadFunction(name); err != nil {
		return exception("Function %s could't find in your project.", name)
	}

	path := ctx.String("path")
	if path == "" {
		path = c.ChooseResource()
	}
	rs, err := c.LoadResource(path)
	if err != nil {
		return exception("Endpoint %s couldn't find in your project.\n", path)
	}

	var method string
	if m := ctx.String("method"); m != "" {
		method = strings.ToUpper(m)
	} else {
		method = util.ChooseMethod("ANY")
	}

	ig := rs.GetIntegration(method)
	if ig == nil {
		ig = entity.NewIntegration("lambda", name, rs.Path)
		rs.AddIntegration(method, ig)
		f.log.Infof("Function %s mouted to resource %s.\n", name, path)
		return nil
	}
	switch ig.IntegrationType {
	case "lambda":
		return exception("Resource already mounted as lambda. Unmount before use it.")
	case "s3":
		return exception("Resource already mounted as storage. Unmount before use it.")
	}
	return exception("Undefined integration type.")
}

// unmountFunction makes integration between API Gateway resource and Lambda function
//
// >>> doc
//
// ## Mount function
//
// Delete function integration to destination resource.
//
// ```
// $ ginger function unmount [options]
// ```
//
// | option   | description                                                      |
// |:--------:|:----------------------------------------------------------------:|
// | --path   | Resource path. If this option isn't supplied, ginger will ask it |
// | --method | Integration method                                               |
//
// <<< doc
func (f *Function) unmountFunction(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		path = c.ChooseResource()
	}
	rs, err := c.LoadResource(path)
	if err != nil {
		return exception("Endpoint %s couldn't find in your project.\n", path)
	}

	var method string
	if m := ctx.String("method"); m != "" {
		method = strings.ToUpper(m)
	} else {
		method = util.ChooseMethod("ANY")
	}

	ig := rs.GetIntegration(method)
	if ig == nil {
		return exception("No integration found, Abort.")
	}
	switch ig.IntegrationType {
	case "s3":
		return exception("Storage already integrated. to remove it, run 'ginger storage unmount'.")
	case "lambda":
		if rs.Id != "" {
			api := request.NewAPIGateway(c)
			api.DeleteMethod(c.RestApiId, rs.Id, method)
			api.DeleteIntegration(c.RestApiId, rs.Id, method)
		}
		rs.DeleteIntegration(method)
	default:
		return exception("Undefined integration type.")
	}
	f.log.Infof("Function unmounted for resource %s.\n", path)
	return nil
}
