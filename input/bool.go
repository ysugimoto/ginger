package input

import (
	"fmt"
	"strings"
)

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
