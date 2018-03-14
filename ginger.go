package main

import (
	"fmt"
	"os"

	"github.com/ysugimoto/ginger/command"
	"github.com/ysugimoto/go-args"
)

var Version = "dev"

func main() {
	ctx := args.New().
		Alias("help", "h", nil).
		Alias("name", "n", "").
		Alias("profile", "", "").
		Alias("role", "", "").
		Alias("hook", "", "").
		Alias("path", "p", "").
		Alias("method", "m", "").
		Alias("memory", "", 128).
		Alias("timeout", "", 3).
		Alias("body", "b", "").
		Alias("stage", "s", "").
		Alias("event", "e", "").
		Alias("force", "", nil).
		Alias("filter", "", "").
		Alias("bucket", "", "").
		Alias("directory", "d", "").
		Alias("delete", "", nil).
		Alias("message", "", "").
		Parse(os.Args[1:])

	var cmd command.Command
	switch ctx.At(0) {
	case command.VERSION:
		cmd = command.NewVersion()
	case command.INIT:
		cmd = command.NewInit()
	case command.INSTALL:
		cmd = command.NewInstall()
	case command.CONFIG:
		cmd = command.NewConfig()
	case command.FUNCTION, command.FN:
		cmd = command.NewFunction()
	case command.RESOURCE, command.R:
		cmd = command.NewResource()
	case command.DEPLOY, command.D:
		cmd = command.NewDeploy()
	case command.STORAGE, command.S:
		cmd = command.NewStorage()
	case command.STAGE:
		cmd = command.NewStage()
	case command.SCHEDULER, command.SC:
		cmd = command.NewScheduler()
	default:
		cmd = command.NewHelp()
	}

	if ctx.Has("help") || ctx.At(1) == "help" {
		fmt.Println(cmd.Help())
	} else {
		cmd.Run(ctx)
	}
}
