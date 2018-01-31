package command

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/ysugimoto/go-args"
)

const (
	INIT    = "init"
	INSTALL = "install"
	CONFIG  = "config"
	CREATE  = "create"
	BUILD   = "build"
	DEPLOY  = "deploy"
)

type Command interface {
	Run(ctx *args.Context) error
	Help() string
}

func commandEnvironment() (envs []string) {
	cwd, _ := os.Getwd()
	vendorPath := filepath.Join(cwd, "vendor")

	var found bool
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") {
			found = true
			envs = append(envs, fmt.Sprintf("%s:%s", e, vendorPath))
		} else {
			envs = append(envs, e)
		}
	}
	if !found {
		envs = append(envs, fmt.Sprintf("GOPATH=%s", vendorPath))
	}
	return envs
}
