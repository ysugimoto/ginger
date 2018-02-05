package command

import (
	"bytes"
	"fmt"
	"os"

	"archive/zip"
	"io/ioutil"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
	"github.com/ysugimoto/go-args"
)

const (
	DEPLOY_FUNCTION = "function"
	DEPLOY_FN       = "fn"
	DEPLOY_API      = "api"
)

type Deploy struct {
	Command
	log *logger.Logger
}

func NewDeploy() *Deploy {
	return &Deploy{
		log: logger.WithNamespace("ginger.deploy"),
	}
}

func (d *Deploy) Help() string {
	return `
	ginger deploy [subcommand] [options]

Subcommand:
  function: Deploy functions (default: all, one of function if --name option supplied)
  api:      Deploy apis (default: all, one of api if --name option supplied)

Options:
  --all:  Deploy all functions/apis
  --name: target fucntion/api name
`
}

func (d *Deploy) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		d.log.Error("Configuration file could not load. Run `ginger init` before.")
		return nil
	}
	switch ctx.At(1) {
	case DEPLOY_FUNCTION, DEPLOY_FN:
		d.log.AddNamespace("function")
		return d.deployFunction(c, ctx)
	case DEPLOY_API:
		d.log.AddNamespace("api")
		return d.deployAPI(c, ctx)
	default:
		if ctx.Has("all") {
			d.log.AddNamespace("all")
			d.deployFunction(c, ctx)
			d.deployAPI(c, ctx)
		} else {
			fmt.Println(d.Help())
		}
		return nil
	}
}

func (d *Deploy) deployFunction(c *config.Config, ctx *args.Context) error {
	targets := c.Functions
	if ctx.Has("name") {
		name := ctx.String("name")
		if !c.Functions.Exists(name) {
			d.log.Errorf("Target function %s doesn't exist", name)
			return nil
		}
		targets = entity.Functions{c.Functions.Find(name)}
	}

	buildDir, err := ioutil.TempDir("", "ginger-builds")
	if err != nil {
		return err
	}

	// Build functions
	defer os.RemoveAll(buildDir)
	builder := newBuilder(c.FunctionPath, buildDir)
	binaries := builder.build(targets)

	// Deploy to AWS
	lambda := request.NewLambda()
	for fn, binary := range binaries {
		// Create zip buffer
		buf := new(bytes.Buffer)
		z := zip.NewWriter(buf)
		bin, _ := ioutil.ReadFile(binary)
		if file, err := z.Create(fn.Name); err != nil {
			return err
		} else if _, err := file.Write(bin); err != nil {
			return err
		}
		if err := z.Close(); err != nil {
			return err
		}

		if err := lambda.DeployFunction(fn.Name, buf.Bytes()); err != nil {
			d.log.Error("Failed to deploy function", err)
		}
	}
	d.log.Infof("Function deployed successfully! %d functions has deployed", len(binaries))
	return nil
}

func (d *Deploy) deployAPI(c *config.Config, ctx *args.Context) error {
	return nil
}
