package command

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"archive/zip"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
)

const (
	DEPLOYFUNCTION = "function"
	DEPLOYFN       = "fn"
	DEPLOYRESOURCE = "resource"
	DEPLOYR        = "r"
	DEPLOYSTORAGE  = "storage"
	DEPLOYSCHEDULE = "schedule"
	DEPLOYS        = "s"
	DEPLOYALL      = "all"
	DEPLOYHELP     = "help"
)

// Deploy is the struct that manages function and api deployment.
// deploy syncs between local and AWS.
type Deploy struct {
	Command
	log *logger.Logger
}

func NewDeploy() *Deploy {
	return &Deploy{
		log: logger.WithNamespace("ginger.deploy"),
	}
}

// Show deloy command help
func (d *Deploy) Help() string {
	return commandHeader() + `
deploy - Deploy management functions and apis.

Usage:
  $ ginger deploy [subcommand] [options]

Subcommand:
  function : Deploy functions (default: all, one of function if --name option supplied)
  resource : Deploy resources (default: all, one of path if --name option supplied)
  storage  : Deploy storage
  schedule : Deploy schedulers
  all      : Deploy both of functions and apis
  help     : Show this help

Options:
  --name  : Target fucntion name
  --stage : Target api stage
`
}

// Run the deploy command
//
// >>> doc
//
// ## Deploy all
//
// Deploy all functions, resources, storage items.
//
// ```
// $ ginger deploy all [options]
// ```
//
// | option  | description                                                         |
// |:-------:|:--------------------------------------------------------------------|
// | --stage | Stage name. If this option is supplied, create deployment to stage. |
//
// <<< doc
func (d *Deploy) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		d.log.Error("Configuration file could not load. Run `ginger init` before.")
		return errors.New("")
	}
	var err error
	defer func() {
		if err != nil {
			d.log.Error(err.Error())
			debugTrace(err)
		} else {
			d.log.Info("Deployment finished.")
		}
		c.SortResources()
		c.Write()
	}()

	switch ctx.At(1) {
	case DEPLOYFUNCTION, DEPLOYFN:
		if err = d.runHook(c); err != nil {
			return err
		}
		err = d.deployFunction(c, ctx)
	case DEPLOYRESOURCE, DEPLOYR:
		if err = d.runHook(c); err != nil {
			return err
		}
		err = d.deployResource(c, ctx)
	case DEPLOYSCHEDULE, DEPLOYS:
		if err = d.runHook(c); err != nil {
			return err
		}
		err = d.deploySchedulers(c, ctx)
	case DEPLOYSTORAGE:
		if err = d.runHook(c); err != nil {
			return err
		}
		err = d.deployStorage(c, ctx)
	case DEPLOYALL:
		if err = d.runHook(c); err != nil {
			return err
		}
		d.log.Print("========== Function Deployment ==========")
		if err = d.deployFunction(c, ctx); err != nil {
			return err
		}
		d.log.Print("========== Storage Deployment ==========")
		if err = d.deployStorage(c, ctx); err != nil {
			return err
		}
		d.log.Print("========== Scheduler Deployment ==========")
		if err = d.deploySchedulers(c, ctx); err != nil {
			return err
		}
		d.log.Print("========== Resource Deployment ==========")
		if err = d.deployResource(c, ctx); err != nil {
			return err
		}

		if s := ctx.String("stage"); s != "" {
			d.log.Print("========== Stage Deployment ==========")
			if err = d.deployStage(c, ctx); err != nil {
				return err
			}
		}
	default:
		fmt.Println(d.Help())
	}
	return nil
}

