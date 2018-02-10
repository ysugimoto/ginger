package input

import (
	"fmt"
	"strings"
)

// Bool works confirmation on input accepts 'y' or 'n' and returns bool.
// If input is 'y', returns true, otherwise returns false.
func Bool(m string) bool {
	var input string
LOOP:
	for {
		fmt.Printf("%s [y/n]: ", m)
		fmt.Scanln(&input)
		switch strings.TrimRight(input, "\n") {
		case "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Println("Please type \"y\" or \"n\"")
			goto LOOP
		}
	}
}
