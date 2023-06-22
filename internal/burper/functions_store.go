package burper

import (
	"strings"
)

var Functions = make(map[string]Function)

func Add(function Function) bool {
	Functions[strings.ToLower(function.Name)] = function
	return true
}
