package command

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func debugTrace(err error) {
	if d := os.Getenv("DEBUG"); d == "" {
		return
	}
	fmt.Println("StackTrace =====================")
	fmt.Printf("%+v\n", err)
}

func exception(message string, binds ...interface{}) error {
	return errors.New(fmt.Sprintf(message, binds...))
}
