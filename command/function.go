package command

import (
	"fmt"
	"os"

	"io/ioutil"
	"path/filepath"

	"github.com/ysugimoto/ginger/assets"
	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
	"github.com/ysugimoto/go-args"
)

const (
	FUNCTION_CREATE = "create"
	FUNCTION_DELETE = "delete"
)

type Function struct {
	Command
	log *logger.Logger
}

func NewFunction() {
	return &Function{
		log: logger.WithNamespace("ginger.function"),
	}
}

func (f *Function) Help() string {
	return `
ginger fn [subcommand] [options]

Subcommand:
  create: Create new function
  delete: Delete function

Options:
  -n, name: [Required] Function name
`
}

func (f *Function) Run(ctx *args.Context) error {
	c, err := config.Load()
	if err != nil || !c.Exists() {
		f.log.Error("Configuration file could not load. Run `ginger init` before.")
		return nil
	}
	switch ctx.At(1) {
	case FUNCTION_CREATE:
		return f.createFunction(c, ctx)
	case FUNCTION_DELETE:
		return f.deleteFunction(c, ctx)
	case FUNCTION_INVOKE:
		return f.invokeFunction(c, ctx)
	case FUNCTION_UPDATE:
		return f.updateFunction(c, ctx)
	default:
		fmt.Println(f.Help())
		return nil
	}
}

func (f *Function) createFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.At(1)
	if name == "" {
		f.log.Error("Function name didn't supplied. Run with --name option.")
		return nil
	} else if c.Functions.Exists(name) {
		f.log.Error("Function already defined.")
		return nil
	}
	fn := &entity.Function{
		Name: name,
	}
	fnPath := filepath.Join(c.FunctionPath, name)
	if err := os.Mkdir(fnPath, 0755); err != nil {
		f.log.Errorf("Couldn't create directory: %s", fnPath)
		return nil
	}
	err := ioutil.WriteFile(
		filepath.Join(fnPahth, "main.go"),
		f.buildTemplate(name, ctx.String("event")),
		0644,
	)
	if err != nil {
		f.log.Errorf("Create function error: %s", err.Error())
		return nil
	}

	c.Functions = append(c.Functions, fn)
	c.Write()
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
	return []byte(fmt.Sprintf(tmpl, binds...))
}

func (f *Function) deleteFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.At(1)
	if name == "" {
		f.log.Error("Function name didn't supplied. Run with --name option.")
		return nil
	} else if !c.Functions.Exists(name) {
		f.log.Error("Function not defined.")
		return nil
	}
	fn := c.Functions.Find(name)
	lambda := request.NewLambda()
	if lambda.FunctionExists(name) {
		if input.Bool("Function exists on AWS. Also delete from there?") {
			if err := lambda.DeleteFunction(name); err != nil {
				f.log.Error("Failed to delete from AWS. Please delete manually.")
			}
		}
	}
	if err := os.RemoveAll(filepath.Join(c.FunctionPath, name)); err != nil {
		f.log.Errorf("Delete dierectory error: %s", err)
	}
	return nil
}
