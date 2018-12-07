package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"io/ioutil"
	"os/exec"
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
	// Subcommands
	FUNCTIONCREATE  = "create"
	FUNCTIONDELETE  = "delete"
	FUNCTIONINVOKE  = "invoke"
	FUNCTIONDEPLOY  = "deploy"
	FUNCTIONMOUNT   = "mount"
	FUNCTIONUNMOUNT = "unmount"
	FUNCTIONLIST    = "list"
	FUNCTIONHELP    = "help"
	FUNCTIONLOG     = "log"
	FUNCTIONBUILD   = "build"
	FUNCTIONTEST    = "test"
	FUNCTIONRUN     = "run"

	// Event names
	eventNameNone       = "(None)"
	eventNameAPIGateway = "API Gateway"
	eventNameS3         = "S3"
	eventNameCloudWatch = "CloudWatch Event"
	eventNameSQS        = "SQS Event"
	eventNameKinesis    = "Kinesis Event"

	// Event source filename
	eventSourceFileName   = "event.json"
	clientContextFileName = "context.json"
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
	return commandHeader() + `
function - (AWS Lambda) management command.

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
  build   : Build function
  test    : Run unit test
  run     : Run function on local
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
func (f *Function) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		f.log.Error("Configuration file could not load. Run `ginger init` before.")
		return errors.New("")
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
	case FUNCTIONCREATE:
		err = f.createFunction(c, ctx)
	case FUNCTIONDELETE:
		err = f.deleteFunction(c, ctx)
	case FUNCTIONINVOKE:
		err = f.invokeFunction(c, ctx)
	case FUNCTIONDEPLOY:
		err = NewDeploy().deployFunction(c, ctx)
	case FUNCTIONMOUNT:
		err = f.mountFunction(c, ctx)
	case FUNCTIONLIST:
		err = f.listFunction(c, ctx)
	case FUNCTIONLOG:
		err = f.logFunction(c, ctx)
	case FUNCTIONBUILD:
		err = f.buildFunction(c, ctx)
	case FUNCTIONTEST:
		err = f.testFunction(c, ctx)
	case FUNCTIONRUN:
		err = f.runFunction(c, ctx)
	default:
		fmt.Println(f.Help())
	}
	return err
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
// |:-------:|:---------------------------------------------------------------------------------------------------------|
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
			eventNameNone,
			eventNameAPIGateway,
			eventNameS3,
			eventNameCloudWatch,
			eventNameSQS,
			eventNameKinesis,
		})
	}

	fn := &entity.Function{
		Name:        name,
		MemorySize:  int64(m),
		Timeout:     int64(ctx.Int("timeout")),
		Role:        c.DefaultLambdaRole,
		Environment: make(map[string]*string),
	}
	fnPath := filepath.Join(c.FunctionPath, name)
	if err := os.Mkdir(fnPath, 0755); err != nil {
		return exception("Couldn't create directory: %s", fnPath)
	}
	// Create main.go from template
	if err := ioutil.WriteFile(
		filepath.Join(c.FunctionPath, name, "main.go"),
		f.buildTemplate(name, event),
		0644,
	); err != nil {
		return exception("Create function error: %s", err.Error())
	}
	// Create event.json from template
	if err := ioutil.WriteFile(
		filepath.Join(c.FunctionPath, name, "event.json"),
		f.buildEventJson(event),
		0644,
	); err != nil {
		return exception("Create event.json error: %s", err.Error())
	}

	c.Queue[name] = fn
	f.log.Infof("Function \"%s\" created successfully.\n", name)
	return nil
}

// buildTemplate makes lambda function boilterplace from supplied arguments.
func (f *Function) buildTemplate(name, eventSource string) []byte {
	tmpl, _ := assets.Assets.Open("/main.go.template")
	b := new(bytes.Buffer)

	io.Copy(b, tmpl)
	binds := []interface{}{}
	camelName := strcase.ToCamel(name)

	switch eventSource {
	case eventNameAPIGateway:
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			camelName,
			"request events.APIGatewayProxyRequest",
			"(events.APIGatewayProxyResponse, error)",
			`events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{"X-Ginger-Response": "succeed"},
		Body: "Hello, ginger lambda!",
	}, nil`,
			camelName,
		)
	case eventNameS3:
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			camelName,
			"s3Event events.S3Event",
			"error",
			"nil",
			camelName,
		)
	case eventNameCloudWatch:
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			camelName,
			"cloudWatchEvent events.CloudWatchEvent",
			"error",
			"nil",
			camelName,
		)
	case eventNameSQS:
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			camelName,
			"sqsEvent events.SQSEvent",
			"error",
			"nil",
			camelName,
		)
	case eventNameKinesis:
		binds = append(binds,
			"\n\t\"github.com/aws/aws-lambda-go/events\"",
			camelName,
			"kinesisEvent events.KinesisEvent",
			"error",
			"nil",
			camelName,
		)
	default:
		binds = append(binds,
			"",
			camelName,
			"input map[string]string",
			"(map[string]string, error)",
			"input, nil",
			camelName,
		)
	}
	return []byte(fmt.Sprintf(b.String(), binds...))
}

// buildEventJson makes event.json for test from supplied arguments.
func (f *Function) buildEventJson(eventSource string) []byte {
	assetPath := "/events/default.json"
	switch eventSource {
	case eventNameAPIGateway:
		assetPath = "/events/apigateway.json"
	case eventNameS3:
		assetPath = "/events/s3.json"
	case eventNameCloudWatch:
		assetPath = "/events/cloudwatch.json"
	case eventNameSQS:
		assetPath = "/events/sqs.json"
	case eventNameKinesis:
		assetPath = "/events/kinesis.json"
	}
	src, _ := assets.Assets.Open(assetPath)
	b := new(bytes.Buffer)
	io.Copy(b, src)
	return b.Bytes()
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
// |:-------:|:-------------------------|
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
// |:-------:|:------------------------------------------------------------------------------------------------------|
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
// |:-------:|:---------------------------------------------------------------------------------------------------------|
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
// |:--------:|:-----------------------------------------------------------------|
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
		return exception("Function %s couldn't find in your project.", name)
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
// ## Unmount function
//
// Delete function integration to destination resource.
//
// ```
// $ ginger function unmount [options]
// ```
//
// | option   | description                                                      |
// |:--------:|:-----------------------------------------------------------------|
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

