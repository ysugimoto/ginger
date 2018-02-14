package command

import (
	"encoding/base64"
	"fmt"

	"github.com/ysugimoto/ginger/internal/assets"
	"github.com/ysugimoto/go-args"
)

// Help is the struct that displays global command usage.
type Help struct {
	Command
}

func NewHelp() *Help {
	return &Help{}
}

func (h *Help) Run(ctx *args.Context) {
	if ctx.At(0) == "ale" {
		b, _ := assets.Asset("ale")
		d, _ := base64.StdEncoding.DecodeString(string(b))
		fmt.Println(string(d))
	} else {
		fmt.Println(h.Help())
	}
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
  stage    : Manage APIGateway stages
  deploy   : Deploy function or api resource
  log      : Tail function log

Options:
  -h, --help: Show help

To see subcommand help, run "ginger [subcommand] help".`

	return COMMAND_HEADER + help
}
