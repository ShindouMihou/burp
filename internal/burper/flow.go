package burper

import (
	"bufio"
	"burp/pkg/fileutils"
	"bytes"
	"io"
)

type Flow struct {
	Source []byte
	Result [][]byte
	Heap   map[string][]byte
}

func (flow *Flow) Bytes() []byte {
	return bytes.Join(flow.Result, NewlineKey)
}

func (flow *Flow) String() string {
	return string(flow.Bytes())
}

func FromBytes(source []byte) (*Flow, error) {
	return New(bytes.NewReader(source))
}

func FromString(source string) (*Flow, error) {
	return FromBytes([]byte(source))
}

func FromFile(file string) (*Flow, error) {
	f, err := fileutils.Open(file)
	if err != nil {
		return nil, err
	}
	defer fileutils.Close(f)
	return New(f)
}

func New(reader io.Reader) (*Flow, error) {
	flow := &Flow{Heap: make(map[string][]byte)}
	buf := bufio.NewScanner(reader)
	for buf.Scan() {
		line := make([]byte, len(buf.Bytes()))
		copy(line, buf.Bytes())

		calls, err := parse(line)
		if err != nil {
			return nil, err
		}
		for _, call := range calls {
			line, err = call.Exec(line, flow)
			if err != nil {
				return nil, err
			}
		}
		flow.Result = append(flow.Result, line)
	}
	if buf.Err() != nil {
		return nil, buf.Err()
	}
	return flow, nil
}
