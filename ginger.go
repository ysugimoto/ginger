package main

import (
	"fmt"
	"os"

	"github.com/ysugimoto/ginger/command"
	"github.com/ysugimoto/go-args"
)

func main() {
	ctx := args.New().
		Alias("help", "h", nil).
		Alias("name", "n", "").
		Alias("profile", "", "").
		Alias("region", "", "").
		Alias("role", "", "").
		Alias("path", "p", "").
		Alias("method", "m", "GET").
		Alias("body", "b", "").
		Alias("stage", "s", "").
		Alias("all", "", "").
		Alias("event", "e", "").
		Alias("memory", "m", 128).
		Alias("timeout", "t", 3).
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
	}

	if cmd == nil {
		cmd = command.NewHelp()
		fmt.Println(cmd.Help())
	} else if ctx.Has("help") {
		fmt.Println(cmd.Help())
	} else if err := cmd.Run(ctx); err != nil {
		fmt.Printf("Command %s failed: %s\n", ctx.At(0), err.Error())
	}
}
