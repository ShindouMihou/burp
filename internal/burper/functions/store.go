package functions

import (
	burper "burp/internal/burper"
)

var _ = burper.Add(burper.Function{
	Name: "Store",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		args, err := call.ExecStack(flow)
		if err != nil {
			return nil, call.FormatErr(err)
		}
		if len(args) < 2 {
			return nil, call.MissingArgumentErr("value", "string")
		}
		key, value := args[0], args[1]
		if len(args) > 2 {
			for _, arg := range args[2:] {
				value += "," + arg
			}
		}
		flow.Heap[key] = []byte(value)
		return burper.WhitespaceKey, nil
	},
})
