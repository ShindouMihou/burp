package functions

import (
	burper "burp/internal/burper"
)

var _ = burper.Add(burper.Function{
	Name: "use",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, call.MissingArgumentErr("key", "string")
		}
		v, e := flow.Heap[call.Args[0]]
		if !e {
			return nil, call.Err("no such element called \"" + call.Args[0] + "\"")
		}
		return v, nil
	},
})
