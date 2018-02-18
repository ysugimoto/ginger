package input

import (
	"fmt"
	"strings"
)

// String() works as string input and returns its value.
func String(m string) string {
	var input string
	fmt.Printf(color+prefix+" %s: "+reset, m)
	fmt.Scanln(&input)
	input = strings.TrimRight(input, "\n")
	if !strings.HasPrefix(input, "/") {
		input = "/" + input
	}
	return input
}
