package command

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/ysugimoto/go-args"
)

const COMMAND_HEADER = `============================================================
 ginger: Go runtime lambda function framework
============================================================`

const (
	INIT     = "init"
	INSTALL  = "install"
	CONFIG   = "config"
	FUNCTION = "function"
	FN       = "fn"
	API      = "api"
	CREATE   = "create"
	BUILD    = "build"
	DEPLOY   = "deploy"
	LOG      = "log"
)

// Command is the interface implemented by structs that can run the command
// and show help as usage.
type Command interface {
	Run(ctx *args.Context)
	Help() string
}

var environments map[string]string

// On initialize phase, collect the current environments to map inside
// to override and supplie some values on execute command
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

// buildEnv overrides environment variable supplied by argument map.
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
