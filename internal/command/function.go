package command

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/mattn/go-tty"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/internal/assets"
	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/entity"
	"github.com/ysugimoto/ginger/internal/input"
	"github.com/ysugimoto/ginger/internal/logger"
	"github.com/ysugimoto/ginger/internal/request"
)

const (
	FUNCTION_CREATE = "create"
	FUNCTION_DELETE = "delete"
	FUNCTION_INVOKE = "invoke"
	FUNCTION_CONFIG = "config"
	FUNCTION_DEPLOY = "deploy"
	FUNCTION_LIST   = "list"
	FUNCTION_HELP   = "help"
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
	return `
funtion - (AWS Lambda) management command.

Usage:
  $ ginger fn [operation] [options]

Operation:
  create : Create new function
  delete : Delete function
  invoke : Invoke function
  config : Modify function setting
  deploy : Deploy functions
  list   : List functions
  help   : Show this help

Options:
  -n, --name    : [all] Function name (required)
  -e, --event   : [create] Purpose of function event [s3|apigateway]
  -e, --event   : [invoke] Event source (JSON string) or "@file" for filename
  -m, --memory  : [config] Memory size configuration (must be a multiple of 64 MB)
  -t, --timeout : [config] Function timeout configuration
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
	case FUNCTION_CONFIG:
		err = f.configFunction(c, ctx)
	case FUNCTION_DEPLOY:
		err = NewDeploy().deployFunction(c, ctx)
	case FUNCTION_LIST:
		err = f.listFunction(c, ctx)
	default:
		fmt.Println(COMMAND_HEADER + f.Help())
	}
}

// createFunction creates new function in local.
func (f *Function) createFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if c.Functions.Exists(name) {
		return exception("Function already defined.")
	}

	m := ctx.Int("memory")
	if m < 128 {
		return exception("Memory size must be greater than 128.")
	} else if m%64 > 0 {
		return exception("Memory size must be multiple of 64.")
	}
	fn := &entity.Function{
		Name:       name,
		MemorySize: int64(m),
		Timeout:    int64(ctx.Int("timeout")),
	}
	fnPath := filepath.Join(c.FunctionPath, name)
	if err := os.Mkdir(fnPath, 0755); err != nil {
		return exception("Couldn't create directory: %s", fnPath)
	}
	if err := ioutil.WriteFile(
		filepath.Join(c.FunctionPath, name, "main.go"),
		f.buildTemplate(name, ctx.String("event")),
		0644,
	); err != nil {
		return exception("Create function error: %s", err.Error())
	}

	c.Functions = append(c.Functions, fn)
	f.log.Info("Function created successfully!")
	return nil
}

// buildTemplate makes lambda function boilterplace from supplied arguments.
func (f *Function) buildTemplate(name, eventSource string) []byte {
	tmpl, _ := assets.Asset("main.go.template")
	binds := []interface{}{}

	switch eventSource {
	case "apigateway":
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			strcase.ToCamel(name),
			"request events.APIGatewayProxyRequest",
			"(events.APIGatewayProxyResponse, error)",
			"events.APIGatewayProxyResponse{}, nil",
			strcase.ToCamel(name),
		)
	case "s3":
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			strcase.ToCamel(name),
			"s3Event events.S3Event",
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
func (f *Function) deleteFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if !c.Functions.Exists(name) {
		return exception("Function not defined.")
	}

	f.log.Printf("Deleting function: %s\n", name)
	lambda := request.NewLambda(c)
	f.log.Print("Checking lambda function exintence...")
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
	c.Functions = c.Functions.Remove(name)
	f.log.Info("Function deleted successfully.")
	return nil
}

// configFunction modifies function configuration.
// We can modifies memorysize and timeout.
func (f *Function) configFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if !c.Functions.Exists(name) {
		return exception("Function %s does not defined.", name)
	}

	m := ctx.Int("memory")
	if m < 128 {
		return exception("Memory size must be greater than 128.")
	} else if m%64 > 0 {
		return exception("Memory size must be multiple of 64.")
	}
	fn := c.Functions.Find(name)
	fn.MemorySize = int64(m)
	fn.Timeout = int64(ctx.Int("timeout"))
	lambda := request.NewLambda(c)
	return lambda.UpdateFunctionConfiguration(fn)
}

// invokeFunction invokes lambda function which deployed in AWS.
// Make sure function is deployed to AWS before call it.
func (f *Function) invokeFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if !c.Functions.Exists(name) {
		return exception("Function not defined.")
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
func (f *Function) listFunction(c *config.Config, ctx *args.Context) error {
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
	for i, fn := range c.Functions {
		d := "no"
		if fn.Arn != "" {
			d = "yes"
		}
		fmt.Printf("%-36s %-16s %-16s %-4s\n", fn.Name, fmt.Sprintf("%d MB", fn.MemorySize), fmt.Sprintf("%d sec", fn.Timeout), d)
		if i != len(c.Functions)-1 {
			fmt.Println(strings.Repeat("-", w))
		}
	}
	return nil
}
