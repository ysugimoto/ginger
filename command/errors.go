package command

import (
	"fmt"
	"strings"

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
	return errors.New(strings.TrimRight(fmt.Sprintf(message, binds...), "\n") + "\n")
}
