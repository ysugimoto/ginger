package command

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/ysugimoto/ginger/assets"
	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
	"github.com/ysugimoto/go-args"
)

const (
	FUNCTION_CREATE = "create"
	FUNCTION_DELETE = "delete"
	FUNCTION_INVOKE = "invoke"
	FUNCTION_CONFIG = "config"
)

type Function struct {
	Command
	log *logger.Logger
}

func NewFunction() *Function {
	return &Function{
		log: logger.WithNamespace("ginger.function"),
	}
}

func (f *Function) Help() string {
	return `
Usage:
  $ ginger fn [operation] [options]

Operation:
  create : Create new function
  delete : Delete function
  invoke : Invoke function

Options:
  -n, --name    : [all] Function name (required)
  -e, --event   : [create] Purpose of function event [s3|apigateway]
  -e, --event   : [invoke] Event source (JSON string) or "@file" for filename
  -m, --memory  : [config] Memory size configuration (must be a multiple of 64 MB)
  -t, --timeout : [config] Function timeout configuration
`
}

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
	default:
		fmt.Println(COMMAND_HEADER + f.Help())
	}
}

func (f *Function) createFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if c.Functions.Exists(name) {
		return exception("Function already defined.")
	}
	fn := &entity.Function{
		Name: name,
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
		if input.Bool("Function exists on AWS. Also delete from there?") {
			if err := lambda.DeleteFunction(name); err != nil {
				f.log.Error("Failed to delete from AWS. Please delete manually.")
			}
		}
	}
	f.log.Printf("Deleting files...")
	if err := os.RemoveAll(filepath.Join(c.FunctionPath, name)); err != nil {
		return exception("Delete dierectory error: %s", err.Error())
	}
	c.Functions = c.Functions.Remove(name)
	f.log.Infof("Function deleted successfully.")
	return nil
}

func (f *Function) configFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if !c.Functions.Exists(name) {
		return exception("Function %s does not defined.", name)
	}

	fn := c.Functions.Find(name)
	fn.MemorySize = int64(ctx.Int("memory"))
	fn.Timeout = int64(ctx.Int("timeout"))
	lambda := request.NewLambda(c)
	return lambda.UpdateFunctionConfiguration(fn)
}

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
