package command

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"os/exec"
	"path/filepath"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/go-args"
)

var dependencyPackages = []string{
	"github.com/aws/aws-lambda-go",
	"github.com/aws/aws-sdk-go",
}

type Init struct {
	Command
}

func NewInit() *Init {
	return &Init{}
}

func (i *Init) Run(ctx *args.Context) (err error) {
	config := NewConfig()
	if config.Exists {
		return fmt.Errorf("[Init] Configuration already exists!")
	}
	fn := entity.Function{Name: "example"}
	config.addFunc(fn)
	config.write()
	fmt.Println("[Init] Configuration file created.")
	fmt.Println("[Init] Install vendor libraries.")
	vendor := filepath.Join(config.Root, "vendor")
	var wg sync.WaitGroup
	for _, pkg := range dependencyPackages {
		wg.Add(1)
		go i.installDependencies(vendor, pkg, &wg)
	}
	wg.Wait()
	fmt.Println("[Init] ginger initalized successfully!")
	return nil
}

func (i *Init) installDependencies(vendor, pkg string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("[Init] Installing %s...\n", pkg)
	cmd := exec.Command("go", "get", pkg)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GOPATH=") {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", vendor))
		} else {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Run()
}
