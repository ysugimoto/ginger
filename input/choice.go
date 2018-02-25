package input

import (
	"fmt"

	"github.com/ysugimoto/cho"

	"github.com/ysugimoto/ginger/internal/colors"
)

// Choice displays selection on supplied list.
func Choice(m string, exacts []string) string {
	fmt.Println(colors.Blue(prefix + m + ": "))
	ret := make(chan string, 1)
	terminate := make(chan struct{})
	go cho.Run(exacts, ret, terminate)
	selected := ""
LOOP:
	for {
		select {
		case selected = <-ret:
			break LOOP
		case <-terminate:
			break LOOP
		}
	}
	if selected != "" {
		fmt.Println(selected)
	}
	return selected
}
