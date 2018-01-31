package command

import (
	"fmt"
	"os"
	"sync"

	"os/exec"
	"path/filepath"

	"github.com/ysugimoto/go-args"
)

var dependencyPackages = []string{
	"github.com/aws/aws-lambda-go",
	"github.com/aws/aws-sdk-go",
}

type Install struct {
	Command
	conf *Config
}

func NewInstall() *Install {
	return &Install{
		conf: NewConfig(),
	}
}

func (i *Install) Run(ctx *args.Context) (err error) {
	fmt.Println("[Install] Install vendor libraries.")
	vendor := filepath.Join(i.conf.Root, "vendor")
	if _, err := os.Stat(vendor); err != nil {
		os.Mkdir(vendor, 0755)
	}
	var wg sync.WaitGroup
	for _, pkg := range dependencyPackages {
		wg.Add(1)
		fmt.Printf("[Init] Installing %s to %s...\n", pkg, vendor)
		go i.installDependencies(pkg, &wg)
	}
	wg.Wait()
	return nil
}

func (i *Install) installDependencies(pkg string, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("go", "get", pkg)
	cmd.Env = commandEnvironment()
	cmd.Run()
}
