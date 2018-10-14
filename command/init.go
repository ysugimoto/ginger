package command

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/internal/colors"
	"github.com/ysugimoto/ginger/logger"
)

const awsRegionInquiry = `[AWS region]
Region name which you want to deploy resources.
`

const lambdaRoleInquiry = `[Need lambda execution role]
If you will use Lambda function, Need to set "LambdaExecutionRole".
This is because lambda requires permission to run function itself or external service.
`

const s3BucketInquiry = `[S3 bucket name]
If you will use S3 storage for static file serve, We recommend to set 'S3BucketName'.
`

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
//
// >>> doc
//
// ## Initialize project
//
// Initialize ginger project at current directory.
//
// ```
// $ ginger init [options]
// ```
//
// If you want to use (probably almost case yes) external Go package, we suggest you should put project directory under the `$GOPATH` to enable to detect vendor tree.
//
// For example:
//
// ```
// cd $GOATH/src/github.com/your/project
// ginger init
// ```
//
// The ginger init command will work as following:
//
// - Create `Ginger.toml` file which is used for project configuration
// - Create `functions` directory which is used for function management
// - Create `stages` directory which is used for stage variable management
// - Create `.ginger`  directory which is used for put dependency packags. Those packages will be loaded on deploy phase..
// - Install dependency packages.
//
// Note that the `Ginger.toml` is readable and configurable, but almost values are added or updated via subcommands.
// So we don't recommend you change this file manually.
//
// And, when initializing project, ginger asks some questions.
//
// #### LambdaExecutionRole
//
// When ginger deploys function to AWS Lambda, execution role is necessary.
// So you should input lambda exection role to use as default. You can create role on AWS IAM.
// See: https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example-create-iam-role.html
//
// Or, you can use specific role by each function by adding `Function.toml`.
//
// #### S3BucketName
//
// ginger uses S3 bucket name project directory name as defaut. You can change this name.

// #### Region
//
// Destination AWS region which ginger create resources.
//
// <<< doc
func (i *Init) Run(ctx *args.Context) {
	c := config.Load()
	if c.Exists() {
		i.log.Warn("Config file found. Project has already initialized.")
		return
	}
	if _, err := os.Stat(c.FunctionPath); err != nil {
		i.log.Printf("Create functions directory: %s\n", c.FunctionPath)
		os.Mkdir(c.FunctionPath, 0755)
		i.ensureKeepFile(c.FunctionPath)
	}
	if _, err := os.Stat(c.StoragePath); err != nil {
		i.log.Printf("Create storage directory: %s\n", c.StoragePath)
		os.Mkdir(c.StoragePath, 0755)
		i.ensureKeepFile(c.StoragePath)
	}
	if _, err := os.Stat(c.StagePath); err != nil {
		i.log.Printf("Create stages directory: %s\n", c.StagePath)
		os.Mkdir(c.StagePath, 0755)
		i.ensureKeepFile(c.StagePath)
	}
	if _, err := os.Stat(c.SchedulerPath); err != nil {
		i.log.Printf("Create scheduler directory: %s\n", c.SchedulerPath)
		os.Mkdir(c.SchedulerPath, 0755)
		i.ensureKeepFile(c.SchedulerPath)
	}

	c.ProjectName = filepath.Base(c.Root)
	c.Profile = ""
	c.DefaultLambdaRole = ""
	c.Region = "us-east-1"

	if p := ctx.String("profile"); p != "" {
		c.Profile = p
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
			i.log.Print("Lambda Execution Role isn't set. If you want to set, run 'ginger config --role [role name]'\n")
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
	fmt.Println(colors.Yellow(awsRegionInquiry))
	if region := input.String(fmt.Sprintf("Input region name (default: %s)", c.Region)); region != "" {
		c.Region = region
	}
	c.Write()
	NewInstall().Run(ctx)
	i.log.Info("ginger initalized successfully!")
}

// Ensure .keep file and create if not exist
// The .keep file is needed for adding directory to git
func (i *Init) ensureKeepFile(dir string) {
	keepFile := filepath.Join(dir, ".keep")
	if _, err := os.Stat(keepFile); err == nil {
		return
	}
	// Create empty file
	fp, _ := os.OpenFile(keepFile, os.O_CREATE|os.O_RDONLY, 0666)
	fp.Close()
}
