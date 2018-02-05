package command

import (
	"sync"

	"os/exec"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/go-args"
)

var dependencyPackages = []string{
	"github.com/aws/aws-lambda-go",
	"github.com/aws/aws-sdk-go",
}

type Install struct {
	Command
	log *logger.Logger
}

func NewInstall() *Install {
	return &Install{
		log: logger.WithNamespace("ginger.install"),
	}
}

func (i *Install) Run(ctx *args.Context) (err error) {
	c := config.Load()
	if !c.Exists() {
		i.log.Error("Configuration file could not load. Run `ginger init` before.")
		return nil
	}
	var wg sync.WaitGroup
	for _, pkg := range dependencyPackages {
		wg.Add(1)
		go i.installDependencies(pkg, &wg)
	}
	wg.Wait()
	return nil
}

func (i *Install) installDependencies(pkg string, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("go", "get", pkg)
	cmd.Env = buildEnv
	cmd.Run()
}
