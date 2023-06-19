package burper

import (
	"bytes"
	"errors"
)

func (call *Call) Exec(line []byte, tree *Tree) ([]byte, error) {
	function, exists := Functions[call.Function]
	if !exists {
		return nil, errors.New("cannot find any function named " + call.Function)
	}
	value, err := function.Transformer(call, tree)
	if err != nil {
		return nil, err
	}
	result := bytes.Replace(line, call.Source.FullMatch, value, 1)
	if call.As != nil {
		tree.Store[*call.As] = value
	}
	return result, nil
}
