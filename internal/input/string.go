package input

import (
	"fmt"
	"strings"
)

func String(m string) string {
	var input string
	fmt.Printf("%s: ", m)
	fmt.Scanln(&input)
	input = strings.TrimRight(input, "\n")
	if !strings.HasPrefix(input, "/") {
		input = "/" + input
	}
	return input
}