// runHook runs deployment hook command
func (d *Deploy) runHook(c *config.Config) error {
	// If deploy hook doesn't spcify, skip it
	if c.DeployHookCommand == "" {
		return nil
	}
	hook := c.DeployHookCommand
	d.log.Infof("Deploy hook command execute: %s\n", hook)
	parts := strings.Split(hook, " ")
	var cmd *exec.Cmd
	if len(parts) > 1 {
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		cmd = exec.Command(parts[0])
	}
	cmd.Dir = c.Root
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// deployFunction deploys functions to AWS Lambda.
//
// >>> doc
//
// ## Deploy functions
//
// Build and deploy lambda functions to AWS.
//
// ```
// $ ginger deploy function [options]
// ```
//
// | option | description                                                       |
// |:------:|:------------------------------------------------------------------|
// | --name | Function name. if this option didn't supply, deploy all functions |
//
// <<< doc
func (d *Deploy) deployFunction(c *config.Config, ctx *args.Context) error {
	d.log.AddNamespace("function")
	defer d.log.RemoveNamespace("function")
	targets := []*entity.Function{}

	if name := ctx.String("name"); name != "" {
		fn, err := c.LoadFunction(name)
		if err != nil {
			return exception("Target function \"%s\" doesn't exist", name)
		}
		targets = append(targets, fn)
	} else {
		var err error
		targets, err = c.LoadAllFunctions()
		if err != nil {
			return exception("Failed to list functions: %s", err.Error())
		}
	}

	if len(targets) == 0 {
		d.log.Warn("No functions found. Skip to deploy to Lambda.")
		return nil
	}

	buildDir, err := ioutil.TempDir("", "ginger-builds")
	if err != nil {
		return exception(err.Error())
	}
	defer os.RemoveAll(buildDir)

	// Validate lambda exection roles
	for _, f := range targets {
		if f.Role != "" {
			continue
		}
		if c.DefaultLambdaRole == "" {
			return exception("Lambda execution role is empty for fucntion \"%s\". Please set default role or function specific role.\n", f.Name)
		} else {
			d.log.Warnf("Function \"%s\" execution role is empty. ginger uses default role via project configuration.\n", f.Name)
		}
	}

	// Build functions
	builder := newBuilder(c.FunctionPath, buildDir, c.LibPath)
	if err := builder.build(targets); err != nil {
		return exception("Failed to build lambda function. %s", err.Error())
	}

	// Deploy to AWS
	lambda := request.NewLambda(c)
	for _, fn := range targets {
		// Check binary existence
		binPath := filepath.Join(buildDir, fn.Name)
		if _, err := os.Stat(binPath); err != nil {
			d.log.Errorf("Build function binary not found for %s. skip it\n", fn.Name)
			continue
		}
		if fn.Role == "" {
			fn.Role = c.DefaultLambdaRole
		}
		d.log.Printf("Archiving zip for %s...\n", fn.Name)
		buffer, err := d.archive(fn, binPath)
		if err != nil {
			d.log.Errorf("Archive error for %s: %s\n", fn.Name, err.Error())
			continue
		}
		d.log.Printf("Deploying function %s to AWS Lambda...\n", fn.Name)
		if arn, err := lambda.DeployFunction(fn, buffer); err == nil {
			d.log.Infof("Function %s deployed successfully!\n", fn.Name)
			fn.Arn = arn
		}
	}
	return nil
}

func (d *Deploy) deploySchedulers(c *config.Config, ctx *args.Context) error {
	cw := request.NewCloudWatch(c)
	lambda := request.NewLambda(c)

	scs, err := c.LoadAllSchedulers()
	if err != nil {
		return exception("Failed to get all schedulers: %s", err.Error())
	} else if len(scs) == 0 {
		d.log.Warn("No schedules found. Skip to deploy CloudWatch.")
		return nil
	}
	for _, sc := range scs {
		arn, err := cw.GetScheduleArn(sc.Name)
		if err != nil {
			return nil
		}
		arn, err = cw.CreateOrUpdateSchedule(sc)
		if err != nil {
			return nil
		}
		if sc.Functions == nil {
			continue
		}
		for _, name := range sc.Functions {
			fn, err := lambda.GetFunction(name)
			if err != nil {
				return nil
			}
			if err := lambda.AddCloudWatchPermission(*fn.FunctionName, arn); err != nil {
				return nil
			}
			if err = cw.PutTarget(sc.Name, fn.FunctionArn); err != nil {
				return nil
			}
		}
	}
	return nil
}

// archive archives built application binary to zip.
func (d *Deploy) archive(fn *entity.Function, binPath string) ([]byte, error) {
	buf := new(bytes.Buffer)
	z := zip.NewWriter(buf)
	bin, err := ioutil.ReadFile(binPath)
	if err != nil {
		return nil, exception("Binary file read error: %s", err.Error())
	}
	header := &zip.FileHeader{
		Name:           fn.Name,
		Method:         zip.Deflate,
		ExternalAttrs:  0777 << 16,
		CreatorVersion: 3 << 8,
	}
	if f, err := z.CreateHeader(header); err != nil {
		return nil, exception("Failed to create zip header: %s", err.Error())
	} else if _, err := f.Write(bin); err != nil {
		return nil, exception("Failed to write binary to zip stream: %s", err.Error())
	} else if err := z.Close(); err != nil {
		return nil, exception("Failed to close zip stream: %s", err.Error())
	}
	return buf.Bytes(), nil
}

// deployAPI deploys resources to AWS APIGateway.
//
// >>> doc
//
// ## Deploy resources
//
// Deploy API Gateway resources to AWS.
//
// ```
// $ ginger deploy resource [options]
// ```
//
// If resource has some integrations, create integration as well.
//
// | option  | description                                                         |
// |:-------:|:--------------------------------------------------------------------|
// | --stage | Stage name. If this option is supplied, create deployment to stage. |
//
// <<< doc
func (d *Deploy) deployResource(c *config.Config, ctx *args.Context) (err error) {
	if len(c.Resources) == 0 {
		d.log.Warn("No resources found. Skip to deploy APIGateway.")
		return nil
	}
	d.log.AddNamespace("resource")
	defer d.log.RemoveNamespace("resource")
	api := request.NewAPIGateway(c)

	if c.RestApiId == "" {
		d.log.Print("REST API hasn't created yet. Creating...")
		c.RestApiId, err = api.CreateRestAPI(fmt.Sprintf("ginger-%s", c.ProjectName))
		if err != nil {
			return
		}
	}

	// Probably root "/" resource created automatically, check existence in local
	var rootId string
	if r, err := c.LoadResource("/"); err != nil {
		rootId, err = api.GetResourceIdByPath(c.RestApiId, "/")
		if err != nil {
			return nil
		}
		rs := entity.NewResource(rootId, "/")
		c.Resources = append(c.Resources, rs)
	} else if r.Id == "" {
		r.Id, err = api.GetResourceIdByPath(c.RestApiId, "/")
		rootId = r.Id
	} else {
		rootId = r.Id
	}

	for _, r := range c.Resources {
		// If "Id" exists, the resource has already been deployed
		if r.Id != "" && api.ResourceExists(c.RestApiId, r.Id) {
			d.log.Infof("Resource %s has already been deployed.\n", r.Path)
		} else {
			api.CreateResourceRecursive(c.RestApiId, r.Path)
		}
		if igs := r.GetIntegrations(); igs != nil {
			for method, integration := range igs {
				if err = api.PutIntegration(c.RestApiId, r.Id, method, integration); err != nil {
					return nil
				}
			}
		}
	}
	// Obviously succeed, returns nil
	return nil
}

// deployStorage deploys storage items to AWS S3.
//
// >>> doc
//
// ## Deploy storage items
//
// Deploy storage files to S3.
//
// ```
// $ ginger deploy storage
// ```
//
// <<< doc
func (d *Deploy) deployStorage(c *config.Config, ctx *args.Context) error {
	d.log.AddNamespace("storage")
	defer d.log.RemoveNamespace("storage")
	bucket := c.S3BucketName
	s3 := request.NewS3(c)

	// Upload local objects to remote
	locals, err := d.listLocalObjects(c.StoragePath)
	if err != nil {
		return exception("Failed to list local storage files: %s", err.Error())
	} else if len(locals) == 0 {
		d.log.Warn("No files found. Skip to deoloy S3.")
		return nil
	}

	d.log.Warn("Deploying storage local -> S3...")

	// Ensure bucket exists on AWS
	if err := s3.EnsureBucketExists(bucket); err != nil {
		return exception("The bucket %s creation error: %s", bucket, err.Error())
	}

	for _, so := range locals {
		d.log.Printf("Uploading local %s -> s3://%s/%s...\n", so.Key, bucket, so.Key)
		s3.PutObject(bucket, so)
	}
	return nil
}

func (d *Deploy) listLocalObjects(root string) ([]*entity.StorageObject, error) {
	objects := make([]*entity.StorageObject, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		} else if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		so := entity.NewStorageObject(rel, info)
		if err = so.Load(path); err != nil {
			return err
		}
		objects = append(objects, so)
		return nil
	})
	return objects, err
}

func (d *Deploy) deployStage(c *config.Config, ctx *args.Context) error {
	d.log.AddNamespace("stage")
	defer d.log.RemoveNamespace("stage")

	name := ctx.String("stage")
	_, err := c.LoadStage(name)
	if err != nil {
		d.log.Warnf("Stage \"%s\" doesn't exists. Create...\n", name)
		fileName := filepath.Join(c.StagePath, fmt.Sprintf("%s.toml", name))
		template := fmt.Sprintf("name = \"%s\"\n\n[variables]\n", name)
		if err = ioutil.WriteFile(fileName, []byte(template), 0644); err != nil {
			return exception("Create stage error: %s", err.Error())
		}
	}
	api := request.NewAPIGateway(c)
	if err = api.Deploy(c.RestApiId, name, ctx.String("message")); err != nil {
		return nil
	}

	return nil
}
