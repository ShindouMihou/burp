package burper

import (
	"burp/pkg/utils"
	"bytes"
	"errors"
	"strings"
)

type Function struct {
	Name        string
	Transformer Transformer
}

type Transformer = func(call *FunctionCall, flow *Flow) ([]byte, error)

var Functions = make(map[string]Function)

// Add adds a function to the function memory which can be called by using its name.
func Add(function Function) bool {
	Functions[strings.ToLower(function.Name)] = function
	return true
}

// Exec calls the function that was being called in the Burper statement.
func (call *FunctionCall) Exec(line []byte, flow *Flow) ([]byte, error) {
	function, exists := Functions[call.Function]
	if !exists {
		return nil, errors.New("cannot find any function named " + call.Function)
	}
	value, err := function.Transformer(call, flow)
	if err != nil {
		return nil, err
	}
	result := bytes.Replace(line, call.Source.FullMatch, value, 1)
	if call.As != nil {
		flow.Heap[*call.As] = value
	}
	return result, nil
}

// ExecStack assesses the arguments of the function to identify whether it is calling another function
// and executes the nested function if it is calling another function.
func (call *FunctionCall) ExecStack(flow *Flow) ([]string, error) {
	var stack []string
	for _, a := range call.Args {
		arg := []byte(a)
		if utils.HasPrefix(arg, CompletePrefixKey) {
			call, err := extractFunctionCalls(utils.Ptr(matchedFunctionCall{Start: 0, End: 0, FullMatch: arg, Match: arg}))
			if err != nil {
				return nil, call.FormatErr(err)
			}
			res, err := call.Exec(arg, flow)
			if err != nil {
				return nil, call.FormatErr(err)
			}
			arg = res
		}
		stack = append(stack, string(arg))
	}
	return stack, nil
}
