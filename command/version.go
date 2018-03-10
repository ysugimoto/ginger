package command

import (
	"fmt"

	"github.com/ysugimoto/go-args"
)

// Version is the struct that displays version info.
type Version struct {
	Command
}

func NewVersion() *Version {
	return &Version{}
}

// Display build version.
//
// >>> doc
//
// ## Show version
//
// Show binary release version.
//
// ```
// $ ginger version
// ```
//
// <<< doc
func (v *Version) Run(ctx *args.Context) {
	fmt.Println(v.Help())
}

func (v *Version) Help() string {
	return version
}
