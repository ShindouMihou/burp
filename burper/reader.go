package burper

import (
	"burp/utils"
	"bytes"
)

func Parse(line []byte) ([]Call, error) {
	var calls []Call
	matches := findMatches(line)
	for _, match := range matches {
		match := match
		call, err := extractComponents(&match)
		if err != nil {
			// Likely that it's not a burp call.
			if err == MalformedBurpCall {
				continue
			}
			return nil, err
		}
		calls = append(calls, *call)
	}
	return calls, nil
}

type Call struct {
	Source   *Origin
	Function string
	Args     []string
	As       *string
}

type Origin struct {
	Start     int
	End       int
	Match     []byte
	FullMatch []byte
}

func findMatches(line []byte) []Origin {
	var matches []Origin
	start, end := -1, -1

	for index, char := range line {
		index, char := index, char
		if char == '[' {
			if start != -1 {
				continue
			}
			start = index
		} else if char == ']' && start != -1 {
			end = index
			matches = append(matches, Origin{Start: start, End: end, Match: line[start+1 : end], FullMatch: line[start : end+1]})

			start, end = -1, -1
		}
	}
	return matches
}

func extractComponents(origin *Origin) (*Call, error) {
	if !utils.HasPrefix(origin.Match, COMPLETE_PREFIX_KEY) {
		return nil, MalformedBurpCall
	}
	parts := bytes.SplitN(origin.Match, SEPERATOR_KEY, 2)
	if len(parts) != 2 {
		return nil, MalformedBurpCall
	}
	burp := Call{Source: origin}
	call := bytes.TrimSpace(parts[1:][0])
	var components [][]byte

	var args []byte
	argsOpenIndex := 0

	end := 0
	stack := 0
	for index, char := range call {
		index, char := index, char
		if char == '(' {
			if stack == 0 {
				argsOpenIndex = index + 1
			}
			stack++
			continue
		}
		if char == ')' {
			stack--
			if stack == 0 {
				args = call[argsOpenIndex:index]
			}
			continue
		}
		if char == ' ' && stack == 0 {
			components = append(components, call[end:index])
			end = index + 1
			continue
		}
		if (index + 1) == len(call) {
			components = append(components, call[end:(index+1)])
		}
	}
	if len(components) == 0 {
		components = [][]byte{call}
	}
	size := len(components) - 1
	if size > 1 && bytes.EqualFold(components[size-1], AS_TOKEN) {
		burp.As = utils.Ptr(string(components[size]))
	}
	burp.Function = string(bytes.ToLower(utils.Cut(components[0], '(')))
	burp.Args = extractArguments(args)
	return &burp, nil
}

func extractArguments(source []byte) []string {
	var args []string
	stack, start := 0, 0
	for index, char := range source {
		index, char := index, char
		if char == '(' {
			stack++
			continue
		}
		if char == ')' {
			stack--
			continue
		}
		if char == ',' {
			a := source[start:index]
			if !utils.IsWhitespace(a) {
				a = bytes.TrimSpace(source[start:index])
			}
			args = append(args, string(a))
			start = index + 1
			continue
		}
	}
	if start < len(source) {
		args = append(args, string(bytes.TrimSpace(source[start:])))
	}
	return args
}
