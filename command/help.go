package command

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/ysugimoto/ginger/assets"
	"github.com/ysugimoto/go-args"
)

// Help is the struct that displays global command usage.
type Help struct {
	Command
}

func NewHelp() *Help {
	return &Help{}
}

func (h *Help) Run(ctx *args.Context) error {
	if ctx.At(0) == "ale" {
		file, _ := assets.Assets.Open("/ale")
		b := new(bytes.Buffer)
		io.Copy(b, file)
		d, _ := base64.StdEncoding.DecodeString(b.String())
		fmt.Println(string(d))
	} else {
		fmt.Println(h.Help())
	}
	return nil
}

func (h *Help) Help() string {
	help := `
Usage:
  $ ginger [subcommand] [options]

SubCommands:
  init      : Initialize project
  install   : Install ginger dependencies
  config    : Update project configurations
  function  : Manage Go runtime Lambda functions
  scheduler : Manage CloudWatchEvent scheduler
  resource  : Manage APIGateway resources
  stage     : Manage APIGateway stages
  deploy    : Deploy function or api resource

Options:
  -h, --help: Show help

To see subcommand help, run "ginger [subcommand] help".`

	return commandHeader() + help
}
