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

func (h *Help) Run(ctx *args.Context) error {
	return nil
}

func (h *Help) Help() string {
	help := `=================================================
 ginger: Go runtime lambda function framework
=================================================
Usage:
  ginger [subcommand] [options]

SubCommands:
  init: Initialize project
  create: Create lambda function boilerplate

Options:
  -h, --help: Show help`

	return help
}
