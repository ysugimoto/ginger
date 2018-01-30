package command

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/go-args"
)

type Create struct {
	Command
	conf *Config
}

func NewCreate() *Create {
	return &Create{
		conf: NewConfig(),
	}
}

func (c *Create) Run(ctx *args.Context) (err error) {
	if !c.conf.Exists {
		fmt.Println("[Create] Configuration file isn't exist. Run `ginger init` befgore.")
		os.Exit(1)
	}
	name := ctx.At(1)
	if name == "" {
		return fmt.Errorf("[Create]: function name must be supplied.")
	}
	fnPath := filepath.Join(c.conf.Root, name)
	if s, err := os.Stat(fnPath); err == nil {
		if s.IsDir() {
			return fmt.Errorf("[Function] directory already exists at %s.", fnPath)
		}
		return fmt.Errorf("[Function] file exists at %s.", fnPath)
	}
	fn := entity.Function{
		Name:       name,
		APIGateway: nil,
	}
	fmt.Printf("Creating new function [%s]...\n", name)
	if input.Bool("Do you want to use AWS API Gateway?") {
		method := input.Choice("Which method do you want to handle?", []string{"GET", "POST", "PUT", "DELETE"})
		path := input.String("What endpoint do you want to handle?")
		fn.APIGateway = &entity.APIGateway{
			Path:   path,
			Method: method,
		}
	}
	if err := fn.Save(c.conf.Root); err != nil {
		return err
	}
	c.conf.addFunc(fn)
	c.conf.write()
	fmt.Println("Function created successfully!")
	return nil
}
