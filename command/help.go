package command

import (
	"github.com/ysugimoto/go-args"
)

type Help struct {
	Command
}

func NewHelp() *Help {
	return &Help{}
}

func (h *Help) Run(ctx *args.Context) {
	// noop
}

func (h *Help) Help() string {
	help := `
Usage:
  $ ginger [subcommand] [options]

SubCommands:
  init     : Initialize project
  install  : Install ginger dependencies
  function : Manage Go runtime Lambda functions
  api      : Manage APIGateway resources
  deploy   : Deploy function or api resource

Options:
  -h, --help: Show help

To see subcommand help, run "ginger [subcommand] help".`

	return COMMAND_HEADER + help
}
