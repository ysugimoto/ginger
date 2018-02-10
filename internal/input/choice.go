package input

import (
	"fmt"

	"github.com/ysugimoto/cho"
)

func Choice(m string, exacts []string) string {
	fmt.Println(m)
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
	return selected
}
