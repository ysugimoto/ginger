package command

import (
	"fmt"
	"os"

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
		i.log.Info("Config file found. Project has already initialized.")
		return nil
	}
	if _, err := os.Stat(c.FunctionPath); err != nil {
		os.Mkdir(c.FunctionPath, 0755)
	}
	if _, err := os.Stat(c.APIPath); err != nil {
		os.Mkdir(c.APIPath, 0755)
	}
	if _, err := os.Stat(c.VendorPath); err != nil {
		os.Mkdir(c.VendorPath, 0755)
	}
	var profile string
	var region string
	var sess *session.Session
	if p := ctx.String("profile"); p != "" {
		profile = p
		creds := credentials.NewSharedCredentials("", p)
		sess = session.New(aws.NewConfig().WithCredentials(creds))
		region = *sess.Config.Region
	} else {
		profile = "default"
		sess = session.Must(session.NewSession())
		region = *sess.Config.Region
	}
	if region == "" {
		region = "ua-east-1"
	}
	c.Project = entity.Project{
		Profile: profile,
		Region:  region,
	}
	c.Write()
	NewInstall().Run(ctx)
	fmt.Println("[Init] ginger initalized successfully!")
	return nil
}
