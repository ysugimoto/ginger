package command

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/ysugimoto/go-args"
)

const (
	INIT     = "init"
	INSTALL  = "install"
	FUNCTION = "function"
	FN       = "fn"

	API = "api"

	CREATE = "create"
	BUILD  = "build"
	DEPLOY = "deploy"
)

type Command interface {
	Run(ctx *args.Context) error
	Help() string
}

var environments map[string]string

func init() {
	cwd, _ := os.Getwd()
	vendorPath := filepath.Join(cwd, "vendor")
	environments = make(map[string]string)

	for _, e := range os.Environ() {
		spl := strings.SplitN(e, "=", 2)
		environments[spl[0]] = spl[1]
	}

	if v, ok := environments["GOPATH"]; ok {
		environments["GOPATH"] = fmt.Sprintf("%s:%s", v, vendorPath)
	} else {
		environments["GOPATH"] = vendorPath
	}
}

func buildEnv(overrides map[string]string) []string {
	if overrides == nil {
		overrides = make(map[string]string)
	}
	envs := []string{}
	for k, v := range environments {
		if ov, ok := overrides[k]; ok {
			envs = append(envs, k+"="+ov)
			delete(overrides, k)
		} else {
			envs = append(envs, k+"="+v)
		}
	}
	for k, v := range overrides {
		envs = append(envs, k+"="+v)
	}
	return envs
}
