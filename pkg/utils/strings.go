package utils

import "strings"

func HasPrefixStr(source, prefix string) bool {
	return len(source) >= len(prefix) && strings.EqualFold(source[0:len(prefix)], prefix)
}

func HasSuffixStr(source, suffix string) bool {
	return len(source) >= len(suffix) && strings.EqualFold(source[len(source)-len(suffix):], suffix)
}
