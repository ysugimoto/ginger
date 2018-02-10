package command

import (
	"fmt"

	"github.com/pkg/errors"
)

var debug string = ""

func debugTrace(err error) {
	if debug != "enable" {
		return
	}
	fmt.Println("StackTrace =====================")
	fmt.Printf("%+v\n", err)
}

func exception(message string, binds ...interface{}) error {
	return errors.New(fmt.Sprintf(message, binds...))
}
