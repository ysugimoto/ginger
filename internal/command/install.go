package command

import (
	"os"
	"sync"

	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/logger"
)

var dependencyPackages = []string{
	"github.com/aws/aws-lambda-go",
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

func (i *Install) Help() string {
	return "No Help"
}

func (i *Install) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		i.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}

	i.log.Print("Install function dependencies.")

	tmpDir, _ := ioutil.TempDir("", "ginger-tmp-packages")
	defer os.RemoveAll(tmpDir)

	var wg sync.WaitGroup
	for _, pkg := range dependencyPackages {
		wg.Add(1)
		i.log.Printf("Installing %s...\n", pkg)
		go i.installDependencies(pkg, tmpDir, &wg)
	}
	wg.Wait()

	// Recursive copy
	if err := i.movePackages(filepath.Join(tmpDir, "src"), c.VendorPath); err != nil {
		i.log.Error(err.Error)
	}
}

func (i *Install) installDependencies(pkg, tmpDir string, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("go", "get", pkg)
	cmd.Env = buildEnv(map[string]string{
		"GOPATH": tmpDir,
	})
	cmd.Run()
}

func (i *Install) movePackages(src, dest string) error {
	items, err := ioutil.ReadDir(src)
	if err != nil {
		return exception("Failed to read directort: %s", src)
	}
	for _, item := range items {
		from := filepath.Join(src, item.Name())
		to := filepath.Join(dest, item.Name())
		if err := os.Rename(from, to); err != nil {
			return exception("Failed to move file: %s => %s", from, to)
		}
	}
	return nil
}
