package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ysugimoto/ginger/internal/colors"
)

// String() works as string input and returns its value.
func String(m string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(colors.Blue(prefix+" %s: "), m)
	input, _ := reader.ReadString('\n')
	return strings.TrimRight(input, "\n")
}
