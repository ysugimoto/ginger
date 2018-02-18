package command

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"os/signal"

	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
)

type Log struct {
	Command
	log *logger.Logger
}

func NewLog() *Log {
	return &Log{
		log: logger.WithNamespace("ginger.log"),
	}
}

func (l *Log) Help() string {
	return COMMAND_HEADER + `
log - Tail cloudwatch logs for lambda function.

Usage:
  $ ginger log [options]

Options:
  -n, --name : Function name (required)
`
}

func (l *Log) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		l.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	var err error
	defer func() {
		if err != nil {
			l.log.Error(err)
			debugTrace(err)
		}
	}()
	err = l.tailLogs(c, ctx)
}

func (l *Log) tailLogs(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Function name didn't supplied. Run with --name option.")
	} else if !c.Functions.Exists(name) {
		return exception("Function not defined.")
	}

	l.log.Warnf("Tailing cloudwatch logs for function \"%s\"...\n", name)
	ctc, cancel := context.WithCancel(context.Background())
	cwl := request.NewCloudWatchLogsRequest(c)
	go cwl.TailLogs(
		ctc,
		fmt.Sprintf("/aws/lambda/%s", name),
		ctx.String("filter"),
	)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-ch:
		cancel()
	}
	return nil
}
