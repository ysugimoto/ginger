package command

import (
	"fmt"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/go-args"
)

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
	NewInstall().Run(ctx)
	fmt.Println("[Init] ginger initalized successfully!")
	return nil
}
