package command

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/internal/colors"
	"github.com/ysugimoto/ginger/logger"
)

const lambdaRoleInquiry = `[Need lambda execution role]
If you will use Lambda function, Need to set "LambdaExecutionRole".
This is because lambda requires permission to run function itself or external service.`

const s3BucketInquiry = `[S3 bucket name]
If you will use S3 storage for static file serve, We recommend to set 'S3BucketName'.`

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
	if _, err := os.Stat(c.StoragePath); err != nil {
		i.log.Printf("Create storage directory: %s\n", c.StoragePath)
		os.Mkdir(c.StoragePath, 0755)
	}
	if _, err := os.Stat(c.StagePath); err != nil {
		i.log.Printf("Create stages directory: %s\n", c.StagePath)
		os.Mkdir(c.StagePath, 0755)
	}

	c.ProjectName = filepath.Base(c.Root)
	c.Profile = ""
	c.DefaultLambdaRole = ""
	c.Region = "us-east-1"

	if p := ctx.String("profile"); p != "" {
		c.Profile = p
		i.log.Printf("Profile set as %s\n", c.Profile)
	}
	if r := i.regionFromProfile(c.Profile); r != "" {
		c.Region = r
		i.log.Printf("Region set as %s\n", c.Region)
	}
	if r := ctx.String("region"); r != "" {
		c.Region = r
		i.log.Printf("Region set as %s\n", c.Region)
	}
	if r := ctx.String("role"); r != "" {
		c.DefaultLambdaRole = r
	}
	if b := ctx.String("bucket"); b != "" {
		c.S3BucketName = b
	}

	if c.DefaultLambdaRole == "" {
		fmt.Println(colors.Yellow(lambdaRoleInquiry))
		if role := input.String("Input lambda execution role ARN (empty to skip)"); role == "" {
			i.log.Warn("Lambda Execution Role isn't set. If you want to set, run 'ginger config --role [role name]'")
		} else {
			c.DefaultLambdaRole = role
		}
	}
	if c.S3BucketName == "" {
		fmt.Println(colors.Yellow(s3BucketInquiry))
		if bucketName := input.String(fmt.Sprintf("Input bucket name (default: ginger-%s, empty to skip)", c.ProjectName)); bucketName == "" {
			c.S3BucketName = fmt.Sprintf("ginger-%s", c.ProjectName)
		} else {
			c.S3BucketName = bucketName
		}
	}
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
