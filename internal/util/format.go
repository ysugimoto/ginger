package util

import (
	"strings"

	"github.com/ysugimoto/ginger/input"
)

func FormatPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

func ChooseMethod(def string) string {
	methods := []string{
		"ANY",
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"PATCH",
		"OPTIONS",
		"HEAD",
	}
	selected := input.Choice("Choose method", methods)
	if selected == "" {
		return def
	}
	return selected
}
