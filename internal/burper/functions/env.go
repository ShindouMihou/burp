package functions

import (
	burper "burp/internal/burper"
	"burp/pkg/utils"
	"os"
)

var _ = burper.Add(burper.Function{
	Name: "env",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		args, err := call.ExecStack(tree)
		if err != nil {
			return nil, burper.CreateError(call, err.Error())
		}
		if len(args) == 0 {
			return nil, burper.CreateMissingArgumentError(call, utils.Array("key", "string", "default", "string?"))
		}
		key := args[0]
		e, ok := os.LookupEnv(key)
		if !ok {
			if len(args) == 2 {
				return []byte(args[1]), nil
			}
			return nil, burper.CreateError(call, "no environment variable named "+key)
		}
		return []byte(e), nil
	},
})
