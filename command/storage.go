package command

import (
	"errors"
	"fmt"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
	"github.com/ysugimoto/go-args"
)

const (
	STORAGEDEPLOY  = "deploy"
	STORAGEMOUNT   = "mount"
	STORAGEUNMOUNT = "unmount"
)

type Storage struct {
	Command
	log *logger.Logger
}

func NewStorage() *Storage {
	return &Storage{
		log: logger.WithNamespace("ginger.storage"),
	}
}

func (s *Storage) Help() string {
	return commandHeader() + `
storage - (AWS S3) management command.

Usage:
  $ ginger storage [operation] [options]

Operation:
  mount   : Mount function to destination path
  unmount : Unmount function from destination path
  deploy  : Deploy functions
  help    : Show this help

Options:
  -p, --path      : [mount] Path name
  -b, --bucket    : [mount] Bucket name (default use configuation)
  -d, --directory : [mount] Mount target directory on S3
`
}

func (s *Storage) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		s.log.Error("Configuration file could not load. Run `ginger init` before.")
		return errors.New("")
	}
	var err error
	defer func() {
		if err != nil {
			s.log.Error(err.Error())
			debugTrace(err)
		}
		c.Write()
	}()

	switch ctx.At(1) {
	case STORAGEDEPLOY:
		err = NewDeploy().deployStorage(c, ctx)
	case STORAGEMOUNT:
		err = s.mountStorage(c, ctx)
	case STORAGEUNMOUNT:
		err = s.unmountStorage(c, ctx)
	default:
		fmt.Println(s.Help())
	}
	return err
}

// mountStorage makes integration between API Gateway resource and S3 storage
func (s *Storage) mountStorage(c *config.Config, ctx *args.Context) error {
	bucket := ctx.String("bucket")
	if bucket == "" {
		bucket = c.S3BucketName
	}
	if bucket != c.S3BucketName {
		s.log.Warnf("Target bucket %s is external bucket. ginger only manages integration.\n", bucket)
	}
	path := ctx.String("path")
	if path == "" {
		path = c.ChooseResource()
	}
	if _, err := c.LoadResource(path); err != nil {
		return exception("Resource %s could not find in your project.\n", path)
	}
	rs, err := c.LoadResource(path + "/{proxy+}")
	if err != nil {
		s.log.Warnf("Create %s/{proxy+} sub-resource in order to map sub-path request.\n", path)
		rs = entity.NewResource("", path+"/{proxy+}")
		c.Resources = append(c.Resources, rs)
	}

	dir := ctx.String("directory")
	if dir == "" {
		dir = "/"
	}

	// On storage integration, we only support "GET" method.
	method := "GET"
	ig := rs.GetIntegration(method)
	if ig == nil {
		ig = entity.NewIntegration("s3", bucket+dir, rs.Path)
		rs.AddIntegration(method, ig)
		s.log.Infof("mounted storage %s to resource %s.\n", bucket+dir, path)
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

// mountStorage makes integration between API Gateway resource and S3 storage
func (s *Storage) unmountStorage(c *config.Config, ctx *args.Context) error {
	path := ctx.String("path")
	if path == "" {
		path = c.ChooseResource()
	}
	if _, err := c.LoadResource(path); err != nil {
		return exception("Resource %s could not find in your project.\n", path)
	}
	rs, err := c.LoadResource(path + "/{proxy+}")
	if err != nil {
		return exception("Proxy resource could not find in %s.\n", path)
	}

	// On storage integration, we only support "GET" method.
	method := "GET"
	ig := rs.GetIntegration(method)
	if ig == nil {
		return exception("No integration found, Abort.")
	}
	switch ig.IntegrationType {
	case "lambda":
		return exception("Lambda function already integrated. to remove it, run 'ginger function unmount'.")
	case "s3":
		if rs.Id != "" {
			api := request.NewAPIGateway(c)
			api.DeleteMethod(c.RestApiId, rs.Id, method)
			api.DeleteIntegration(c.RestApiId, rs.Id, method)
		}
		rs.DeleteIntegration(method)
	default:
		return exception("Undefined integration type.")
	}
	s.log.Infof("Storage unmounted for resource %s.\n", path)
	return nil
}
