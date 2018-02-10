package command

import (
	"fmt"

	"github.com/pkg/errors"
)

var debug string = ""

// debugTrace displays stacktrace.
// This fucntion will be enable only when debug build.
func debugTrace(err error) {
	if debug != "enable" {
		return
	}
	fmt.Println("StackTrace =====================")
	fmt.Printf("%+v\n", err)
}

// exception makes formatted string error interface.
func exception(message string, binds ...interface{}) error {
	return errors.New(fmt.Sprintf(message, binds...))
}
