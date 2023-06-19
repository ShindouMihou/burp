package burper

import (
	"bufio"
	"bytes"
	"io"
)

type Tree struct {
	Source []byte
	Result [][]byte
	Store  map[string][]byte
}

func (tree *Tree) Bytes() []byte {
	return bytes.Join(tree.Result, NEWLINE_KEY)
}

func (tree *Tree) String() string {
	return string(tree.Bytes())
}

func FromBytes(source []byte) (*Tree, error) {
	return New(bytes.NewReader(source))
}

func FromString(source string) (*Tree, error) {
	return New(bytes.NewReader([]byte(source)))
}

func New(reader io.Reader) (*Tree, error) {
	tree := &Tree{Store: make(map[string][]byte)}
	buf := bufio.NewScanner(reader)
	for buf.Scan() {
		line := buf.Bytes()
		calls, err := Parse(line)
		if err != nil {
			return nil, err
		}
		for _, call := range calls {
			line, err = call.Exec(line, tree)
			if err != nil {
				return nil, err
			}
		}
		tree.Result = append(tree.Result, line)
	}
	return tree, nil
}
