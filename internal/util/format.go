package util

import (
	"strings"
)

func FormatPath(path string) string {
	return "/" + strings.Trim(path, "/")
}
