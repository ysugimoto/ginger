package command

import (
	"fmt"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/entity"
	"github.com/ysugimoto/ginger/internal/input"
	"github.com/ysugimoto/ginger/internal/logger"
	"github.com/ysugimoto/ginger/internal/request"
	"github.com/ysugimoto/go-args"
)

const (
	STORAGE_DEPLOY = "deploy"
	STORAGE_LINK   = "link"
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
	return "No Help for now."
}

func (s *Storage) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		s.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
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
	case STORAGE_DEPLOY:
		err = NewDeploy().deployStorage(c, ctx)
	case STORAGE_LINK:
		err = s.linkStorage(c, ctx)
	default:
		fmt.Println(s.Help())
	}
}

func (s *Storage) linkStorage(c *config.Config, ctx *args.Context) error {
	bucket := ctx.String("bucket")
	if bucket == "" {
		bucket = c.Project.S3BucketName
	}
	if bucket != c.Project.S3BucketName {
		s.log.Warnf("Target bucket %s is external bucket. ginger only manages integration.", bucket)
	}
	path := ctx.String("path")
	if path == "" {
		return exception("Endpoint path is required. Run with -p, --path option.")
	} else if !c.API.Exists(path) {
		return exception("Endpoint %s does not defined in your project.\n", path)
	}

	rs := c.API.Find(path)
	if rs.Id != "" && rs.Integration != nil {
		// If interagraion already exists. need to delete it.
		inquiry := fmt.Sprintf("%s has already have integration type of %s. Override it?", rs.Path, rs.Integration.IntegrationType)
		if !input.Bool(inquiry) {
			return exception("Canceled.")
		}
		api := request.NewAPIGateway(c)
		api.DeleteMethod(c.API.RestId, rs.Id, rs.Integration.Method())
		api.DeleteIntegration(c.API.RestId, rs.Id, rs.Integration.Method())
		rs.Integration = nil
	}

	rs.Integration = &entity.Integration{
		IntegrationType: "s3",
		Bucket:          bucket,
	}
	s.log.Infof("Linked storage %s to resource %s.\n", bucket, path)
	return nil
}
