package burper

import (
	"burp/pkg/utils"
	"bytes"
)

func parse(line []byte) ([]FunctionCall, error) {
	var calls []FunctionCall
	matches := findFunctionCalls(line)
	for _, match := range matches {
		match := match
		call, err := extractFunctionCalls(&match)
		if err != nil {
			// Likely that it's not a burp call.
			if err == ErrMalformedBurpCall {
				continue
			}
			return nil, err
		}
		calls = append(calls, *call)
	}
	return calls, nil
}

type FunctionCall struct {
	Source     *matchedFunctionCall
	Identifier string
	Args       []string
	As         *string
}

type matchedFunctionCall struct {
	Start     int
	End       int
	Match     []byte
	FullMatch []byte
}

func findFunctionCalls(line []byte) []matchedFunctionCall {
	var matches []matchedFunctionCall
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
			matches = append(matches, matchedFunctionCall{Start: start, End: end, Match: line[start+1 : end], FullMatch: line[start : end+1]})

			start, end = -1, -1
		}
	}
	return matches
}

func extractFunctionCalls(match *matchedFunctionCall) (*FunctionCall, error) {
	if !utils.HasPrefix(match.Match, CompletePrefixKey) {
		return nil, ErrMalformedBurpCall
	}
	// Splits "burp:" and "Identifier(args)" apart which leaves us with two parts.
	parts := bytes.SplitN(match.Match, SeperatorKey, 2)
	if len(parts) != 2 {
		return nil, ErrMalformedBurpCall
	}
	functionCall := FunctionCall{Source: match}

	callText := bytes.TrimSpace(parts[1])
	var components [][]byte

	var args []byte
	argsOpenIndex := 0

	stack, end := 0, 0
	hasParenthesis := false
	for index, char := range callText {
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
				hasParenthesis = true
				args = callText[argsOpenIndex:index]
			}
			continue
		}
		if char == ' ' && stack == 0 {
			components = append(components, callText[end:index])
			end = index + 1
			continue
		}
		if (index + 1) == len(callText) {
			components = append(components, callText[end:(index+1)])
		}
	}
	if len(components) == 0 {
		components = [][]byte{callText}
	}
	size := len(components) - 1
	if size > 1 && bytes.EqualFold(components[size-1], AsToken) {
		functionCall.As = utils.Ptr(string(components[size]))
	}
	functionCall.Identifier = string(bytes.ToLower(utils.Cut(components[0], '(')))
	if hasParenthesis {
		functionCall.Args = extractFunctionArguments(args)
	} else {
		functionCall.Args = []string{functionCall.Identifier}
		functionCall.Identifier = "use"
	}
	return &functionCall, nil
}

func extractFunctionArguments(source []byte) []string {
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
