package utils

import "strings"

func HasPrefixStr(source, prefix string) bool {
	return len(source) >= len(prefix) && strings.EqualFold(source[0:len(prefix)], prefix)
}