// buildFunction builds Lambda function from go source files to binary
//
// >>> doc
//
// ## Build function
//
// Build function binary for your runtime.
//
// ```
// $ ginger function build [options]
// ```
//
// | option   | description                                                      |
// |:--------:|:-----------------------------------------------------------------|
// | --name   | Target function name                                             |
//
// <<< doc
func (f *Function) buildFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseFunction()
	}
	if _, err := c.LoadFunction(name); err != nil {
		return exception("Function %s couldn't find in your project.", name)
	}

	arguments := []string{"-o", filepath.Join(c.FunctionPath, name, name)}
	if err := execGoCommand(context.Background(), c, name, "build", arguments); err != nil {
		return exception("Failed to build %s function: %s\n", name, err.Error())
	}
	f.log.Infof("Function %s built successfully.\n", name)
	return nil
}

// testFunction run tests with ginger's environment
//
// >>> doc
//
// ## Test function
//
// Run test on project environment.
//
// ```
// $ ginger function test [options]
// ```
//
// | option   | description                                                      |
// |:--------:|:-----------------------------------------------------------------|
// | --name   | Target function name                                             |
//
// <<< doc
func (f *Function) testFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseFunction()
	}
	if _, err := c.LoadFunction(name); err != nil {
		return exception("Function %s couldn't find in your project.", name)
	}

	return execGoCommand(context.Background(), c, name, "test", nil)
}

