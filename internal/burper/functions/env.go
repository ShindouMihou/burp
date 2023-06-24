package functions

import (
	burper "burp/internal/burper"
	"os"
)

var _ = burper.Add(burper.Function{
	Name: "env",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		args, err := call.ExecStack(flow)
		if err != nil {
			return nil, call.FormatErr(err)
		}
		if len(args) == 0 {
			return nil, call.MissingArgumentErr("key", "string", "default", "string?")
		}
		key := args[0]
		e, ok := os.LookupEnv(key)
		if !ok {
			if len(args) == 2 {
				return []byte(args[1]), nil
			}
			return nil, call.Err("no environment variable named " + key)
		}
		return []byte(e), nil
	},
})
