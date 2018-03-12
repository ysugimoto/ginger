package command

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"github.com/mattn/go-tty"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
)

const (
	STAGECREATE = "create"
	STAGEDELETE = "delete"
	STAGEDEPLOY = "deploy"
	STAGELIST   = "list"
	STAGEHELP   = "help"
)

// Stage is the struct of AWS API Gateway stage operation command.
// This struct will be dispatched on "ginger stage" subcommand.
type Stage struct {
	Command
	log *logger.Logger
}

func NewStage() *Stage {
	return &Stage{
		log: logger.WithNamespace("ginger.stage"),
	}
}

// Show function command help.
func (s *Stage) Help() string {
	return commandHeader() + `
stage - (AWS API Gateway) stage management command.

Usage:
  $ ginger stage [operation] [options]

Operation:
  create : Create new stage
  delete : Delete stage
  deploy : Deploy stage
  list   : List stages
  help   : Show this help

Options:
  -n, --name : [all] Stage name (required)
`
}

// Run the command.
func (s *Stage) Run(ctx *args.Context) {
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
	case STAGECREATE:
		err = s.createStage(c, ctx)
	case STAGEDELETE:
		err = s.deleteStage(c, ctx)
	case STAGEDEPLOY:
		// err = NewDeploy().deployStage(c, ctx)
	case STAGELIST:
		err = s.listStage(c, ctx)
	default:
		fmt.Println(s.Help())
	}
}

// createStage creates new stage.
//
// >>> doc
//
// ## Create new stage
//
// Create new API Gateway stage.
//
// ```
// $ ginger stage create [options]
// ```
//
// | option  | description           |
// |:-------:|:----------------------|
// | --name  | [Required] Stage name |
//
// If REST API has already been created, also create stage to API Gateway.
//
// <<< doc
func (s *Stage) createStage(c *config.Config, ctx *args.Context) (err error) {
	name := ctx.String("name")
	if name == "" {
		return exception("Stage name didn't supplied. Run with --name option.")
	} else if _, err := c.LoadStage(name); err == nil {
		return exception("Stage already defined.")
	}

	fileName := filepath.Join(c.StagePath, fmt.Sprintf("%s.toml", name))
	template := fmt.Sprintf("name = \"%s\"\n\n[variables]\n", name)
	if err = ioutil.WriteFile(fileName, []byte(template), 0644); err != nil {
		return exception("Create stage json error: %s", err.Error())
	}

	s.log.Infof("Stage created. To manage stage variables, edit stages/%s.toml.\n", name)
	return nil
}

// deleteStage deletes stage.
//
// >>> doc
//
// ## Delete stage
//
// Delete stage.
//
// ```
// $ ginger stage delete [options]
// ```
//
// | option  | description                                    |
// |:-------:|:-----------------------------------------------|
// | --name  | [Required] Stage name which you want to delete |
//
// If stage has been deployed on AWS API Gateway, also delete it.
// It means no longer access to that stage API.
//
// <<< doc
func (s *Stage) deleteStage(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Stage name didn't supplied. Run with --name option.")
	} else if _, err := c.LoadStage(name); err != nil {
		return exception("Stage not defined.")
	}

	if c.RestApiId != "" {
		if !ctx.Has("force") && !input.Bool("Also deletes all deployments for this stage. Are you sure?") {
			s.log.Warn("Abort.")
			return nil
		}

		api := request.NewAPIGateway(c)
		s.log.Print("Checking stage exintence...")
		if api.StageExists(c.RestApiId, name) {
			if err := api.DeleteStage(c.RestApiId, name); err != nil {
				s.log.Error("Failed to delete from AWS. Please delete manually.")
			}
		} else {
			s.log.Print("Not found in AWS API Gateway. Skip it.")
		}
	}

	s.log.Print("Deleting files...")
	if err := os.Remove(filepath.Join(c.StagePath, fmt.Sprintf("%s.toml", name))); err != nil {
		return exception("Delete file error: %s", err.Error())
	}
	s.log.Info("Stage deleted successfully.")
	return nil
}

// listStage shows registered stages.
//
// >>> doc
//
// ## List stages`
//
// Display list of created stages.
//
// ```
// $ ginger stage list
// ```
//
// <<< doc
func (s *Stage) listStage(c *config.Config, ctx *args.Context) error {
	api := request.NewAPIGateway(c)
	stages := api.GetStages(c.RestApiId)
	localStages, _ := c.LoadAllStages()
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
	fmt.Printf("%-24s %-24s %-12s %-12s %-12s\n", "StageName", "DeploymentId", "Deployed", "Created", "LastUpdated")
	fmt.Println(line)
	for i, stg := range localStages {
		c := "-"
		u := "-"
		j := "-"
		d := "No"
		for _, rs := range stages {
			if *rs.StageName == stg.Name {
				created := *rs.CreatedDate
				updated := *rs.LastUpdatedDate
				c = created.Format("2016-01-02 15:04:05")
				u = updated.Format("2016-01-02 15:04:05")
				j = *rs.DeploymentId
				d = "Yes"
				break
			}
		}
		fmt.Printf("%-24s %-24s %-12s %-12s %-12s\n", stg.Name, j, d, c, u)
		if i != len(localStages)-1 {
			fmt.Println(strings.Repeat("-", w))
		}
	}
	return nil
}
