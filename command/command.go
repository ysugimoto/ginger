package command

import (
	"github.com/ysugimoto/go-args"
)

const (
	INIT     = "init"
	INSTALL  = "install"
	FUNCTION = "function"
	FN       = "fn"

	CREATE = "create"
	BUILD  = "build"
	DEPLOY = "deploy"
)

type Command interface {
	Run(ctx *args.Context) error
	Help() string
}
