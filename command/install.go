package command

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"sync"

	"go/parser"
	"go/token"
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

	// aws-lamda-go is required as default
	// Usually 3 depth is enough
	packages := []*strset.Set{
		strset.New("github.com/aws/aws-lambda-go/lambda"),
		strset.New(),
		strset.New(),
	}

	// Parse import section and add to set
	for _, f := range files {
		t := token.NewFileSet()
		ast, err := parser.ParseFile(t, f, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, i := range ast.Imports {
			// Import path value also contains double quotes, so we trim them
			pkg := strings.Trim(i.Path.Value, `"`)
			if !strings.Contains(pkg, ".") {
				continue
			}
			// Get import path depth
			level := len(strings.Split(pkg, "/")) - 1
			// Expand slice if depth is not enough
			for len(packages)-1 < level {
				packages = append(packages, strset.New())
			}
			packages[level].Add(pkg)
		}
	}

	return packages, nil
}

// walk function directory recursively and list files which have ".go" extension.
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

func (i *Install) Run(ctx *args.Context) error {
	c := config.Load()
	if !c.Exists() {
		i.log.Error("Configuration file could not load. Run `ginger init` before.")
		return errors.New("")
	}

	i.log.Print("Install function dependencies.")

	if _, err := os.Stat(c.LibPath); err != nil {
		i.log.Printf("Create library directory: %s\n", c.LibPath)
		if err := os.Mkdir(c.LibPath, 0755); err != nil {
			i.log.Error("Failed to create directory: " + c.LibPath)
			return err
		}
	}

	deps, err := findDependencyPackages(c.FunctionPath)
	if err != nil {
		i.log.Errorf("Find dependency error: %s\n", err.Error())
		return err
	}

	errList := make([]string, 0)
	for _, pkgs := range deps {
		var wg sync.WaitGroup
		errChan := make(chan error, pkgs.Size())
		pkgs.Each(func(item string) bool {
			wg.Add(1)
			i.log.Printf("Installing/Resolving %s...\n", item)
			go func() {
				defer wg.Done()
				if err := i.installDependencies(item, c.LibPath); err != nil {
					errChan <- err
				}
			}()
			return true
		})
		go func() {
			wg.Wait()
			close(errChan)
		}()
		for err := range errChan {
			errList = append(errList, err.Error())
		}
	}

	if len(errList) > 0 {
		i.log.Errorf("Failed to install some packages:\n%s\n", strings.Join(errList, "\n"))
		return errors.New("")
	}
	i.log.Info("Dependencies installed successfully.")
	return nil
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
// ginger detects imports from your *.go file and install inside `.ginger` directory.
//
// <<< doc
func (i *Install) installDependencies(pkg, tmpDir string) error {
	buffer := new(bytes.Buffer)
	cmd := exec.Command("go", "get", pkg)
	if tmpDir != "" {
		cmd.Env = buildEnv(map[string]string{
			"GOPATH": tmpDir,
		})
	}
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	if err := cmd.Run(); err != nil {
		return errors.New(string(buffer.Bytes()))
	}
	return nil
}
