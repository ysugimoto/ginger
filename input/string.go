package input

import (
	"fmt"
	"strings"

	"github.com/ysugimoto/ginger/internal/colors"
)

// String() works as string input and returns its value.
func String(m string) string {
	var input string
	fmt.Printf(colors.Blue(prefix+" %s: "), m)
	fmt.Scanln(&input)
	return strings.TrimRight(input, "\n")
}
