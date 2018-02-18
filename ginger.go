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
		Alias("region", "", "").
		Alias("role", "", "").
		Alias("hook", "", "").
		Alias("path", "p", "").
		Alias("method", "m", "").
		Alias("body", "b", "").
		Alias("stage", "s", "").
		Alias("event", "e", "").
		Alias("memory", "m", 128).
		Alias("timeout", "t", 3).
		Alias("force", "", nil).
		Alias("filter", "", "").
		Alias("bucket", "", "").
		Alias("directory", "d", "").
		Alias("delete", "", nil).
		Parse(os.Args[1:])

	var cmd command.Command
	switch ctx.At(0) {
	case command.INIT:
		cmd = command.NewInit()
	case command.INSTALL:
		cmd = command.NewInstall()
	case command.CONFIG:
		cmd = command.NewConfig()
	case command.FUNCTION, command.FN:
		cmd = command.NewFunction()
	case command.API:
		cmd = command.NewAPI()
	case command.DEPLOY:
		cmd = command.NewDeploy()
	case command.LOG:
		cmd = command.NewLog()
	case command.STORAGE, command.S:
		cmd = command.NewStorage()
	case command.INTEGRATE, command.I:
		cmd = command.NewIntegrate()
	case command.STAGE:
		cmd = command.NewStage()
	default:
		cmd = command.NewHelp()
	}

	if ctx.Has("help") || ctx.At(1) == "help" {
		fmt.Println(cmd.Help())
	} else {
		cmd.Run(ctx)
	}
}