// runFunction invokes lambda function locally
//
// >>> doc
//
// ## Run function
//
// Run Lambda function locally.
// The `--event` argument accepts event payload of JSON file. In default, use (function-directory)/event.json file if exists. Or, you can use template JSON which corredspons to event source in ginger bundled following:
//  - s3
//  - apigateway
//  - sqs
//  - kinesis
//  - cloudwatch
// For example, you can run function with s3 event source as:
//
// ```
// $ ginger fn run --name example-function --event s3
// ```
//
// And, additional client context data also can provide. put (function-directory)/context.json and defined some JSON values.
//
// ```
// $ ginger function run [options]
// ```
//
// | option   | description                                                      |
// |:--------:|:-----------------------------------------------------------------|
// | --name   | Target function name                                             |
// | --event  | Event payload JSON file path                                     |
//
// <<< doc
func (f *Function) runFunction(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseFunction()
	}
	fn, err := c.LoadFunction(name)
	if err != nil {
		return exception("Function %s couldn't find in your project.", name)
	}

	tmpDir, err := ioutil.TempDir("", "ginger-local-invoke")
	if err != nil {
		return exception("Failed to create temporary directory: %s", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	// Build binary and put to temprary directory
	bin := filepath.Join(tmpDir, name)
	if err := execGoCommand(context.Background(), c, name, "build", []string{"-o", bin}); err != nil {
		return exception("Failed to build %s binary: %s ", name, err.Error())
	}
	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start up Lambda RPC inside goroutine
	// We run server by using built binary because `go run main.go` process cannot kill its process properly.
	// The `go run main.go` makes temporary binary and run at `/var/folders/xxxxx/exe/main`,
	// and if we kill process via cmd.Process.Kill(), then that process won't kill, so RPC process runs forever.
	go func() {
		f.log.Infof("Starting local %s Lambda...\n", name)
		cmd := exec.CommandContext(parentCtx, bin)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = buildEnv(map[string]string{
			"_LAMBDA_SERVER_PORT": LAMBDARPCPORT,
		})
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
		f.log.Infof("Shutting down local %s Lambda...\n", name)
	}()

	// Factory source JSON
	source := []byte("{}")
	event := ctx.String("event")
	if event == "" {
		// If event isn't supplied, try to retrieve function directory's event.json
		eventFile := filepath.Join(c.FunctionPath, name, eventSourceFileName)
		if _, err := os.Stat(eventFile); err == nil {
			if buf, err := ioutil.ReadFile(eventFile); err == nil {
				source = buf
			}
		}
	} else if src, err := assets.Assets.Open(fmt.Sprintf("/events/%s.json", event)); err == nil {
		// If event supplied as "event template name", use template JSON from compiled assets
		buf := new(bytes.Buffer)
		io.Copy(buf, src)
		source = buf.Bytes()
	}

	// Use context data if exsits
	clientContext := []byte{}
	contextFile := filepath.Join(c.FunctionPath, name, clientContextFileName)
	if _, err := os.Stat(contextFile); err == nil {
		if buf, err := ioutil.ReadFile(contextFile); err == nil {
			clientContext = buf
		}
	}

	// Wait until lambda RPC server has been started (maybe a second is enough)
	time.Sleep(1 * time.Second)
	resp, err := execLambdaRPC(fn.Timeout, source, clientContext)
	if err != nil {
		return exception("Failed to call Lambda RPC: %s", err.Error())
	}
	if resp.Error != nil {
		f.log.Errorf("Lambda responded error:\nType: %s\nMessage: %s\n", resp.Error.Type, resp.Error.Message)
		if len(resp.Error.StackTrace) > 0 {
			f.log.Warn("StackTrace")
			for _, frame := range resp.Error.StackTrace {
				f.log.Warnf("%s at line %d: %s\n", frame.Path, frame.Line, frame.Label)
			}
		}
		return exception("Failed to run lambda function")
	} else if resp.Payload != nil {
		f.log.Printf("payload received:\n%s\n", string(resp.Payload))
	}
	return nil
}
