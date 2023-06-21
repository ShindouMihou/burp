package functions

import (
	"burp/burper"
	"burp/utils"
)

var _ = burper.Add(burper.Function{
	Name: "Store",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		args, err := call.ExecStack(tree)
		if err != nil {
			return nil, burper.CreateError(call, err.Error())
		}
		if len(args) < 2 {
			return nil, burper.CreateMissingArgumentError(call, utils.Array("value", "string"))
		}
		key, value := args[0], args[1]
		if len(args) > 2 {
			for _, arg := range args[2:] {
				value += "," + arg
			}
		}
		tree.Store[key] = []byte(value)
		return burper.WHITESPACE_KEY, nil
	},
})
