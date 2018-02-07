package command

import (
	"os"

	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/go-args"
)

type Init struct {
	Command
	log *logger.Logger
}

func NewInit() *Init {
	return &Init{
		log: logger.WithNamespace("ginger.init"),
	}
}

func (i *Init) Run(ctx *args.Context) (err error) {
	c := config.Load()
	if c.Exists() {
		i.log.Warn("Config file found. Project has already initialized.")
		return nil
	}
	if _, err := os.Stat(c.FunctionPath); err != nil {
		i.log.Printf("Create functions directory: %s\n", c.FunctionPath)
		os.Mkdir(c.FunctionPath, 0755)
	}
	if _, err := os.Stat(c.VendorPath); err != nil {
		i.log.Printf("Create vendor directory: %s\n", c.VendorPath)
		os.Mkdir(c.VendorPath, 0755)
	}
	project := entity.Project{
		Name:                filepath.Base(c.Root),
		Profile:             "",
		Region:              "us-east-1",
		LambdaExecutionRole: "",
	}
	var sess *session.Session
	if p := ctx.String("profile"); p != "" {
		project.Profile = p
		creds := credentials.NewSharedCredentials("", p)
		sess = session.New(aws.NewConfig().WithCredentials(creds))
		if *sess.Config.Region != "" {
			project.Region = *sess.Config.Region
		}
	} else {
		sess = session.Must(session.NewSession())
		if *sess.Config.Region != "" {
			project.Region = *sess.Config.Region
		}
	}
	if r := ctx.String("region"); r != "" {
		project.Region = r
	}
	if r := ctx.String("role"); r != "" {
		project.LambdaExecutionRole = r
	} else {
	}

	i.log.Printf("Region set as %s\n", project.Region)
	i.log.Printf("Profile set as %s\n", project.Profile)
	if project.LambdaExecutionRole == "" {
		i.log.Warn("Lambda Execution Role isn't set. Please add 'lambda_execution_role' field before deploy function.")
	} else {
		i.log.Printf("Lambda role set as %s\n", project.LambdaExecutionRole)
	}
	c.Project = project
	c.Write()
	NewInstall().Run(ctx)
	i.log.Info("ginger initalized successfully!")
	return nil
}
