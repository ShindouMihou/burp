package utils

import (
	"bytes"
	"unicode"
)

func HasPrefix(source, prefix []byte) bool {
	return len(source) >= len(prefix) && bytes.EqualFold(source[0:len(prefix)], prefix)
}

func Cut(source []byte, before byte) []byte {
	for index, char := range source {
		if char == before {
			return source[:index]
		}
	}
	return source
}

func Replace(source []byte, start int, end int, replacement []byte) []byte {
	var result []byte
	for index, char := range source {
		if index == start {
			result = append(result, replacement...)
		}
		if index >= start && index <= end {
			continue
		}
		result = append(result, char)
	}
	return result
}

func IsWhitespace(source []byte) bool {
	for _, char := range source {
		if !unicode.IsSpace(rune(char)) {
			return false
		}
	}
	return true
}

func IsNumeric(source []byte) bool {
	for _, char := range source {
		if !unicode.IsNumber(rune(char)) {
			return false
		}
	}
	return true
}
