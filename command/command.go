package command

import (
	"os"
	"strings"

	"github.com/ysugimoto/go-args"
)

var version string = "dev"

func commandHeader() string {
	return `============================================================
 ginger: Go runtime lambda function framework ` + version + `
============================================================`
}

const (
	VERSION   = "version"
	INIT      = "init"
	INSTALL   = "install"
	CONFIG    = "config"
	DEPLOY    = "deploy"
	D         = "d"
	STAGE     = "stage"
	ST        = "st" // alias for "stage"
	FUNCTION  = "function"
	FN        = "fn" // alias for "function"
	RESOURCE  = "resource"
	R         = "r" // alias for "resource"
	STORAGE   = "storage"
	S         = "s" // alias for "storage"
	SCHEDULER = "scheduler"
	SC        = "sc" // alias for schedule
)

// Command is the interface implemented by structs that can run the command
// and show help as usage.
type Command interface {
	Run(ctx *args.Context) error
	Help() string
}

var environments map[string]string

// On initialize phase, collect the current environments to map inside
// to override and supplies some values on execute command
func init() {
	environments = make(map[string]string)

	for _, e := range os.Environ() {
		spl := strings.SplitN(e, "=", 2)
		environments[spl[0]] = spl[1]
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
