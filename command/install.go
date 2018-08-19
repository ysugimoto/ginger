package command

import (
	"os"
	"strings"
	"sync"

	"go/parser"
	"go/token"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/scylladb/go-set/strset"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/logger"
)

// findDependencyPackages finds import packages in your lambda functions.
// walk functions directory recursively, and add to set.
func findDependencyPackages(root string) ([]*strset.Set, error) {
	files, err := listFunctionScriptFiles(root)
	if err != nil {
		return nil, err
	}

	// aws-lamda-go is required of default
	packages := []*strset.Set{
		strset.New("github.com/aws/aws-lambda-go"),
		strset.New(),
		strset.New(),
		strset.New(),
	}

	// Parse impopr section and add to set
	for _, f := range files {
		t := token.NewFileSet()
		ast, err := parser.ParseFile(t, f, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, i := range ast.Imports {
			pkg := strings.Trim(i.Path.Value, `"`)
			if !strings.Contains(pkg, ".") {
				continue
			}
			level := len(strings.Split(pkg, "/")) - 3
			if level < 0 {
				continue
			}
			packages[level].Add(pkg)
		}
	}

	return packages, nil
}

// walk function directoryy recursively and list files which has ".go" extension.
func listFunctionScriptFiles(root string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Install is the struct for install ginger project dependencies.
// This command installs project dependencies.
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

	if _, err := os.Stat(c.LibPath); err != nil {
		i.log.Printf("Create library directory: %s\n", c.LibPath)
		os.Mkdir(c.LibPath, 0755)
	}

	tmpDir, _ := ioutil.TempDir("", "ginger-tmp-packages")
	defer os.RemoveAll(tmpDir)

	deps, err := findDependencyPackages(c.FunctionPath)
	if err != nil {
		i.log.Errorf("Find dependency error: %s\n", err.Error())
		return
	}

	for _, pkgs := range deps {
		var wg sync.WaitGroup
		pkgs.Each(func(item string) bool {
			if _, err := os.Stat(filepath.Join(c.LibPath, "src", item)); err == nil {
				return true
			}
			wg.Add(1)
			i.log.Printf("Installing/Resolving %s...\n", item)
			go i.installDependencies(item, tmpDir, &wg)
			return true
		})
		wg.Wait()
	}
	// Recursive copy
	if err := i.movePackages(tmpDir, c.LibPath); err != nil {
		i.log.Error(err.Error())
	}

	i.log.Info("Installed dependencies successfully.")
}

// installDependencies installs dependencies via "go get".
//
// >>> doc
//
// ## Install dependencies
//
// Install dependency packages for build lambda function.
//
// ```
// $ ginger install
// ```
//
// This command is run automatically on initialize, but if you checkout project after initialize,
// You can install dependency packages via this command.
//
// <<< doc
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
		return exception("Failed to read directory: %s", src)
	}
	for _, item := range items {
		from := filepath.Join(src, item.Name())
		to := filepath.Join(dest, item.Name())
		if err := os.Rename(from, to); err != nil {
			return exception("Failed to move file: %s => %s, %s", from, to, err.Error())
		}
	}
	return nil
}
