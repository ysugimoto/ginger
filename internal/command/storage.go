package command

import (
	"fmt"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/logger"
	"github.com/ysugimoto/go-args"
)

const (
	STORAGE_DEPLOY = "deploy"
)

type Storage struct {
	Command
	log *logger.Logger
}

func NewStorage() *Storage {
	return &Storage{
		log: logger.WithNamespace("ginger.storage"),
	}
}

func (s *Storage) Help() string {
	return "No Help for now."
}

func (s *Storage) Run(ctx *args.Context) {
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
	case STORAGE_DEPLOY:
		err = NewDeploy().deployStorage(c, ctx)
	default:
		fmt.Println(s.Help())
	}
}
