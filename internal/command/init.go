package command

import (
	"os"

	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/entity"
	"github.com/ysugimoto/ginger/internal/logger"
)

// Init is the struct for initalize ginger project.
// This command generates config file and some
// directory structure.
type Init struct {
	Command
	log *logger.Logger
}

func NewInit() *Init {
	return &Init{
		log: logger.WithNamespace("ginger.init"),
	}
}

// Display init command help.
func (i *Init) Help() string {
	return "No Help"
}

// Run the init command.
func (i *Init) Run(ctx *args.Context) {
	c := config.Load()
	if c.Exists() {
		i.log.Warn("Config file found. Project has already initialized.")
		return
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

	if p := ctx.String("profile"); p != "" {
		project.Profile = p
		i.log.Printf("Profile set as %s\n", project.Profile)
	}
	if r := i.regionFromProfile(project.Profile); r != "" {
		project.Region = r
		i.log.Printf("Region set as %s\n", project.Region)
	}
	if r := ctx.String("region"); r != "" {
		project.Region = r
		i.log.Printf("Region set as %s\n", project.Region)
	}
	if r := ctx.String("role"); r != "" {
		project.LambdaExecutionRole = r
	}

	if project.LambdaExecutionRole == "" {
		i.log.Warn("Lambda Execution Role isn't set. Please run 'ginger config --role [role name]' before you deploy function.")
	} else {
		i.log.Printf("Lambda role set as %s\n", project.LambdaExecutionRole)
	}
	c.Project = project
	c.Write()
	NewInstall().Run(ctx)
	i.log.Info("ginger initalized successfully!")
}

// Try to get region from supplied profile.
func (i *Init) regionFromProfile(profile string) string {
	var sess *session.Session
	if profile != "" {
		creds := credentials.NewSharedCredentials("", profile)
		sess = session.New(aws.NewConfig().WithCredentials(creds))
	} else {
		sess = session.Must(session.NewSession())
	}
	return *sess.Config.Region
}
